package config

import (
	"crypto/tls"
	"flag"
	"log"
	"time"
)

type Config struct {
	workDir                   string
	Server                    string
	Key                       []byte
	Certificate               []byte
	cert                      tls.Certificate
	LastEid                   int64
	MaxInputReportingInterval time.Duration
}

var config Config

var (
	DefaultMaxInputReportingInterval time.Duration
	DefaultServer                    string
	DefaultWorkDir                   string
	build                            string
)

func init() {
	flag.StringVar(&DefaultWorkDir, "work_dir", defaultWorkDir(), "Path to productimon working dir")
	flag.StringVar(&DefaultServer, "server", "127.0.0.1:4201", "Server Address (this will get overriden by config file, if exists)")
	if build == "DEBUG" {
		flag.DurationVar(&DefaultMaxInputReportingInterval, "max_input_reporting_interval", 5*time.Second, "Maximum duration to split an activity event (shorter means more accurate)")
	} else {
		flag.DurationVar(&DefaultMaxInputReportingInterval, "max_input_reporting_interval", 60*time.Second, "Maximum duration to split an activity event (shorter means more accurate)")
	}
}

func newConfig() *Config {
	return &Config{
		Server:                    DefaultServer,
		MaxInputReportingInterval: DefaultMaxInputReportingInterval,
		workDir:                   DefaultWorkDir,
	}
}

func (c *Config) Cert() tls.Certificate {
	if len(c.cert.Certificate) == 0 {
		var err error
		if c.cert, err = tls.X509KeyPair(c.Certificate, c.Key); err != nil {
			log.Println(err)
		}
	}
	return c.cert
}
