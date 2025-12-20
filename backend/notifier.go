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
	"strconv"
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
const envExpectedTokenIssuer = "EXPECTED_TOKEN_ISSUER"
const envExpectedTokenAudience = "EXPECTED_TOKEN_AUDIENCE"
const envExpectedTokenTtl = "TOKEN_TTL"
const authHeaderName = "X-Token"
const ERROR_EXIT = 42
const ERROR_OK = 0

func createLogger() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

func saveEnvirnomentInDB(addrSaver *sms.AddressSaver, dbl repo.DBSerializer, generator func(repo.DbType) *repo.BBoltAddrBookRepo) {
	writeRepo := repo.LockAndGetRepoRW(dbl, generator)
	defer func() { dbl.Unlock() }()

	err := addrSaver.BBoltSave(writeRepo)
	if err != nil {
		panic(fmt.Errorf("unable to save address book data read from environment into DB: %v", err))
	}

	log.Println("Saved address book from environment to database")
}

func createAddressBook(dbl repo.DBSerializer, generator func(repo.DbType) *repo.BBoltAddrBookRepo) sms.SmsAddressBook {
	var addrBook sms.SmsAddressBook
	var addressSaver *sms.AddressSaver
	var addrBookJsonByte []byte
	var addrBookJson string
	var err error

	addrBookB64, addrBookInEnvironment := os.LookupEnv(envAddressBook)

	if addrBookInEnvironment {
		addrBookJsonByte, err = base64.StdEncoding.DecodeString(addrBookB64)
		if err != nil {
			panic(fmt.Errorf("unable to read address book from environment variable: %v", err))
		}
		log.Println("Read address book from environment")

		addrBookJson = string(addrBookJsonByte)
		addressSaver, err = sms.NewAddressSaverFromJson(addrBookJson)
		if err != nil {
			panic(fmt.Errorf("unable to parse address book in environment: %v", err))
		}
	}

	addrBook = sms.NewDBAddressBook(dbl, generator)

	if addrBookInEnvironment {
		saveEnvirnomentInDB(addressSaver, dbl, generator)
		log.Println("Address book in environment merged into DB")
	}

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

func getTokenDefinitionsFromEnv() {
	temp, ok := os.LookupEnv(envExpectedTokenIssuer)
	if ok {
		tools.ExpectedJwtIssuer = temp
	}

	temp, ok = os.LookupEnv(envExpectedTokenAudience)
	if ok {
		tools.ExpectedJwtAudience = temp
	}

	temp, ok = os.LookupEnv(envExpectedTokenTtl)
	if ok {
		ttl, err := strconv.ParseInt(temp, 10, 64)
		if (err == nil) && (ttl > 0) {
			tools.TokenTtl = ttl
		}
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
	getTokenDefinitionsFromEnv()

	boltPath, ok := os.LookupEnv(envDbPath)
	if !ok {
		log.Printf("environment variable '%s' not found in environment", envDbPath)
		return ERROR_EXIT
	}

	rawDB, err := repo.InitDB(&dbOpened, boltPath)
	if err != nil {
		log.Println(err)
		return ERROR_EXIT
	}
	defer func() {
		rawDB.Close()
		log.Println("bbolt DB closed")
	}()

	dbl := repo.NewBoltDBLocker(rawDB)
	dblAddr := repo.NewBoltDBLocker(rawDB)

	metricCollector := tools.NewMetricsCollector()
	metricCollector.Start()
	defer func() { metricCollector.Stop() }()

	smsAddressBook := createAddressBook(dblAddr, repo.NewBBoltAddressBookRepo)

	authSecret := createAuthSecret()
	//authWrapper := tools.MakeWrapper(*authSecret, createLogger(), tools.ApiKeyAuthenticator)
	authWrapper := tools.MakeWrapper(*authSecret, createLogger(), tools.JwtHs256Authenticator)
	//authWrapper := tools.NullAuthenticator

	smsLogger := createLogger()
	smsController := controller.NewSmsController(smsLogger, smsAddressBook)
	smsController.AddHandlersWithAuth(authWrapper)

	notificationController := controller.NewNotificationController(dbl, createLogger(), repo.NewBBoltNotificationRepo)
	notificationController.AddHandlersWithAuth(authWrapper)

	reminderController := controller.NewReminderController(dbl, smsAddressBook, createLogger(), repo.NewBBoltReminderRepo)
	reminderController.AddHandlersWithAuth(authWrapper)

	addrBookController := controller.NewAddressBookController(dblAddr, dbl, createLogger(), repo.NewBBoltAddressBookRepo)
	addrBookController.AddHandlersWithAuth(authWrapper)

	infoController := controller.NewGeneralController(dbl, createLogger(), metricCollector)
	infoController.AddHandlersWithAuth(authWrapper)

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
