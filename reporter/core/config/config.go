package config

import (
	"crypto/tls"
	"flag"
	"log"
	"strings"
	"time"

	"git.yiad.am/productimon/internal"
)

type Config struct {
	workDir                   string
	Server                    string
	Key                       []byte
	Certificate               []byte
	cert                      tls.Certificate
	LastEid                   int64
	MaxInputReportingInterval time.Duration
	TrackingOptions           map[string]bool
}

type Options []string

var (
	DefaultMaxInputReportingInterval time.Duration
	DefaultServer                    string
	DefaultWorkDir                   string
	DefaultOptions                   Options
)

func (opts *Options) init() {
	*opts = []string{"foreground_program", "autorun"}
}

func (opts *Options) Set(opt string) error {
	*opts = append(*opts, opt)
	return nil
}

func (opts *Options) String() string {
	return strings.Join(*opts, ",")
}

func (opts *Options) Map() map[string]bool {
	trackingOpts := make(map[string]bool)
	for _, opt := range *opts {
		trackingOpts[opt] = true
	}
	return trackingOpts
}

func init() {
	DefaultOptions.init()
	flag.StringVar(&DefaultWorkDir, "work_dir", defaultWorkDir(), "Path to productimon working dir")
	flag.StringVar(&DefaultServer, "server", "my.productimon.com", "Server Address (this will get overriden by config file, if exists)")
	if internal.IsDebugBuild() {
		flag.DurationVar(&DefaultMaxInputReportingInterval, "max_input_reporting_interval", 5*time.Second, "Maximum duration to split an activity event (shorter means more accurate)")
	} else {
		flag.DurationVar(&DefaultMaxInputReportingInterval, "max_input_reporting_interval", 60*time.Second, "Maximum duration to split an activity event (shorter means more accurate)")
	}
	flag.Var(&DefaultOptions, "default_options", "Default tracking options to be enabled (default "+DefaultOptions.String()+")")
}

func newConfig() *Config {
	return &Config{
		Server:                    DefaultServer,
		MaxInputReportingInterval: DefaultMaxInputReportingInterval,
		workDir:                   DefaultWorkDir,
		TrackingOptions:           DefaultOptions.Map(),
	}
}

func (c *Config) Cert() tls.Certificate {
	if len(c.cert.Certificate) == 0 {
		c.ReloadCert()
	}
	return c.cert
}

func (c *Config) ReloadCert() {
	var err error
	if c.cert, err = tls.X509KeyPair(c.Certificate, c.Key); err != nil {
		log.Println(err)
	}
}

func (c *Config) SetOptions(options ...string) {
	// clear the map first
	c.TrackingOptions = make(map[string]bool)
	for _, opt := range options {
		c.TrackingOptions[opt] = true
	}
	log.Printf("Options updated: %v", c.TrackingOptions)
}

func (c *Config) IsOptionEnabled(opt string) bool {
	return c.TrackingOptions[opt]
}
