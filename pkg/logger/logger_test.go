package logger

import (
	"bytes"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

type testStringer struct {
	s string
}

func (ts testStringer) String() string {
	return ts.s
}

// TestNewLogWriter проверяет создание логгера.
func TestNewLogWriter(t *testing.T) {
	filename := "test.log"
	debug := true
	lw := NewLogWriter(filename, debug)

	if lw.logFileName != filename {
		t.Errorf("ожидали имя файла %q, получили %q", filename, lw.logFileName)
	}
	if lw.debugm != debug {
		t.Errorf("ожидали debugm = %v, получили %v", debug, lw.debugm)
	}
	if lw.LogChannel == nil {
		t.Error("LogChannel не должен быть nil")
	}
	if lw.masked != true {
		t.Errorf("ожидали masked = true по умолчанию, получили %v", lw.masked)
	}
}

// TestPrefix проверяет формирование префикса.
func TestPrefix(t *testing.T) {
	lw := NewLogWriter("", false)

	tests := []struct {
		name   string
		opt    []string
		expect string
	}{
		{"без опций", []string{}, "[MAIN] "},
		{"одна опция", []string{"MODULE"}, "[MODULE] "},
		{"две опции", []string{"MODULE", "FUNC"}, "[MODULE] (FUNC) "},
		{"три опции", []string{"MODULE", "FUNC", "EXTRA"}, "[MODULE] (FUNC) EXTRA "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lw.Prefix(tt.opt...)
			if result != tt.expect {
				t.Errorf("ожидали %q, получили %q", tt.expect, result)
			}
		})
	}
}

// TestProcessDebug проверяет отправку DEBUG-сообщения в канал (без горутины).
func TestProcessDebug(t *testing.T) {
	lw := NewLogWriter("", true)
	// делаем канал буферизированным, чтобы не блокироваться
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessDebug("test debug", "MODULE")

	select {
	case entry := <-lw.LogChannel:
		expectedPrefix := "[MODULE] DEBUG"
		if entry.t != expectedPrefix {
			t.Errorf("ожидали префикс %q, получили %q", expectedPrefix, entry.t)
		}
		if entry.text != "test debug" {
			t.Errorf("ожидали текст %q, получили %q", "test debug", entry.text)
		}
	default:
		t.Error("канал пуст, сообщение не отправлено")
	}
}

// TestProcessDebugDisabled проверяет, что при debugm=false DEBUG не отправляется.
func TestProcessDebugDisabled(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessDebug("should not appear", "MODULE")

	select {
	case <-lw.LogChannel:
		t.Error("DEBUG сообщение отправлено, хотя debugm=false")
	default:
		// ok
	}
}

// TestProcessError проверяет отправку ERROR.
func TestProcessError(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessError("error text", "MODULE")

	select {
	case entry := <-lw.LogChannel:
		expectedPrefix := "[MODULE] ERROR"
		if entry.t != expectedPrefix {
			t.Errorf("ожидали префикс %q, получили %q", expectedPrefix, entry.t)
		}
		if entry.text != "error text" {
			t.Errorf("ожидали текст %q, получили %q", "error text", entry.text)
		}
	default:
		t.Error("канал пуст")
	}
}

// TestProcessErrorAny проверяет отправку с несколькими аргументами.
func TestProcessErrorAny(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessErrorAny("error", 123, true)

	select {
	case entry := <-lw.LogChannel:
		if entry.t != "ERROR" {
			t.Errorf("ожидали префикс ERROR, получили %q", entry.t)
		}
		expectedText := "error 123 true "
		if entry.text != expectedText {
			t.Errorf("ожидали текст %q, получили %q", expectedText, entry.text)
		}
	default:
		t.Error("канал пуст")
	}
}

// TestProcessCritical проверяет отправку CRITICAL (вывод в stdout не проверяем).
func TestProcessCritical(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessCritical("critical text", "MODULE")

	select {
	case entry := <-lw.LogChannel:
		expectedPrefix := "[MODULE] CRITICAL"
		if entry.t != expectedPrefix {
			t.Errorf("ожидали префикс %q, получили %q", expectedPrefix, entry.t)
		}
		if entry.text != "critical text" {
			t.Errorf("ожидали текст %q, получили %q", "critical text", entry.text)
		}
	default:
		t.Error("канал пуст")
	}
}

// TestProcessInfo проверяет INFO.
func TestProcessInfo(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessInfo("info text", "MODULE")

	select {
	case entry := <-lw.LogChannel:
		expectedPrefix := "[MODULE] INFO"
		if entry.t != expectedPrefix {
			t.Errorf("ожидали префикс %q, получили %q", expectedPrefix, entry.t)
		}
		if entry.text != "info text" {
			t.Errorf("ожидали текст %q, получили %q", "info text", entry.text)
		}
	default:
		t.Error("канал пуст")
	}
}

