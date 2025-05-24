package main

import (
	"log"
	"net/http"
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

	_, raw, err := repo.InitDB(&dbOpened, envDbPath)
	if err != nil {
		log.Println(err)
		return ERROR_EXIT
	}
	defer func() {
		raw.Close()
		log.Println("bbolt DB closed")
	}()

	smsSender, smsAddressBook := createSender()
	smsHandler := NewSmsHandler(createLogger(), smsSender, smsAddressBook)
	http.HandleFunc("POST /notifier/api/send/{recipient}", smsHandler.Handle)

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
