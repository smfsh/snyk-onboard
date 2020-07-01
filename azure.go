package main

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/git"

	"github.com/microsoft/azure-devops-go-api/azuredevops"

	"github.com/spf13/viper"
)

func createAzureRepo(name string) error {
	azOrg := viper.Get("azOrg").(string)
	azToken := viper.Get("azKey").(string)

	u, err := url.Parse("https://dev.azure.com")
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, azOrg)
	azURL := u.String()

	connection := azuredevops.NewPatConnection(azURL, azToken)
	ctx := context.Background()

	client, err := git.NewClient(ctx, connection)
	if err != nil {
		return err
	}

	repo, err := client.GetRepository(ctx, git.GetRepositoryArgs{
		RepositoryId: &name,
		Project:      &azOrg,
	})
	if err != nil && !strings.Contains(err.Error(), "TF401019") {
		return err
	}
	if repo == nil {
		repo, err = client.CreateRepository(ctx, git.CreateRepositoryArgs{
			GitRepositoryToCreate: &git.GitRepositoryCreateOptions{
				Name: &name,
			},
			Project: &azOrg,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s created successfully at %s\n", *repo.Name, *repo.WebUrl)
	} else {
		fmt.Printf("%s already exists in this Azure DevOps account\n", name)
	}

	err = pushUpstream(name, "azure", *repo.RemoteUrl, azOrg, azToken)
	if err != nil {

	}

	return nil
}
