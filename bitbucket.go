package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ktrysmt/go-bitbucket"

	"github.com/spf13/viper"
)

func createBitBucketRepo(name string) error {
	bbUser := viper.Get("bbUser").(string)
	bbToken := viper.Get("bbKey").(string)

	fmt.Printf("Attempting to create %s in %s's Bitbucket\n", name, bbUser)

	client := bitbucket.NewBasicAuth(bbUser, bbToken)

	repo, err := client.Repositories.Repository.Get(&bitbucket.RepositoryOptions{
		Owner:    bbUser,
		RepoSlug: name,
	})
	if err != nil && !strings.Contains(err.Error(), "404 Not Found") {
		return err
	}
	if repo == nil {
		repo, err = client.Repositories.Repository.Create(&bitbucket.RepositoryOptions{
			Owner:    bbUser,
			RepoSlug: name,
		})
		if err != nil {
			return err
		}
		var bbURL = repo.Links["html"].(map[string]interface{})["href"].(string)
		fmt.Printf("%s created successfully at %s\n", repo.Slug, bbURL)
	} else {
		fmt.Printf("%s already exists in this Bitbucket account\n", name)
	}

	var httpURL string
	for _, l := range repo.Links["clone"].([]interface{}) {
		for _, v := range l.(map[string]interface{}) {
			if strings.Contains(v.(string), "https://") && strings.Contains(v.(string), ".git") {
				httpURL = v.(string)
			}
		}
	}
	if httpURL == "" {
		return errors.New("unable to find http git link from Bitbucket repository")
	}

	err = pushUpstream(name, "bitbucket", httpURL, bbUser, bbToken)
	if err != nil {
		return err
	}

	return nil
}
