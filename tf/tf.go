package tf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	golog "log"
	"os"
	"sort"
	"strings"

	"github.com/asteris-llc/pony/cli"
	"github.com/asteris-llc/pony/tf/cloud"
	"github.com/asteris-llc/pony/tf/plugin"

	"github.com/hashicorp/terraform/config/module"
	"github.com/hashicorp/terraform/terraform"
	tfcli "github.com/mitchellh/cli"

	log "github.com/sirupsen/logrus"
)

const (
	StatePath = "pony.state"
)

type Tf struct {
	context       *terraform.Context
	tree          *module.Tree
	cli           *cli.Cli
	tempDir       string
	globals       *variables
	state         *terraform.State
	cloudList     *cloud.CloudList
	cloudProvider cloud.CloudProvider
}

func New() *Tf {
	tf := new(Tf)

	tf.cli = cli.New(os.Stdin, os.Stdout)

	tdir, err := ioutil.TempDir("", "pony")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.RemoveAll(tdir); err != nil {
		log.Fatal(err)
	}
	tf.tempDir = tdir

	tf.globals = newVariables()
	tf.cloudList = cloud.New(tf.cli)

	return tf
}

func (tf *Tf) String() string {
	rval := bytes.NewBufferString("Tf structure:\n")

	return rval.String()
}

func (tf *Tf) Clean() {
	log.Debugf("Running clean()")
	os.RemoveAll(tf.tempDir)
	os.Remove(tf.tempDir)
}

func (tf *Tf) Context(destroy bool) error {
	golog.SetOutput(ioutil.Discard)

	providers, err := plugin.Providers()
	if err != nil {
		return err
	}

	provisioners, err := plugin.Provisioners()
	if err != nil {
		return err
	}

	ctx, err := terraform.NewContext(&terraform.ContextOpts{
		Destroy:      destroy,
		Hooks:        []terraform.Hook{NewUiHook(&tfcli.BasicUi{Writer: os.Stdout})},
		Module:       tf.tree,
		Providers:    providers,
		Provisioners: provisioners,
		State:        tf.state,
	})
	if err != nil {
		return err
	}

	tf.context = ctx

	return nil
}

func (tf *Tf) Plan() error {
	p, err := tf.context.Plan()
	if err != nil {
		return err
	}

	if false {
		tf.formatPlan(p)
	}

	return nil
}

func (tf *Tf) Apply() error {
	var s *terraform.State
	var applyErr error

	doneCh := make(chan struct{})
	ShutdownCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		s, applyErr = tf.context.Apply()
	}()

	select {
	case <-ShutdownCh:
		go tf.context.Stop()
	case <-doneCh:
	}

	tf.state = s

	if err := tf.writeState(); err != nil {
		return err
	}

	return applyErr
}

func (tf *Tf) writeState() error {
	f, err := os.Create(StatePath)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.MarshalIndent(tf.state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}

func (tf *Tf) formatPlan(p *terraform.Plan) {
	created := []string{}
	destroyed := []string{}
	updated := []string{}

	for _, m := range p.Diff.Modules {
		for r, _ := range m.Resources {
			name := extractResource(r)
			switch m.ChangeType() {
			case terraform.DiffCreate:
				created = append(created, name)
			case terraform.DiffDestroy:
				destroyed = append(destroyed, name)
			case terraform.DiffUpdate:
				updated = append(updated, name)
			}
		}
	}

	fmt.Println("Resources Created:")
	outputResources(created)

	fmt.Println("\nResources Destroyed:")
	outputResources(destroyed)

	fmt.Println("\nResources Updated:")
	outputResources(updated)

	fmt.Printf("\n%+v\n", p)
}

func outputResources(s []string) {
	if len(s) == 0 {
		fmt.Println("  <none>")
	} else {
		sort.Strings(s)
		for _, r := range s {
			fmt.Printf("  %s\n", r)
		}
	}
}

func extractResource(r string) string {
	fields := strings.Split(r, ".")

	return fields[len(fields)-1]
}
