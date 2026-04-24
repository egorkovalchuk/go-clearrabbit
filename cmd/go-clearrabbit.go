package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/egorkovalchuk/go-clearrabbit/pkg/data"
	"github.com/egorkovalchuk/go-clearrabbit/pkg/logger"
)

//Power by  Egor Kovalchuk

const (
	// логи
	logFileName  = "report.log"
	confFileName = "config.json"
	versionutil  = "0.0.1"
)

var (
	//Configuration
	cfg data.Config
	//FSM connect
	LoginRabbit string
	PassRabbit  string

	//режим работы сервиса(дебаг мод)
	debugm bool
	//Запись в лог
	filer *os.File
	//запрос помощи
	help bool
	//ошибки
	err error
	//запрос версии
	version bool

	logs *logger.LogWriter
)

// чтение конфига
func readconf(cfg *data.Config, confname string) {
	file, err := os.Open(confname)
	if err != nil {
		logs.ProcessError(err)
		fmt.Println(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		logs.ProcessError(err)
		fmt.Println(err)
	}

	file.Close()
}

func init() {
	logs = logger.NewLogWriter(logFileName, debugm)
	go logs.LogWriteForGoRutineStruct()
}

func main() {

	//start program
	filer, err = os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logs.ProcessPanic(err)
	}

	logs.ProcessInfo("- - - - - - - - - - - - - - -")
	logs.ProcessInfo("Start Clear All queues ")

	flag.BoolVar(&debugm, "d", false, "Start debug mode")
	flag.BoolVar(&version, "v", false, "Version")
	var confname string
	flag.StringVar(&confname, "c", confFileName, "start with users config")
	flag.StringVar(&LoginRabbit, "l", "", "RabbitMQ Login")
	flag.StringVar(&PassRabbit, "p", "", "RabbitMQ Password")
	flag.BoolVar(&help, "h", false, "Use -h for help")
	flag.Parse()

	readconf(&cfg, confname)

	//Получение помощи
	if help {
		data.Helpstart()
		return
	}

	//получение версии
	if version {
		fmt.Println("Version utils " + versionutil)
		return
	}

	if LoginRabbit == "" || PassRabbit == "" {
		logs.ProcessInfo("Please set your login, password")
		fmt.Println("Please set your login, password")
		return
	}

	logs.ProcessDebug("Start with debug mode")
	StartClear()
	logs.ProcessInfo("Finish")
	<-time.After(1 * time.Second)
}

func StartClear() {
	for _, i := range cfg.ServerList {
		request := "http://" + i + "/api/queues" //+ url.QueryEscape()
		logs.ProcessInfo("Request list queue to " + i)

		logs.ProcessDebug("Request "+request, i)

		resp, err := http.NewRequest("GET", request, nil)
		if err != nil {
			logs.ProcessError(err, i)
			continue
		}

		resp.SetBasicAuth(LoginRabbit, PassRabbit)
		cli := &http.Client{}
		rsp, err := cli.Do(resp)

		if err != nil {
			logs.ProcessError(err, i)
			continue
		} else {
			if rsp.StatusCode >= 200 && rsp.StatusCode <= 299 {
				logs.ProcessInfo("HTTP Status is in the 2xx range", i)
			} else {
				logs.ProcessError("HTTP Status error "+strconv.Itoa(rsp.StatusCode), i)
				continue
			}

			// проверяем получение картинки, статус 200
			if rsp.StatusCode == 200 {
				queuejson, err := data.JsonQueueParse(rsp)
				if err != nil {
					logs.ProcessError(err, i)
				}

				for _, j := range queuejson {
					requestpurge := "http://" + i + "/api/queues/" + url.QueryEscape(j.Vhost) + "/" + url.QueryEscape(j.Name) + "/contents"
					logs.ProcessInfo("Purge "+j.Vhost+"  "+j.Name, i)
					logs.ProcessDebug(requestpurge, i, j.Name)

					resppurge, err := http.NewRequest("DELETE", requestpurge, nil)
					resppurge.SetBasicAuth(LoginRabbit, PassRabbit)

					if err != nil {

						logs.ProcessError(err, i, j.Name)
					}

					clipurge := &http.Client{}
					rsppurge, err := clipurge.Do(resppurge)

					if err != nil {
						logs.ProcessError(err, i, j.Name)
					}

					if rsppurge.StatusCode >= 200 && rsppurge.StatusCode <= 299 {
						logs.ProcessInfo("HTTP Status is in the 2xx range", i, j.Name)
					} else {
						logs.ProcessError("HTTP Status error "+strconv.Itoa(rsppurge.StatusCode), i, j.Name)
					}

					logs.ProcessDebug(rsppurge.Body, i, j.Name)

					logs.ProcessInfo("Purge Ok", i, j.Name)

				}
				defer rsp.Body.Close()
			}
		}
	}
}
