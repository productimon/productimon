package notifications

import "errors"

var ErrNotRegistered = errors.New("notifications: notifier is not registered")

type Notifier interface {
	Name() string
	Notify(receipient, message string) error
}
