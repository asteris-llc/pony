package google

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/asteris-llc/pony/cli"

	"github.com/hashicorp/terraform/helper/pathorcontents"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/cloudresourcemanager/v1beta1"
	"google.golang.org/api/compute/v1"
)

type Google struct {
	cli                  *cli.Cli
	client               *http.Client
	cloudResourceManager *cloudresourcemanager.Service
	compute              *compute.Service
}

type accountFile struct {
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientId     string `json:"client_id"`
}

func New(cli *cli.Cli) *Google {
	return &Google{
		cli: cli,
	}
}

func (g *Google) Repo() string {
	return "github.com/ChrisAubuchon/pony-config/gce/base"
}

func (g *Google) Base() string {
	return "gce.tf"
}

func (g *Google) GetProviderVars() (map[string]string, error) {
	rval := make(map[string]string)

	// Get credentials file
	g.cli.Println("\nPath to the JSON file for GCE credentials")
	rsp, err := g.cli.AskRequiredWithDefault("Enter a value for credentials", "account.json")
	if err != nil {
		return nil, err
	}
	rval["credentials"] = rsp

	// Authenticate to google
	if err := g.readCredentials(rsp); err != nil {
		return nil, err
	}

	project, err := g.getProject()
	if err != nil {
		return nil, err
	}
	rval["project"] = project

	// Select region
	region, err := g.getRegion(project)
	if err != nil {
		return nil, err
	}
	rval["region"] = region
	fmt.Printf("Region: %s", region)

	// Select zones
	zones, err := g.getZones(project, region)
	if err != nil {
		return nil, err
	}
	rval["zones"] = zones
	fmt.Printf("Zones: %s", zones)

	return rval, nil
}

func (g *Google) readCredentials(path string) error {
	clientScopes := []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/ndev.clouddns.readwrite",
		"https://www.googleapis.com/auth/devstorage.full_control",
	}

	contents, _, err := pathorcontents.Read(path)
	if err != nil {
		return fmt.Errorf("Error reading credentials")
	}

	// Decode the credentials file
	af := new(accountFile)
	r := strings.NewReader(contents)
	dec := json.NewDecoder(r)
	if err := dec.Decode(af); err != nil {
		return fmt.Errorf("Error decoding credentials")
	}

	// Authenticate with credentials
	conf := jwt.Config{
		Email:      af.ClientEmail,
		PrivateKey: []byte(af.PrivateKey),
		Scopes:     clientScopes,
		TokenURL:   "https://accounts.google.com/o/oauth2/token",
	}
	g.client = conf.Client(oauth2.NoContext)

	g.compute, err = compute.New(g.client)
	if err != nil {
		return err
	}

	return nil
}
