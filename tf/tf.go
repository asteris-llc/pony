package tf

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/config/module"
	"github.com/hashicorp/terraform/terraform"

	"github.com/asteris-llc/pony/cli"
)

type Tf struct {
	ctx	*terraform.Context
	m	*module.Tree
	config	*config.Config
	cli	*cli.Cli
	cloud	string
}

func New() *Tf {
	return &Tf{
		cli: cli.New(os.Stdin, os.Stdout),
	}
}

func (t *Tf) String() string {
	rval := bytes.NewBufferString("Tf structure:\n")
	rval.WriteString(fmt.Sprintf("  Cloud Provider: %s\n", t.cloud))

	return rval.String()
}
