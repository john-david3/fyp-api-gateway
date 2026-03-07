package config

type RenderModel struct {
	Connections []Connection
}

type (
	GatewayConfig struct {
		Connections Connection `yaml:"connections" mapstructure:"connections"`
	}

	Connection struct {
		Routes []Routes `yaml:"routes"  mapstructure:"routes"`
	}

	Routes struct {
		Path      string    `yaml:"path"       mapstructure:"path"`
		Url       string    `yaml:"url"        mapstructure:"url"`
		Auth      bool      `yaml:"auth"       mapstructure:"auth"`
		RateLimit RateLimit `yaml:"rate-limit" mapstructure:"rate_limit"`
		ZoneName  string    `yaml:"zone-name"  mapstructure:"zone-name"`
	}

	RateLimit struct {
		Zone int `yaml:"zone" mapstructure:"zone"`
		Rate int `yaml:"rate" mapstructure:"rate"`
	}
)
