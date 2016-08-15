package commands

import (
	"github.com/asteris-llc/pony/tf"

	"github.com/spf13/cobra"
)

func (c *Command) addDestroySub() {
	var stateFile string

	dCmd := &cobra.Command{
		Use: "destroy",
		Short: "Destroy infrastructure",
		Long: "Destroy infrastructure",
		RunE: func (cmd *cobra.Command, args []string) error {
			return c.tf.Destroy(stateFile)
		},
	}

	dCmd.Flags().StringVarP(&stateFile, "state", "s", tf.StatePath, "Path to environment state")

	c.root.AddCommand(dCmd)
}	
