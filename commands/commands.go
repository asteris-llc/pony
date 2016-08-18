package commands

import (
	"github.com/asteris-llc/pony/tf"
	"github.com/asteris-llc/pony/tf/plugin"

	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
)

type Command struct {
	root *cobra.Command
	tf *tf.Tf
	logLevel string
}

func Init() *Command {
	c := Command{}

	c.root = &cobra.Command{
		Use: "pony",
		Short: "Easy installer for mantl",
		Long: "Easy installer for mantl",
		SilenceUsage: true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			c.configureLogging()

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.tf.SelectCloud(); err != nil {
				return err
			}

			return c.tf.Create()
		},
	}
	c.tf = tf.New()

	c.root.PersistentFlags().StringVarP(&c.logLevel, "log-level", "l", "warn", "Logging level")

	plugin.InitPluginCmd(c.root)
	c.addDestroySub()

	return &c
}

func (c *Command) Execute() {
	if err := c.root.Execute(); err != nil {
		c.tf.Clean()
		log.Fatal(err)
	}
}

func (c *Command) configureLogging() {
	l, err := log.ParseLevel(c.logLevel)
        if err != nil {
                log.SetLevel(log.WarnLevel)
                log.Warnf("Invalid log level '%v'. Setting to WARN", c.logLevel)
        } else {
                log.SetLevel(l)
        }
}

