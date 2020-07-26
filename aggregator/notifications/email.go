package notifications

import (
	"fmt"
	"net/smtp"
	"strings"

	"git.yiad.am/productimon/third_party/smtp_login_auth"
)

type emailNotifier struct {
	serverAddr string
	auth       smtp.Auth
	sender     string
}

func (n emailNotifier) Name() string {
	return "email"
}

func (n emailNotifier) Notify(recipient, message string) error {
	to := []string{recipient}
	msg := []byte(fmt.Sprintf("From: Productimon <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: Productimon notification\r\n"+
		"\r\n%s\r\n", n.sender, recipient, message))
	return smtp.SendMail(n.serverAddr, n.auth, n.sender, to, msg)
}

func ensurePort(serverName string) string {
	if !strings.Contains(serverName, ":") {
		serverName += ":25"
	}
	return serverName
}

func NewEmailNotifier(serverName, authUsername, authPassword, sender string) Notifier {
	return &emailNotifier{
		serverAddr: ensurePort(serverName),
		auth:       smtp_login_auth.LoginAuth(authUsername, authPassword),
		sender:     sender,
	}
}

func NewEmailNoAuthNotifier(serverName, sender string) Notifier {
	return &emailNotifier{
		serverAddr: ensurePort(serverName),
		auth:       nil,
		sender:     sender,
	}
}
