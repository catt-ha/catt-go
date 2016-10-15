package main

import (
	"flag"

	"github.com/Sirupsen/logrus"
	"github.com/catt-ha/catt-go/catt"
	"github.com/catt-ha/catt-go/catt/hue"
	"github.com/catt-ha/catt-go/catt/mqtt"

	"github.com/BurntSushi/toml"
)

var log = logrus.New()

type Config struct {
	Bus mqtt.Config `toml:"bus"`
}

func main() {
	cfgPath := flag.String("c", "./config.toml", "path to config file")
	flag.Parse()
	cfg := &Config{}
	toml.DecodeFile(*cfgPath, cfg)
	mq, err := mqtt.NewMqtt(cfg.Bus)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("error starting mqtt connection")
	}
	h, err := hue.NewHue()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("error starting hue connection")
	}
	br := catt.NewBridge(mq, h)

	br.Run()
}
