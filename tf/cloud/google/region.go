package google

import (
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
)

func (g *Google) getRegion(project string) (string, error) {
	regions := []string{}
	ctx := context.Background()

	call := g.compute.Regions.List(project)
	if err := call.Pages(ctx, func(page *compute.RegionList) error {
		for _, v := range page.Items {
			regions = append(regions, v.Name)
		}
		return nil
	}); err != nil {
		return "", err
	}

	rsp, err := g.cli.Select("region", regions)
	if err != nil {
		return "", err
	}

	return rsp, nil
}
