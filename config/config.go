package config

import (
	"github.com/BurntSushi/toml"
	"github.com/vavas/go_services/services/gnats"
	"log"
)

//Server structure
type Server struct {
	Listen  string
	Server  string
	Timeout int
}

// Handler holds the reverse-proxy handler info.
type Handler struct {
	Name string

	// The host & path of the request
	RequestRegexes [][2]string `toml:"request_regexes"`

	// If the service is handled using a gnats-based messages, then it should have external_subject.
	ExternalSubject string `toml:"external_subject"`

	// The service needs to implement at least an internal Status function.
	InternalSubject string `toml:"internal_subject"`
}

//ReadConfig from config file
func ReadConfig(configFile string) (config *Config) {

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalf("Unable to parse config file (please specify with -c) - %v", err)
	}

	return config
}

// Config structure
type Config struct {
	Server   Server
	Gnats    *gnats.Config
	Handlers []*Handler
}
