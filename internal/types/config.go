package types

// EventsConfig holds the events-service service configuration
type EventsConfig struct {
	ServiceName string `mapstructure:"service_name"`

	Client struct {
		Strategy string `mapstructure:"strategy"`
	} `mapstructure:"client"`

	Pulsar struct {
		Broker struct {
			URL string `mapstructure:"url"`
		} `mapstructure:"broker"`
	} `mapstructure:"pulsar"`

	Event struct {
		Producer struct {
			Topic string `mapstructure:"topic"`
		} `mapstructure:"producer"`
		Consumer struct {
			Subscription struct {
				Name string `mapstructure:"name"`
			} `mapstructure:"subscription"`
		} `mapstructure:"consumer"`
	} `mapstructure:"event"`

	GRPC struct {
		Address string `yaml:"address"`
	} `yaml:"grpc"`

	HTTP struct {
		Address string `yaml:"address"`
	} `yaml:"http"`
}
