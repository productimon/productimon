// +build !js

package config

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

// get default work dir
// defaults to $HOME/.productimon
// if unavailable, fall back to cwd
// if still unavailable, return empty string
func defaultWorkDir() string {
	user, err := user.Current()
	if err != nil {
		log.Println(err)
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		return path
	}
	if user.HomeDir == "" {
		return ""
	}
	return filepath.Join(user.HomeDir, ".productimon")
}

func NewConfig() *Config {
	config := newConfig()
	if len(config.workDir) == 0 {
		panic("Cannot determine default working directory, please specify manually via --work_dir flag")
	}
	config.workDir = filepath.Clean(config.workDir)
	if err := os.MkdirAll(config.workDir, 0700); err != nil {
		panic(err)
	}
	file, err := os.Open(filepath.Join(config.workDir, "config.json"))
	if err == nil {
		config.TrackingOptions = make(map[string]bool)
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err = decoder.Decode(&config); err != nil {
			panic(err)
		}
	} else {
		log.Printf("Can't open config.json: %v", err)
	}
	return config
}

func (c *Config) Save() error {
	file, err := os.OpenFile(filepath.Join(c.workDir, "config.json"), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Failed to save config: %v\n", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(c); err != nil {
		log.Printf("Failed to encode config.json: %v\n", err)
		return err
	}
	log.Println("Config saved")
	return nil
}
