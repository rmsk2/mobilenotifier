package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"notifier/sms"
	"os"
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

func main() {
	txt := createSender()

	http.HandleFunc("/notifier/api/test", func(w http.ResponseWriter, r *http.Request) {
		log, err := createLogger()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		log.Println("Looking for database")

		dbPath, ok := os.LookupEnv(envDbPath)
		if !ok {
			log.Printf("Environment variable '%s' not found in environment", envDbPath)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		log.Printf("DB Data file: %s\n", dbPath)

		data, err := os.ReadFile(dbPath)
		if err != nil {
			log.Println("Unable to open data base")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		log.Println("Database opened")

		w.Write(data)
	})

	http.HandleFunc("POST /notifier/api/send/{recipient}", func(w http.ResponseWriter, r *http.Request) {
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
	})

	dirName, ok := os.LookupEnv(envServeLocal)
	if ok {
		http.Handle("/notifier/app/", http.StripPrefix("/notifier/app/", http.FileServer(http.Dir(dirName))))
	}

	http.ListenAndServe(":5000", nil)
}
