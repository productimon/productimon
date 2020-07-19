// +build js

package config

import (
	"errors"
	"log"
)

func defaultWorkDir() string {
	return ""
}

func NewConfig() *Config {
	config := newConfig()
	// TODO: load from localstorage
	return config
}

func (c *Config) Save() error {
	log.Println("TODO: save config")
	return errors.New("NOT YET IMPLEMENTED")
}
