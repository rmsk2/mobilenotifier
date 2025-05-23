package main

import (
	"encoding/json"
	"fmt"
	"io"
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

func createLogger() (*log.Logger, error) {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime), nil
}

func createSender() sms.SmsSender {
	apiKey, ok := os.LookupEnv(envApiKey)
	if !ok {
		return sms.NewDummySender()
	} else {
		return sms.NewIftttSender(apiKey)
	}
}

type SmsMessage struct {
	Message string `json:"message"`
}

func makeSmsHandler(txt sms.SmsSender) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log, err := createLogger()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		recipient := r.PathValue("recipient")
		ok, err := txt.CheckRecipient(recipient)

		if err != nil {
			log.Printf("error accessing recipient info: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		if !ok {
			t := fmt.Sprintf("recipient '%s' is unknown", recipient)
			log.Println(t)
			http.Error(w, t, http.StatusBadRequest)
			return
		}

		log.Printf("Sending SMS to '%s'", recipient)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Unable to read body")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var m SmsMessage
		err = json.Unmarshal(body, &m)
		if err != nil {
			log.Printf("Unable to parse body: '%s'", string(body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = txt.Send(recipient, m.Message)
		if err != nil {
			log.Printf("Sending SMS failed: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		log.Printf("SMS with message '%s' successfully sent", m.Message)
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

	http.HandleFunc("POST /notifier/api/send/{recipient}", makeSmsHandler(createSender()))

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
