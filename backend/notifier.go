package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"notifier/controller"
	"notifier/logic"
	"notifier/repo"
	"notifier/sms"
	"notifier/tools"
	"os"
	"time"

	_ "notifier/docs"
	_ "time/tzdata"

	httpSwagger "github.com/swaggo/http-swagger"
)

const envApiKey = "IFTTT_API_KEY"
const envNotifierApiKey = "NOTIFIER_API_KEY"
const envDbPath string = "DB_PATH"
const envServeLocal string = "LOCALDIR"
const envSwaggerUrl = "SWAGGER_URL"
const envClientTimeZone = "MN_CLIENT_TZ"
const envAddressBook = "MN_ADDR_BOOK"
const envMailSubject = "MN_MAIL_SUBJECT"
const authHeaderName = "X-Token"
const ERROR_EXIT = 42
const ERROR_OK = 0

func createLogger() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

func createAddressBook() sms.SmsAddressBook {
	var addrBook *sms.AddressBook
	var addrBookJsonByte []byte
	var addrBookJson string
	var err error

	addrBookB64, ok := os.LookupEnv(envAddressBook)
	if !ok {
		panic(fmt.Errorf("unable to read address book: %v", err))
	}

	addrBookJsonByte, err = base64.StdEncoding.DecodeString(addrBookB64)
	if err != nil {
		panic(fmt.Errorf("unable to read address book: %v", err))
	}

	addrBookJson = string(addrBookJsonByte)
	addrBook, err = sms.NewAddressBookFromJson(addrBookJson)
	if err != nil {
		panic(fmt.Errorf("unable to parse address book: %v", err))
	}

	addrBook.SetDefaultRecipientIds([]string{"0D69B617-12D0-4491-ADD8-D103CF3925A1"})

	apiKey, ok := os.LookupEnv(envApiKey)
	if !ok {
		dummy := sms.NewDummySender()
		addrBook.AddSender(sms.TypeIFTTT, dummy)
	} else {
		ifft := sms.NewIftttSender(apiKey)
		addrBook.AddSender(sms.TypeIFTTT, ifft)
	}

	addrBook.SetDefaultType(sms.TypeIFTTT)

	mailSender, err := sms.NewMailNotifierFromEnvironment()
	if err == nil {
		mailSubject, ok := os.LookupEnv(envMailSubject)
		if ok {
			mailSender.SetSubject(mailSubject)
		}
		addrBook.AddSender(sms.TypeMail, mailSender)
		log.Println("Mail notifier added")
	} else {
		log.Printf("Mail notifier not added: %v", err)
	}

	return addrBook
}

func createAuthSecret() *tools.AuthSecret {
	return &tools.AuthSecret{
		Secret:     os.Getenv(envNotifierApiKey),
		HeaderName: authHeaderName,
	}
}

func determineClientTZFromEnvironment() {
	tools.SetClientTZ(time.UTC)

	timeZoneStr, ok := os.LookupEnv(envClientTimeZone)
	if !ok {
		log.Printf("No time Zone set. Using UTC. This might not be what you want")
	} else {
		tz, err := time.LoadLocation(timeZoneStr)
		if err != nil {
			log.Printf("Wrong time zone: %v. Using UTC instead. This might not be what you want", err)
		} else {
			tools.SetClientTZ(tz)
		}
	}

	log.Printf("Using client time zone '%s'", tools.ClientTZ())
}

func determineSwaggerURL() string {
	swaggerUrl, ok := os.LookupEnv(envSwaggerUrl)
	if !ok {
		swaggerUrl = "http://localhost:5100/notifier/api/swagger/doc.json"
	}

	return swaggerUrl
}

func run() int {
	dbOpened := false

	determineClientTZFromEnvironment()

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

	metricCollector := tools.NewMetricsCollector()
	metricCollector.Start()
	defer func() { metricCollector.Stop() }()

	smsAddressBook := createAddressBook()
	smsLogger := createLogger()
	authSecret := createAuthSecret()
	authWrapper := tools.MakeWrapper(*authSecret, smsLogger, tools.ApiKeyAuthenticator)
	smsController := controller.NewSmsController(smsLogger, smsAddressBook)
	smsController.AddHandlersWithAuth(authWrapper)

	notificationController := controller.NewNotificationController(dbl, smsAddressBook, createLogger())
	notificationController.AddHandlers()

	reminderController := controller.NewReminderController(dbl, smsAddressBook, createLogger())
	reminderController.AddHandlers()

	infoController := controller.NewGeneralController(dbl, createLogger(), metricCollector)
	infoController.AddHandlers()

	logic.StartWarner(dbl, smsAddressBook, time.NewTicker(60*time.Second), createLogger(), metricCollector.AddEvent)

	dirName, ok := os.LookupEnv(envServeLocal)
	if ok {
		log.Println("Serving webapp locally")
		http.Handle("/notifier/app/", http.StripPrefix("/notifier/app/", http.FileServer(http.Dir(dirName))))
	}

	http.HandleFunc("/notifier/api/swagger/", httpSwagger.Handler(httpSwagger.URL(determineSwaggerURL())))

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
