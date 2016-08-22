package main

import (
	"github.com/asteris-llc/pony/commands"
)

const Name = "pony"
const Version = "0.0.0"

func main() {
	root := commands.Init()
	root.Execute()

}
