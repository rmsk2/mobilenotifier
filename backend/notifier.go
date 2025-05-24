package main

import (
	"log"
	"net/http"
	"notifier/handler"
	"notifier/repo"
	"notifier/sms"
	"os"
)

const envApiKey = "IFTTT_API_KEY"
const envDbPath string = "DB_PATH"
const envServeLocal string = "LOCALDIR"
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
	smsHandler := handler.NewSmsHandler(createLogger(), smsSender, smsAddressBook)
	http.HandleFunc("POST /notifier/api/send/{recipient}", smsHandler.Handle)

	notificationHandler := handler.NewNotifationHandler(dbl, smsAddressBook, createLogger())
	http.HandleFunc("POST /notifier/api/notification", notificationHandler.HandlePost)
	http.HandleFunc("/notifier/api/notification", notificationHandler.HandleList)
	http.HandleFunc("DELETE /notifier/api/notification/delete/{uuid}", notificationHandler.HandleDelete)
	http.HandleFunc("/notifier/api/notification/expiry/{uuid}", notificationHandler.HandleExpiry)

	dirName, ok := os.LookupEnv(envServeLocal)
	if ok {
		log.Println("Serving webapp locally")
		http.Handle("/notifier/app/", http.StripPrefix("/notifier/app/", http.FileServer(http.Dir(dirName))))
	}

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
