package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/egorkovalchuk/go-clearrabbit/data"
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
)

//чтение конфига
func readconf(cfg *data.Config, confname string) {
	file, err := os.Open(confname)
	if err != nil {
		ProcessError(err)
		fmt.Println(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		ProcessError(err)
		fmt.Println(err)
	}

	file.Close()
}

//Запись ошибки с прекращением выполнения
func ProcessError(err error) {
	log.Println(err)
	os.Exit(2)
}

//Запись ошибки
func Error(err error) {
	log.Println(err)
	os.Exit(2)
}

//Запись в лог при включенном дебаге
func ProcessDebug(logtext interface{}) {
	if debugm {
		log.Println(logtext)
	}
}

func main() {

	//start program
	filer, err = os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(filer)
	log.Println("- - - - - - - - - - - - - - -")
	log.Println("Start Clear All queues ")

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
		log.Println("Please set your login, password")
		fmt.Println("Please set your login, password")
		return
	}

	ProcessDebug("Start with debug mode")

	StartClear()

}

func StartClear() {
	for _, i := range cfg.ServerList {
		request := "http://" + i + "/api/queues" //+ url.QueryEscape()
		log.Println("Request list queue to " + i)

		ProcessDebug("Request " + request)

		resp, err := http.NewRequest("GET", request, nil)
		resp.SetBasicAuth(LoginRabbit, PassRabbit)

		if err != nil {

			log.Println(err)
		}

		cli := &http.Client{}
		rsp, err := cli.Do(resp)

		if err != nil {
			log.Println(err)
		}

		if rsp.StatusCode >= 200 && rsp.StatusCode <= 299 {
			log.Println("HTTP Status is in the 2xx range")
		} else {
			log.Println("HTTP Status error " + strconv.Itoa(rsp.StatusCode))
		}

		// проверяем получение картинки, статус 200
		if rsp.StatusCode == 200 {
			queuejson, err := data.JsonQueueParse(rsp)
			if err != nil {
				log.Println(err)
			}

			for _, j := range queuejson {
				requestpurge := "http://" + i + "/api/queues/" + url.QueryEscape(j.Vhost) + "/" + url.QueryEscape(j.Name) + "/contents"
				log.Println("Purge " + j.Vhost + "  " + j.Name)
				ProcessDebug(requestpurge)

				resppurge, err := http.NewRequest("DELETE", requestpurge, nil)
				resppurge.SetBasicAuth(LoginRabbit, PassRabbit)

				if err != nil {

					log.Println(err)
				}

				clipurge := &http.Client{}
				rsppurge, err := clipurge.Do(resppurge)

				if err != nil {
					log.Println(err)
				}

				if rsppurge.StatusCode >= 200 && rsppurge.StatusCode <= 299 {
					log.Println("HTTP Status is in the 2xx range")
				} else {
					log.Println("HTTP Status error " + strconv.Itoa(rsppurge.StatusCode))
				}

				ProcessDebug(rsppurge.Body)

				log.Println("Purge Ok")

			}

		}

		defer rsp.Body.Close()

	}
}
