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
		Path     string   `yaml:"path"     mapstructure:"path"`
		Upstream Upstream `yaml:"upstream" mapstructure:"upstream"`
	}

	Upstream struct {
		Name string `yaml:"name" mapstructure:"name"`
		Port int    `yaml:"port" mapstructure:"port"`
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
