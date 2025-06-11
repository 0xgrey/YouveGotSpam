package utils

import (
	"log"

	"github.com/BurntSushi/toml"
)

type SpoofEmail struct {
	From         string
	To           string
	Subject      string
	Body         string
	Mimetype     string
	TargetDomain string
}

func ParseSpoofEmail(configFile string) SpoofEmail {
	var config SpoofEmail
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalf("Error parsing TOML: %v", err)
	}
	return config
}
