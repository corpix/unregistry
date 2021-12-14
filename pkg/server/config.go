package server

import (
	"time"
)

type Config struct {
	Timeout     *TimeoutConfig     `yaml:"timeout"`
	IPExtractor *IPExtractorConfig `yaml:"ip-extractor"`
}

func (c *Config) Default() {
loop:
	for {
		switch {
		case c.Timeout == nil:
			c.Timeout = &TimeoutConfig{}
		case c.IPExtractor == nil:
			c.IPExtractor = &IPExtractorConfig{}
		default:
			break loop
		}
	}
}

//

type IPExtractorConfig struct {
	TrustCIDR []string `yaml:"trust-cidr"`
}

func (c *IPExtractorConfig) Default() {
loop:
	for {
		switch {
		case len(c.TrustCIDR) == 0:
			c.TrustCIDR = []string{"0.0.0.0/0"}
		default:
			break loop
		}
	}
}

//

type TimeoutConfig struct {
	Read  time.Duration
	Write time.Duration
}

func (c *TimeoutConfig) Default() {
loop:
	for {
		switch {
		case c.Read <= 0:
			c.Read = 5 * time.Second
		case c.Write <= 0:
			c.Write = 5 * time.Second
		default:
			break loop
		}
	}
}
