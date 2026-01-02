package config

type (
	GatewayConfig struct {
		Connection []Connection `yaml:"connection" mapstructure:"connection"`
	}

	Connection struct {
		Host   string  `yaml:"host"    mapstructure:"host"`
		Port   int     `yaml:"port"    mapstructure:"port"`
		Routes []Route `yaml:"routes"  mapstructure:"routes"`
	}

	Route struct {
		Path     string   `yaml:"path"     mapstructure:"path"`
		Upstream Upstream `yaml:"upstream" mapstructure:"upstream"`
	}

	Upstream struct {
		Name string `yaml:"name" mapstructure:"name"`
		Port int    `yaml:"port" mapstructure:"port"`
	}
)
