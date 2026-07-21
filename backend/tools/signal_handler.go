package tools

import (
	"os"
	"os/signal"
	"syscall"
)

type CleanUpFunc func()

var allCleanUpFuncs []CleanUpFunc = []CleanUpFunc{}

func AddCleanUpFunc(f CleanUpFunc) {
	allCleanUpFuncs = append(allCleanUpFuncs, f)
}

func InstallSignalHandler() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		for _, f := range allCleanUpFuncs {
			f()
		}
	}()
}
