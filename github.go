package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func createGitHubRepo(name string) error {

	ghUser := viper.Get("ghUser").(string)
	ghOrg := viper.Get("ghOrg").(string)
	ghToken := viper.Get("ghKey").(string)
	private := false
	description := ""

	if ghOrg == "" {
		fmt.Printf("Attempting to create %s in %s's GitHub\n", name, ghUser)
	} else {
		fmt.Printf("Attempting to create %s in the %s GitHub Org\n", name, ghOrg)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repo, resp, err := client.Repositories.Create(ctx, ghOrg, &github.Repository{
		Name:        &name,
		Private:     &private,
		Description: &description,
	})
	// TODO: add proper error message checking for existence
	if err != nil && resp.StatusCode != 422 {
		return err
	} else if resp.StatusCode == 422 {
		if ghOrg == "" {
			fmt.Printf("%s already exists in this GitHub account\n", name)
			repo, resp, err = client.Repositories.Get(ctx, ghUser, name)
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("%s already exists in this GitHub Org\n", name)
			repo, resp, err = client.Repositories.Get(ctx, ghOrg, name)
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("%s created successfully at %s\n", repo.GetName(), repo.GetURL())
	}

	err = pushUpstream(name, "github", *repo.CloneURL)
	if err != nil {
		return err
	}

	return nil
}
