package main

import (
	"os"
	"time"

	"github.com/namsral/flag"
	"github.com/pkg/errors"
)

// Flag defaults.
const (
	defaultHost = "0.0.0.0"
	defaultPort = 9000
	//defaultFrontend = "https://data-entry.dev.drillinginfo.com/home"
	defaultFrontend        = "*" // temporarily allow access from all sites, for debugging
	defaultLogLevel        = "debug"
	defaultAppName         = "csv-server"
	defaultShutdownTimeout = 10 * time.Second
)

var flagErrorHandling = flag.ContinueOnError

// Config is application configuration.
type Config struct {
	AppName string     `json:"appName"`
	HTTP    HTTPConfig `json:"http"`
}

// Validate validates the app configuration.
func (c Config) Validate() []error {
	errs := c.HTTP.Validate()
	return errs
}

// HTTPConfig is configuration of an HTTP service.
type HTTPConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ShutdownTimeout time.Duration `json:"shutdownTimeout"`
	Frontend        string        `json:"frontend"`
}

// Validate validates HTTPConfig configuration values.
func (c HTTPConfig) Validate() []error {
	var errs []error
	if len(c.Host) == 0 {
		errs = append(errs, errors.Errorf("HTTPConfig requires a non-empty Host config value"))
	}
	if c.Port <= 0 {
		errs = append(errs, errors.Errorf("HTTPConfig requires a postive Port config value"))
	}
	return errs
}

type GrpcConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`

	AllowAllOrigins bool      `json:"allowAllOrigins"`
	AllowedOrigins  *[]string `json:"allowedOrigins"`

	UseWebsockets bool `json:"useWebsockets"`

	EnableTls       bool   `json:"enableTls"`
	TlsCertFilePath string `json:"tlsCertFilePath"`
	TlsKeyFilePath  string `json:"tlsKeyFilePath"`
}

func (c GrpcConfig) Validate() []error {
	var errs []error
	if len(c.Host) == 0 {
		errs = append(errs, errors.Errorf("GRPCConfig requires a non-empty Host config value"))
	}
	if c.Port <= 0 {
		errs = append(errs, errors.Errorf("GRPCConfig requires a postive Port config value"))
	}
	return errs
}

// AuthConfig is configuration of authentication via Azure etc.
type AuthConfig struct {
	//TenantID string `json:"tenantId"`
	JwksURI string `json:"jwksUri"`
}

// Validate validates AuthConfig configuration values.
func (c AuthConfig) Validate() []error {
	var errs []error

	if len(c.JwksURI) == 0 {
		errs = append(errs, errors.Errorf("AuthConfig requires a non-empty JwksURI config value"))
	}

	return errs
}

func GetConfig() (*Config, error) {
	var config Config

	flagset := flag.NewFlagSetWithEnvPrefix(defaultAppName, "", flagErrorHandling)

	// App
	flagset.StringVar(&config.AppName, "app-name", defaultAppName, "Service name.")

	// HTTP
	flagset.StringVar(&config.HTTP.Host, "host", defaultHost, "Host part of listening address.")
	flagset.IntVar(&config.HTTP.Port, "port", defaultPort, "Listening port.")
	flagset.DurationVar(&config.HTTP.ShutdownTimeout, "shutdown-timeout", defaultShutdownTimeout, "Shutdown timeout for http service.")
	flagset.StringVar(&config.HTTP.Frontend, "frontend", defaultFrontend, "Frontend address. (We name it explicitly, because CORS requires so.)")

	if err := flagset.Parse(os.Args[1:]); err != nil {
		return nil, errors.Wrap(err, "parsing flags")
	}

	// Validate the config.
	if errs := config.Validate(); len(errs) > 0 {
		return nil, errors.Errorf("invalid flag(s): %s", errs)
	}
	return &config, nil
}
