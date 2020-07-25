// +build js

package config

import (
	"encoding/json"
	"fmt"
	"log"
	"syscall/js"
)

var localStorage js.Value

func init() {
	localStorage = js.Global().Get("localStorage")
}

func defaultWorkDir() string {
	return ""
}

func NewConfig() *Config {
	config := newConfig()
	if cfgValue := localStorage.Get("config"); !cfgValue.IsUndefined() {
		if cfgStr := cfgValue.String(); len(cfgStr) > 0 {
			if err := json.Unmarshal([]byte(cfgStr), &config); err != nil {
				log.Printf("Failed to unmarshal localStorage config: %v", err)
			}
		}
	}
	return config
}

func (c *Config) Save() error {
	log.Println("Saving config")
	cfgBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}
	localStorage.Set("config", string(cfgBytes))
	return nil
}
