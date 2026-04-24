# Makefile для проекта утилиты смены паролей

# Переменные
BINARY_NAME=clearrabbit
BUILD_DIR=build
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags=-extldflags=-static"-s -w -X main.versionutil=$(VERSION)"
CMD_DIR=./cmd

VERSION=$(shell $(GO) run  $(CMD_DIR) -v | awk '{print $$NF}' || echo "v0.0.0")

# Цели по умолчанию
.PHONY: all build clean test run help

all: clean build

# Сборка проекта
build:
	@echo "Сборка $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Сборка для Linux
build-linux:
	@echo "Сборка для Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)

# Сборка для Windows
build-windows:
	@echo "Сборка для Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(CMD_DIR)

# Очистка
clean:
	@echo "Очистка..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f *.log

# Запуск тестов
test:
	@echo "Запуск тестов..."
	$(GO) test ./... -v

# Запуск приложения
run:
	@echo "Запуск приложения..."
	$(GO) run $(GOFLAGS) ./cmd

# Форматирование кода
fmt:
	@echo "Форматирование кода..."
	$(GO) fmt ./...

# Проверка зависимостей
tidy:
	@echo "Обновление зависимостей..."
	$(GO) mod tidy

# Проверка линтером
lint:
	@echo "Проверка линтером..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint не установлен. Установите: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Создание релизных версий
release: clean build-linux build-windows
	@echo "Создание релизных версий..."
	@cd $(BUILD_DIR) && \
		tar -czf $(BINARY_NAME)-linux-$(VERSION).tar.gz $(BINARY_NAME)-linux && \
		zip $(BINARY_NAME)-windows-$(VERSION).zip $(BINARY_NAME).exe && \
		rm -f $(BINARY_NAME).exe $(BINARY_NAME)-linux

# Показать справку
help:
	@echo "Доступные команды:"
	@echo "  make all        - Очистка и сборка"
	@echo "  make build      - Сборка проекта"
	@echo "  make build-linux - Сборка для Linux"
	@echo "  make build-windows - Сборка для Windows"
	@echo "  make clean      - Очистка сборочных файлов"
	@echo "  make test       - Запуск тестов"
	@echo "  make run        - Запуск приложения"
	@echo "  make fmt        - Форматирование кода"
	@echo "  make tidy       - Обновление зависимостей"
	@echo "  make lint       - Проверка линтером"
	@echo "  make release    - Создание релизных версий"
	@echo "  make help       - Показать эту справку"

# Установка зависимостей для разработки
dev-deps:
	@echo "Установка инструментов разработки..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest

# Проверка версии Go
check-go:
	@echo "Версия Go:"
	@$(GO) version
	@echo "Версия модулей:"
	@$(GO) mod graph | head -5