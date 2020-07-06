// File contains functions used to create repositories on Azure DevOps.

package main

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/operations"

	"github.com/microsoft/azure-devops-go-api/azuredevops/core"

	"github.com/microsoft/azure-devops-go-api/azuredevops/git"

	"github.com/microsoft/azure-devops-go-api/azuredevops"

	"github.com/spf13/viper"
)

func createAzureRepo(name string) error {
	azOrg := viper.Get("azOrg").(string)
	azToken := viper.Get("azKey").(string)
	azProj := "Snyk" // TODO: toggle option for custom name

	u, err := url.Parse("https://dev.azure.com")
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, azOrg)
	azURL := u.String()

	connection := azuredevops.NewPatConnection(azURL, azToken)
	ctx := context.Background()

	coreClient, err := core.NewClient(ctx, connection)
	if err != nil {
		return err
	}

	project, err := coreClient.GetProject(ctx, core.GetProjectArgs{
		ProjectId: &azProj,
	})
	if err != nil && !strings.Contains(err.Error(), "TF200016") {
		return err
	}

	if project == nil {
		fmt.Printf("Unable to find project %s in Azure DevOps Org %s\n", azOrg, azProj)
		fmt.Printf("Creating project %s in Azure DevOps Org %s\n", azOrg, azProj)

		description := "A project for Snyk Repositories"
		visibility := core.ProjectVisibility(core.ProjectVisibilityValues.Public)
		capabilities := map[string]map[string]string{
			"versioncontrol": {
				"sourceControlType": "Git",
			},
			"processTemplate": {
				"templateTypeId": "b8a3a935-7e91-48b8-a94c-606d37c3e9f2",
			},
		}

		// TODO: add attributes to project to get Azure to create it
		qcp, err := coreClient.QueueCreateProject(ctx, core.QueueCreateProjectArgs{
			ProjectToCreate: &core.TeamProject{
				Name:         &azProj,
				Description:  &description,
				Visibility:   &visibility,
				Capabilities: &capabilities,
			},
		})
		if err != nil {
			return err
		}
		opClient := operations.NewClient(ctx, connection)

		for {
			op, err := opClient.GetOperation(ctx, operations.GetOperationArgs{
				OperationId: qcp.Id,
				PluginId:    qcp.PluginId,
			})
			if err != nil {
				return err
			}
			if *op.Status == operations.OperationStatusValues.Succeeded {
				fmt.Printf("Project %s successfully created\n", azProj)
				project, err = coreClient.GetProject(ctx, core.GetProjectArgs{
					ProjectId: &azProj,
				})
				if err != nil && !strings.Contains(err.Error(), "TF200016") {
					return err
				}
				break
			} else {
				fmt.Printf("Creating project...\n")
				time.Sleep(5 * time.Second)
			}
		}
		fmt.Printf("Created project %s in Azure DevOps Org %s\n", *project.Name, azOrg)
	}

	client, err := git.NewClient(ctx, connection)
	if err != nil {
		return err
	}

	repo, err := client.GetRepository(ctx, git.GetRepositoryArgs{
		RepositoryId: &name,
		Project:      &azProj,
	})
	if err != nil && !strings.Contains(err.Error(), "TF401019") {
		return err
	}
	if repo == nil {
		repo, err = client.CreateRepository(ctx, git.CreateRepositoryArgs{
			GitRepositoryToCreate: &git.GitRepositoryCreateOptions{
				Name: &name,
			},
			Project: &azProj,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s created successfully at %s\n", *repo.Name, *repo.WebUrl)
	} else {
		fmt.Printf("%s already exists in this Azure DevOps account\n", name)
	}

	// TODO: Figure out a better way to handle Azure auth than credential embedding
	remoteURL := *repo.RemoteUrl
	i := strings.Index(remoteURL, "@")
	remoteURL = remoteURL[:i] + ":" + azToken + remoteURL[i:]

	err = pushUpstream(name, "azure", remoteURL, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
