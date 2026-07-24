package tools

import (
	"os"
	"os/signal"
	"syscall"
)

type NotifierCancelFunc func()

var allCancelFuncs []NotifierCancelFunc = []NotifierCancelFunc{}

func AddCancelFunc(f NotifierCancelFunc) {
	allCancelFuncs = append(allCancelFuncs, f)
}

func InstallSignalHandler() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		for _, f := range allCancelFuncs {
			f()
		}
	}()
}
