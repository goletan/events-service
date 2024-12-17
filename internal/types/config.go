package types

// EventsConfig holds the events service configuration
type EventsConfig struct {
	Events struct {
		ServiceName string `mapstructure:"service_name"`
		Pulsar      struct {
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
	} `mapstructure:"events"`
}
