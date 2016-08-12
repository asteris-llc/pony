package plugin

import (
	awsprovider "github.com/hashicorp/terraform/builtin/providers/aws"
	googleprovider "github.com/hashicorp/terraform/builtin/providers/google"
	nullprovider "github.com/hashicorp/terraform/builtin/providers/null"
	templateprovider "github.com/hashicorp/terraform/builtin/providers/template"

	remoteexecprovisioner "github.com/hashicorp/terraform/builtin/provisioners/remote-exec"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

var InternalProviders = map[string]plugin.ProviderFunc{
	"aws": awsprovider.Provider,
	"google": googleprovider.Provider,
	"null": nullprovider.Provider,
	"template": templateprovider.Provider,
}

var InternalProvisioners = map[string]plugin.ProvisionerFunc{
	"remote-exec": func() terraform.ResourceProvisioner { return new(remoteexecprovisioner.ResourceProvisioner) },
}
