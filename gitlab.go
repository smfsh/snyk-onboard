package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/spf13/viper"
)

func createGitLabRepo(name string) error {
	glUser := viper.Get("glUser").(string)
	glToken := viper.Get("glKey").(string)
	private := gitlab.PublicVisibility // Public by default, or false
	urlpath := path.Join(glUser, name)

	fmt.Printf("Attempting to create %s in %s's GitLab\n", name, glUser)

	client, err := gitlab.NewClient(glToken)
	if err != nil {
		return err
	}

	project, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:       gitlab.String(name),
		Visibility: gitlab.Visibility(private),
	})
	if err != nil && strings.Contains(err.Error(), "name: [has already been taken]") {
		fmt.Printf("%s already exists in this GitLab account\n", name)
		project, _, err = client.Projects.GetProject(urlpath, &gitlab.GetProjectOptions{})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		fmt.Printf("%s created successfully at %s\n", project.Name, project.WebURL)
	}

	err = pushUpstream(name, "gitlab", project.HTTPURLToRepo, glUser, glToken)
	if err != nil {
		return err
	}

	return nil
}
