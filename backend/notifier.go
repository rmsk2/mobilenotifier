package main

import (
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
const authHeaderName = "X-Token"
const ERROR_EXIT = 42
const ERROR_OK = 0

func createLogger() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

func createAddressBook() sms.SmsAddressBook {
	addrBook := sms.NewAddressBook()

	apiKey, ok := os.LookupEnv(envApiKey)
	if !ok {
		dummy := sms.NewDummySender()
		addrBook.AddSender(sms.TypeSMS, dummy)
	} else {
		ifft := sms.NewIftttSender(apiKey)
		addrBook.AddSender(sms.TypeSMS, ifft)
	}

	addrBook.SetDefaultType(sms.TypeSMS)

	mailSender, err := sms.NewMailNotifierFromEnvironment()
	if err == nil {
		addrBook.AddSender(sms.TypeMail, mailSender)
		sms.AddMailRecipients(addrBook)
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
	authWrapper := tools.NewAuthWrapper[tools.ApiKey](createAuthSecret(), smsLogger)
	smsController := controller.NewSmsController(smsLogger, smsAddressBook)
	smsController.Add(authWrapper)

	notificationController := controller.NewNotificationController(dbl, smsAddressBook, createLogger())
	notificationController.Add()

	reminderController := controller.NewReminderController(dbl, smsAddressBook, createLogger())
	reminderController.Add()

	infoController := controller.NewGeneralController(dbl, createLogger(), metricCollector)
	infoController.Add()

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
