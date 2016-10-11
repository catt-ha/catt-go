package mqtt

type Config struct {
	Broker   string `toml:"broker"`
	ItemBase string `toml:"item_base"`
	ClientId string `toml:"client_id"`
	Tls      bool   `toml:"tls"`
}
