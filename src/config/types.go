package config

type (
	GatewayConfig struct {
		Connections []Connections `yaml:"connections" mapstructure:"connections"`
	}

	Connections struct {
		Host   string   `yaml:"host"    mapstructure:"host"`
		Port   int      `yaml:"port"    mapstructure:"port"`
		Routes []Routes `yaml:"routes"  mapstructure:"routes"`
	}

	Routes struct {
		Path      string    `yaml:"path"       mapstructure:"path"`
		Upstream  Upstream  `yaml:"upstream"   mapstructure:"upstream"`
		RateLimit RateLimit `yaml:"rate-limit" mapstructure:"rate_limit"`
		Auth      Auth      `yaml:"auth"       mapstructure:"auth"`
	}

	Upstream struct {
		Name string `yaml:"name" mapstructure:"name"`
		Port int    `yaml:"port" mapstructure:"port"`
	}

	RateLimit struct {
		Zone int `yaml:"zone" mapstructure:"zone"`
		Rate int `yaml:"rate" mapstructure:"rate"`
	}

	Auth struct {
		Basic bool `yaml:"basic" mapstructure:"basic"`
		JWT   bool `yaml:"jwt"   mapstructure:"jwt"`
	}
)

type (
	ConfigMetadata struct {
		Version   string `yaml:"version"   mapstructure:"version"`
		Checksum  string `yaml:"checksum"  mapstructure:"checksum"`
		Timestamp string `yaml:"timestamp" mapstructure:"timestamp"`
	}

	ConfigPayload struct {
		Version string `yaml:"version" mapstructure:"version"`
		Config  string `yaml:"config"  mapstructure:"config"`
	}
)
