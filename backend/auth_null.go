//go:build noauth
// +build noauth

package main

import "notifier/tools"

func createAuthWrapper() tools.AuthWrapperFunc {
	return tools.NullAuthenticator
}
