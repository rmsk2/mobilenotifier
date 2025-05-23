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

func createLogger() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

func createSender() sms.SmsSender {
	apiKey, ok := os.LookupEnv(envApiKey)
	if !ok {
		return sms.NewDummySender()
	} else {
		return sms.NewIftttSender(apiKey)
	}
}

func InstallSignalHandler(db *bolt.DB, dbOpened *bool) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		if *dbOpened {
			log.Println("Closing BBolt DB")
			db.Close()
		}
		os.Exit(0)
	}()
}

func main() {
	dbOpened := false

	boltPath, ok := os.LookupEnv(envDbPath)
	if !ok {
		log.Fatalf("Environment variable '%s' not found in environment", envDbPath)
	}

	db, err := bolt.Open(boltPath, 0600, nil)
	if err != nil {
		log.Fatalf("Unable to open database file %s: %v\n", boltPath, err)
	}
	dbOpened = true
	defer func() {
		db.Close()
		log.Println("bbolt DB closed")
	}()

	err = repo.CreateBuckets(db)
	if err != nil {
		log.Fatalf("Unable to create buckets in database file %s: %v\n", boltPath, err)
	}

	smsHandler := NewSmsHandler(createLogger(), createSender())
	http.HandleFunc("POST /notifier/api/send/{recipient}", smsHandler.Handle)

	dirName, ok := os.LookupEnv(envServeLocal)
	if ok {
		http.Handle("/notifier/app/", http.StripPrefix("/notifier/app/", http.FileServer(http.Dir(dirName))))
	}

	InstallSignalHandler(db, &dbOpened)

	err = http.ListenAndServe(":5100", nil)
	if err != nil {
		fmt.Println(err)
	}
}
