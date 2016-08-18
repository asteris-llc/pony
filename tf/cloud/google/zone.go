package google

import (
	"sort"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
)

func (g *Google) getZones(project, region string) (string, error) {
	zones := []string{}
	ctx := context.Background()

	filter := "name eq " + region + ".*"

	call := g.compute.Zones.List(project).Filter(filter)
	if err := call.Pages(ctx, func(page *compute.ZoneList) error {
		for _, v := range page.Items {
			zones = append(zones, v.Name)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(zones)

	rlist, err := g.cli.SelectMany("zones", zones)
	if err != nil {
		return "", err
	}

	return strings.Join(rlist, ","), nil
}
