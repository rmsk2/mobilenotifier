//go:build noauth
// +build noauth

package main

import (
	"log"
	"notifier/tools"
)

func createAuthWrapper() tools.AuthWrapperFunc {
	log.Println("Using ********* NO AUTHENTICATION!!!!! *********")
	return tools.NullAuthenticator
}
