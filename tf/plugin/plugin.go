package plugin

import (
	"fmt"
	"os/exec"

	tfplugin "github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/go-plugin"
	"github.com/mitchellh/osext"
	"github.com/spf13/cobra"
)

type Plugin struct {
	Type string
	Name string
}

func InitPluginCmd(root *cobra.Command) {
	p := Plugin{}

	pluginCmd := &cobra.Command{
		Use: "plugin",
		Short: "Run terraform plugin",
		Long: "Run terraform plugin",
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if p.Name == "" {
				return fmt.Errorf("Plugin name must be specified")
			}

			if p.Type == "" {
				return fmt.Errorf("Plugin type must be specified")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch p.Type {
			case "provider":
				pluginFunc, found := InternalProviders[p.Name]
				if !found {
					return fmt.Errorf("Could not load provider: %s", p.Name)
				}
				tfplugin.Serve(&tfplugin.ServeOpts{
					ProviderFunc: pluginFunc,
				})
			case "provisioner":
				pluginFunc, found := InternalProvisioners[p.Name]
				if !found {
					return fmt.Errorf("Could not load provisioner: %s", p.Name)
				}
				tfplugin.Serve(&tfplugin.ServeOpts{
					ProvisionerFunc: pluginFunc,
				})
			default:
				return fmt.Errorf("Invalid plugin type: %s", p.Type)
			}

			return nil
		},
	}

	pluginCmd.Flags().StringVar(&p.Type, "type", "", "Plugin type")
	pluginCmd.Flags().StringVar(&p.Name, "name", "", "Plugin name")

	root.AddCommand(pluginCmd)
}

func Providers() (map[string]terraform.ResourceProviderFactory, error) {
	result := make(map[string]terraform.ResourceProviderFactory)

	exepath, err := pluginPath()
	if err != nil {
		return nil, err
	}

	for k, _ := range InternalProviders {
		result[k] = providerFactory(k, exepath)
	}

	return result, nil
}

func Provisioners() (map[string]terraform.ResourceProvisionerFactory, error) {
	result := make(map[string]terraform.ResourceProvisionerFactory)

	exepath, err := pluginPath()
	if err != nil {
		return nil, err
	}

	for k, _ := range InternalProvisioners {
		result[k] = provisionerFactory(k, exepath)
	}

	return result, nil
}

func providerFactory(path string, exepath string) terraform.ResourceProviderFactory {
	var c plugin.ClientConfig
	c.Cmd = exec.Command(exepath, "plugin", "--type", "provider", "--name", path)
	c.HandshakeConfig = tfplugin.Handshake
	c.Managed = true
	c.Plugins = map[string]plugin.Plugin{
		"provider": &tfplugin.ResourceProviderPlugin{},
		"provisioner": &tfplugin.ResourceProvisionerPlugin{},
	}
	client := plugin.NewClient(&c)

	return func() (terraform.ResourceProvider, error) {
		rpcClient, err := client.Client()
		if err != nil {
			return nil, err
		}

		raw, err := rpcClient.Dispense(tfplugin.ProviderPluginName)
		if err != nil {
			return nil, err
		}

		return raw.(terraform.ResourceProvider), nil
	}
}

func provisionerFactory(path string, exepath string) terraform.ResourceProvisionerFactory {
	var c plugin.ClientConfig
	c.Cmd = exec.Command(exepath, "plugin", "--type", "provisioner", "--name", path)
	c.Managed = true
	c.HandshakeConfig = tfplugin.Handshake
	c.Plugins = map[string]plugin.Plugin{
		"provider": &tfplugin.ResourceProviderPlugin{},
		"provisioner": &tfplugin.ResourceProvisionerPlugin{},
	}
	client := plugin.NewClient(&c)

	return func() (terraform.ResourceProvisioner, error) {
		rpcClient, err := client.Client()
		if err != nil {
			return nil, err
		}

		raw, err := rpcClient.Dispense(tfplugin.ProvisionerPluginName)
		if err != nil {
			return nil, err
		}

		return raw.(terraform.ResourceProvisioner), nil
	}
}

func pluginPath() (string, error) {
	return osext.Executable()
}

// Discover()
//   Discover plugins. Search CWD and the executables path for glob
//
/*
func (d *Deploy) Discover(glob string) *map[string]string {
	m := make(map[string]string)

	if err := d.discover(".", glob, &m); err != nil {
		log.Fatal(err.Error())
	}

	exePath, err := osext.Executable()
	if err != nil {
		log.Fatal(err.Error())
	}

	if err = d.discover(filepath.Dir(exePath), glob, &m); err != nil {
		log.Fatal(err.Error())
	}

	return &m
}

func (d *Deploy) discover(path string, glob string, m *map[string]string) error {
	var err error

	log.WithFields(log.Fields{
		"path": path,
		"glob": glob,
	}).Debug("Run discover")

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	matches, err := filepath.Glob(filepath.Join(path, glob))
	if err != nil {
		return err
	}

	if *m == nil {
		*m = make(map[string]string)
	}

	for _, match := range matches {
		file := filepath.Base(match)

		if idx := strings.Index(file, "."); idx >= 0 {
			file = file[:idx]
		}

		parts := strings.SplitN(file, "-", 3)
		if len(parts) != 3 {
			continue
		}

		(*m)[parts[2]] = match
	}

	return nil
}
*/
