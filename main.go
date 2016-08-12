package main

import (
	"github.com/asteris-llc/pony/commands"
)

func main() {
	root := commands.Init()
	root.Execute()
		
}

/*
	if err := tf.Context(); err != nil {
		fatal(tf, err)
	}

	if err := tf.Plan(); err != nil {
		fatal(tf, err)
	}

	if err := tf.Apply(); err != nil {
		fatal(tf, err)
	}

	log.Println(tf)
}
*/
