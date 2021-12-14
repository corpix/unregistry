package registry

type Config struct {
	Server   *ServerConfig   `yaml:"server"`
	Provider *ProviderConfig `yaml:"provider"`
}

func (c *Config) Default() {
loop:
	for {
		switch {
		case c.Server == nil:
			c.Server = &ServerConfig{}
		case c.Provider == nil:
			c.Provider = &ProviderConfig{}
		default:
			break loop
		}
	}
}
