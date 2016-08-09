package main

import (
	"github.com/asteris-llc/pony/tf"
	log "github.com/sirupsen/logrus"
)

func main() {
	tf := tf.New()

	log.SetLevel(log.DebugLevel)

	if err := tf.SelectCloud(); err != nil {
		log.Fatal(err)
	}

	if err := tf.LoadCloud(); err != nil {
		log.Fatal(err)
	}

	if err := tf.ReadVariables(); err != nil {
		log.Fatal(err)
	}

	log.Println(tf)
}
