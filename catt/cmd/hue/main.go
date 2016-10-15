package main

import (
	"log"

	"github.com/catt-ha/catt-go/catt"
	"github.com/catt-ha/catt-go/catt/hue"
	"github.com/catt-ha/catt-go/catt/mqtt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Bus mqtt.Config `toml:"bus"`
}

func main() {
	cfg := &Config{}
	toml.DecodeFile("./config.toml", cfg)
	mq, err := mqtt.NewMqtt(cfg.Bus)
	if err != nil {
		log.Fatal(err)
	}
	h, err := hue.NewHue()
	if err != nil {
		log.Fatal(err)
	}
	br := catt.NewBridge(mq, h)

	br.Run()
}
