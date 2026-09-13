package config

import (
	"fmt"
	"planeshift/objects"
	"strings"
)

type (
	Config struct {
		VersionDetail    objects.Version
		VersionJSON      string
		OutputFormat     string
		FormatOverridden bool
		NoHeaders        bool
		CACert           string
		CABundle         string
		LogLevel         string
		LogFile          string
		ConfigVersion    string
	}

	Outputtable interface {
		ToJSON() string
		ToYAML() string
		ToGRON() string
		ToTEXT(noHeaders bool) string
	}
)

func (c *Config) outputData(data Outputtable) string {
	switch strings.ToLower(c.OutputFormat) {
	case "raw":
		return fmt.Sprintf("%#v", data)
	case "json":
		return data.ToJSON()
	case "gron":
		return data.ToGRON()
	case "yaml":
		return data.ToYAML()
	case "text", "table":
		return data.ToTEXT(c.NoHeaders)
	default:
		return ""
	}
}

func (c *Config) OutputData(data Outputtable) {
	fmt.Println(c.outputData(data))
}
