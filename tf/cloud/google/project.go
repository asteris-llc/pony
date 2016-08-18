package google

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/cloudresourcemanager/v1beta1"
)

func (g *Google) getProject() (string, error) {
	// Read project list
	projects := make(map[string]string)
	projects, err := g.readProjects()
	if err != nil {
		// Reading project list failed. Ask for response
		g.cli.Println("\nThe ID of the project to apply any resources to")
		for {
			rsp, err := g.cli.AskRequired("Enter a value for project")
			if err != nil {
				return "", err
			}
			if g.validateProject(rsp) {
				return rsp, nil
			}
			g.cli.Printf("Invalid project name: %s\n", rsp)
		}
	} 
	
	// We have a list of projects. Have the user select one
	g.cli.Println("\nAvailable projects")
	g.cli.Println("------------------")
	for _, p := range projects  {
		g.cli.Println(p)
	}
	for {
		rsp, err := g.cli.AskRequired("Enter a value for project")
		if err != nil {
			return "", err
		}

		// Project is in list
		if _, ok := projects[rsp]; ok {
			return rsp, nil
		}
	}
}

func (g *Google) readProjects() (map[string]string, error) {
	rval := make(map[string]string)

	crm, err := cloudresourcemanager.New(g.client)
	if err != nil {
		return nil, err
	}

	call := crm.Projects.List()
	if err := call.Pages(oauth2.NoContext, func (page *cloudresourcemanager.ListProjectsResponse) error {
		for _, v := range page.Projects {
			rval[v.ProjectId] = v.ProjectId
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return rval, nil
}

// Validate that a project string is a valid GCE project
//
func (g *Google) validateProject(p string) bool {
	ctx := context.Background()

	_, err := g.compute.Projects.Get(p).Context(ctx).Do()
	if err != nil {
		return false
	}

	return true
}
