package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func interactiveScanCreds(defaultserver string) (server, username, password, deviceName string) {
	fmt.Printf("Server? [default=\"%s\"]", defaultserver)
	fmt.Scanln(&server)
	if server == "" {
		server = defaultserver
	}
	fmt.Printf("Username? ")
	fmt.Scanln(&username)
	fmt.Printf("Password? ")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err == nil {
		password = string(bytePassword)
		fmt.Printf("\n")
	} else {
		// see https://github.com/golang/go/issues/11914#issuecomment-613715787
		// Cygwin/mintty/git-bash can't reach down to OS api on Windows
		// and returns "handle is invalid" error
		fmt.Scanln(&password)
	}
	fmt.Printf("Device name? ")
	fmt.Scanln(&deviceName)
	return
}
