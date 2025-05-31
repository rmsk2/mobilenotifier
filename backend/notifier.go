package main

import (
	"log"
	"net/http"
	"notifier/controller"
	"notifier/repo"
	"notifier/sms"
	"notifier/warner"
	"os"
	"time"

	_ "notifier/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

const envApiKey = "IFTTT_API_KEY"
const envDbPath string = "DB_PATH"
const envServeLocal string = "LOCALDIR"
const envSwaggerUrl = "SWAGGER_URL"
const ERROR_EXIT = 42
const ERROR_OK = 0

func createLogger() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

func createSender() (sms.SmsSender, sms.SmsAddressBook) {
	apiKey, ok := os.LookupEnv(envApiKey)
	if !ok {
		dummy := sms.NewDummySender()
		return dummy, dummy
	} else {
		ifft := sms.NewIftttSender(apiKey)
		return ifft, ifft
	}
}

func run() int {
	dbOpened := false

	swaggerUrl, ok := os.LookupEnv(envSwaggerUrl)
	if !ok {
		swaggerUrl = "http://localhost:5100/notifier/api/swagger/doc.json"
	}

	boltPath, ok := os.LookupEnv(envDbPath)
	if !ok {
		log.Printf("environment variable '%s' not found in environment", envDbPath)
		return ERROR_EXIT
	}

	dbl, rawDB, err := repo.InitDB(&dbOpened, boltPath)
	if err != nil {
		log.Println(err)
		return ERROR_EXIT
	}
	defer func() {
		rawDB.Close()
		log.Println("bbolt DB closed")
	}()

	smsSender, smsAddressBook := createSender()
	smsController := controller.NewSmsController(createLogger(), smsSender, smsAddressBook)
	smsController.Add()

	notificationController := controller.NewNotificationController(dbl, smsAddressBook, createLogger())
	notificationController.Add()

	t := time.NewTicker(60 * time.Second)
	warner.Start(dbl, smsSender, smsAddressBook, t, createLogger())

	dirName, ok := os.LookupEnv(envServeLocal)
	if ok {
		log.Println("Serving webapp locally")
		http.Handle("/notifier/app/", http.StripPrefix("/notifier/app/", http.FileServer(http.Dir(dirName))))
	}

	http.HandleFunc("/notifier/api/swagger/", httpSwagger.Handler(httpSwagger.URL(swaggerUrl)))

	err = http.ListenAndServe(":5100", nil)
	if err != nil {
		log.Println(err)
		return ERROR_EXIT
	}

	return ERROR_OK
}

func main() {
	os.Exit(run())
}
