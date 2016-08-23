package cloud

import (
	"sort"

	google "github.com/asteris-llc/pony/tf/cloud/google"

	"github.com/asteris-llc/pony/cli"
)

type CloudProvider interface {
	Root() []byte
	GetProviderVars() (map[string]string, error)
	GetConfig(string) ([]byte, error)
}

type CloudList struct {
	keys      []string
	providers map[string]CloudProvider
}

func New(cli *cli.Cli) *CloudList {
	rval := &CloudList{
		providers: map[string]CloudProvider{
			"google": google.New(cli),
		},
	}
	rval.keys = make([]string, len(rval.providers))
	i := 0
	for k := range rval.providers {
		rval.keys[i] = k
		i++
	}
	sort.Strings(rval.keys)

	return rval
}

func (cl *CloudList) Keys() []string {
	return cl.keys
}

func (cl *CloudList) GetProvider(name string) CloudProvider {
	if p, ok := cl.providers[name]; ok {
		return p
	}

	return nil
}