// TestProcessWarm проверяет WARM (опечатка, но оставим как есть).
func TestProcessWarm(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessWarm("warm text", "MODULE")

	select {
	case entry := <-lw.LogChannel:
		expectedPrefix := "[MODULE] WARM"
		if entry.t != expectedPrefix {
			t.Errorf("ожидали префикс %q, получили %q", expectedPrefix, entry.t)
		}
		if entry.text != "warm text" {
			t.Errorf("ожидали текст %q, получили %q", "warm text", entry.text)
		}
	default:
		t.Error("канал пуст")
	}
}

// TestProcessLog проверяет общий метод ProcessLog.
func TestProcessLog(t *testing.T) {
	lw := NewLogWriter("", true)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.ProcessLog("CUSTOM", "custom text", "MODULE")

	select {
	case entry := <-lw.LogChannel:
		expectedPrefix := "[MODULE] CUSTOM"
		if entry.t != expectedPrefix {
			t.Errorf("ожидали префикс %q, получили %q", expectedPrefix, entry.t)
		}
		if entry.text != "custom text" {
			t.Errorf("ожидали текст %q, получили %q", "custom text", entry.text)
		}
	default:
		t.Error("канал пуст")
	}
}

// TestPrintf проверяет форматированный вывод.
func TestPrintf(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	lw.Printf("INFO", "число %d, строка %s", 42, "test")

	select {
	case entry := <-lw.LogChannel:
		if entry.t != "[MAIN] INFO" {
			t.Errorf("ожидали уровень INFO, получили %q", entry.t)
		}
		expectedText := "число 42, строка test"
		if entry.text != expectedText {
			t.Errorf("ожидали текст %q, получили %q", expectedText, entry.text)
		}
	default:
		t.Error("канал пуст")
	}
}

// TestChangeDebugLevel проверяет смену уровня отладки.
func TestChangeDebugLevel(t *testing.T) {
	lw := NewLogWriter("", false)
	lw.LogChannel = make(chan LogStruct, 10)

	// проверяем начальное значение
	if lw.GetDebugLevel() != false {
		t.Error("начальный debugm должен быть false")
	}

	// включаем
	lw.ChangeDebugLevel(true)
	if lw.GetDebugLevel() != true {
		t.Error("после включения debugm должен быть true")
	}
	// проверяем, что в канал отправилось сообщение о включении
	select {
	case entry := <-lw.LogChannel:
		if entry.t != "[MAIN] DEBUG" {
			t.Errorf("ожидали префикс '[MAIN] DEBUG', получили %q", entry.t)
		}
		if entry.text != "Start debug mode" {
			t.Errorf("ожидали текст 'Start debug mode', получили %q", entry.text)
		}
	default:
		t.Error("сообщение о включении debug не отправлено")
	}

	// выключаем
	lw.ChangeDebugLevel(false)
	if lw.GetDebugLevel() != false {
		t.Error("после выключения debugm должен быть false")
	}
	select {
	case entry := <-lw.LogChannel:
		if entry.t != "[MAIN] DEBUG" {
			t.Errorf("ожидали префикс '[MAIN] DEBUG', получили %q", entry.t)
		}
		if entry.text != "Stop debug mode" {
			t.Errorf("ожидали текст 'Stop debug mode', получили %q", entry.text)
		}
	default:
		t.Error("сообщение о выключении debug не отправлено")
	}
}

// TestChangeMasked проверяет смену флага маскирования.
func TestChangeMasked(t *testing.T) {
	lw := NewLogWriter("", false)
	if lw.masked != true {
		t.Error("по умолчанию masked должен быть true")
	}
	lw.ChangeMasked(false)
	if lw.masked != false {
		t.Error("после вызова ChangeMasked(false) masked должен стать false")
	}
	lw.ChangeMasked(true)
	if lw.masked != true {
		t.Error("после вызова ChangeMasked(true) masked должен стать true")
	}
}

