package main

import (
	"fmt"
	"log"
	"net/http"
	"notifier/repo"
	"notifier/sms"
	"os"
	"os/signal"
	"syscall"

	bolt "go.etcd.io/bbolt"
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

func InstallSignalHandler(db *bolt.DB, dbOpened *bool) {
	go func(openFlag *bool) {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		if *dbOpened {
			log.Println("Closing BBolt DB")
			db.Close()
		}
		os.Exit(0)
	}(dbOpened)
}

func initDB(openFlag *bool) (*bolt.DB, error) {
	boltPath, ok := os.LookupEnv(envDbPath)
	if !ok {
		return nil, fmt.Errorf("environment variable '%s' not found in environment", envDbPath)
	}

	db, err := bolt.Open(boltPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to open database file %s: %v", boltPath, err)
	}

	*openFlag = true
	InstallSignalHandler(db, openFlag)

	err = repo.CreateBuckets(db)
	if err != nil {
		db.Close()
		*openFlag = false
		return nil, fmt.Errorf("unable to create buckets in database file %s: %v", boltPath, err)
	}

	return db, nil
}

func run() int {
	dbOpened := false

	db, err := initDB(&dbOpened)
	if err != nil {
		log.Println(err)
		return ERROR_EXIT
	}
	defer func() {
		db.Close()
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
