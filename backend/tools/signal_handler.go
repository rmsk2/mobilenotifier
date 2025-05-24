package tools

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	bolt "go.etcd.io/bbolt"
)

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