// TestHidePassword проверяет функцию hidePassword.
func TestHidePassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "New password",
			input:    "New password: secret123",
			expected: "New password: ****",
		},
		{
			name:     "URL credentials",
			input:    "http://user:pass@example.com",
			expected: "http://user:****@example.com",
		},
		{
			name:     "Credential: path",
			input:    "Credential: some/path",
			expected: "Credential: some/****",
		},
		{
			name:     "SecretKey: path",
			input:    "SecretKey: etc/secret.key",
			expected: "SecretKey: etc/****",
		},
		{
			name:     "encrypt_password JSON",
			input:    `{"encrypt_password": "secret"}`,
			expected: `{"encrypt_password": "****"}`,
		},
		{
			name:     "newPassword JSON",
			input:    `{"newPassword": "mypass"}`,
			expected: `{"newPassword": "****"}`,
		},
		{
			name:     "newPasswordRepeat JSON",
			input:    `{"newPasswordRepeat": "mypass"}`,
			expected: `{"newPasswordRepeat": "****"}`,
		},
		{
			name:     "currentPassword JSON",
			input:    `{"currentPassword": "oldpass"}`,
			expected: `{"currentPassword": "****"}`,
		},
		{
			name:     "IDENTIFIED BY",
			input:    `IDENTIFIED BY "secret"`,
			expected: `IDENTIFIED BY "****"`,
		},
		{
			name:     "REPLACE",
			input:    `REPLACE "secret"`,
			expected: `REPLACE "****"`,
		},
		{
			name:     "no password",
			input:    "just regular text",
			expected: "just regular text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hidePassword(tt.input)
			if result != tt.expected {
				t.Errorf("ожидали %q, получили %q", tt.expected, result)
			}
		})
	}
}

// TestSafeLogText проверяет safeLogText с разными типами.
func TestSafeLogText(t *testing.T) {
	// string
	strRes := safeLogText("New password: 123")
	if strRes != "New password: ****" {
		t.Errorf("safeLogText не замаскировал строку: %q", strRes)
	}

	// Используем объявленный выше тип
	ts := testStringer{"SecretKey: some/key"}
	strRes2 := safeLogText(ts)
	if strRes2 != "SecretKey: some/****" {
		t.Errorf("safeLogText не замаскировал Stringer: %q", strRes2)
	}

	// другой тип (int)
	intRes := safeLogText(12345)
	if intRes != "12345" {
		t.Errorf("safeLogText для int вернул %q, ожидали '12345'", intRes)
	}
}

// TestFullLoggingIntegration проверяет полный цикл: отправка сообщений, обработка канала, запись в буфер с маскированием.
func TestFullLoggingIntegration(t *testing.T) {
	// Создаём временный файл (он будет создан, но мы подменим logger)
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Создаём логгер с файлом
	lw := NewLogWriter(tmpFile.Name(), true)
	// Канал буферизированный, чтобы не блокироваться при отправке до запуска горутины
	lw.LogChannel = make(chan LogStruct, 10)

	// Буфер для перехвата вывода лога
	var buf bytes.Buffer
	testLogger := log.New(&buf, "", 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Запускаем обработчик канала
		lw.LogWriteForGoRutineStruct()
	}()

	// Даём горутине время запуститься и открыть файл (но мы сразу подменим logger)
	time.Sleep(10 * time.Millisecond)
	// Подменяем logger на наш, который пишет в буфер
	lw.logger = testLogger

	// Отправляем несколько сообщений
	lw.ProcessError("error message", "TEST")
	lw.ProcessDebug("debug message", "TEST")
	lw.ProcessInfo("info with password: New password: 123", "TEST")
	lw.ProcessWarm("warm with URL http://user:secret@host", "TEST")

	// Даём время обработаться (можно через закрытие канала)
	close(lw.LogChannel)
	wg.Wait()

	// Проверяем содержимое буфера
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Fatalf("ожидали 4 строки лога, получили %d: %q", len(lines), output)
	}

	// Регулярка для временной метки
	timePattern := `^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `
	for i, line := range lines {
		matched, _ := regexp.MatchString(timePattern, line)
		if !matched {
			t.Errorf("строка %d не начинается с временной метки: %q", i, line)
		}
		// Проверяем наличие уровня и текста
		if !strings.Contains(line, "TEST] ERROR: error message") &&
			!strings.Contains(line, "TEST] DEBUG: debug message") &&
			!strings.Contains(line, "TEST] INFO: info with password: New password: ****") &&
			!strings.Contains(line, "TEST] WARM: warm with URL http://user:****@host") {
			t.Errorf("неожиданное содержимое строки %d: %q", i, line)
		}
	}
}

// TestMaskingDisabled проверяет, что при masked=false пароли не скрываются.
func TestMaskingDisabled(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	lw := NewLogWriter(tmpFile.Name(), true)
	lw.LogChannel = make(chan LogStruct, 10)

	var buf bytes.Buffer
	testLogger := log.New(&buf, "", 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		lw.LogWriteForGoRutineStruct()
	}()
	time.Sleep(10 * time.Millisecond)
	lw.logger = testLogger

	// Отключаем маскирование
	lw.ChangeMasked(false)

	lw.ProcessError("New password: secret", "TEST")

	close(lw.LogChannel)
	wg.Wait()

	output := buf.String()
	if !strings.Contains(output, "New password: secret") {
		t.Errorf("маскирование должно быть отключено, но строка содержит маскировку: %q", output)
	}
	if strings.Contains(output, "****") {
		t.Errorf("появились звездочки при отключённом маскировании: %q", output)
	}
}
