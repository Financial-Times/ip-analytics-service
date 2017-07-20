package config

type Config struct {
	Connection ConnectionConfig `json:"connection"`
}

type ConnectionConfig struct {
	Host string `json:"host"`
}

func connectionConfigToAddress(c ConnectionConfig) (RabbitAddress, error) {
	add := RabbitAddress{
		Host: c.host,
	}
	log.Printf("RabbitMQ Ready to Dial")
	return add, nil
}
