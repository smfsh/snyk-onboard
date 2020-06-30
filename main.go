package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/go-git/go-git/v5/config"

	"github.com/spf13/viper"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

func main() {
	repos, err := parseRepoList("repolist.txt")
	if err != nil {
		log.Panic(err)
	}
	err = cloneRepos(repos)
	if err != nil {
		log.Panic(err)
	}
	err = createRemoteRepos(repos)
	if err != nil {
		log.Panic(err)
	}
}

func parseRepoList(list string) (map[string]string, error) {
	file, err := os.Open(list)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var repos = make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repo := scanner.Text()
		repos[path.Base(repo)] = scanner.Text()
	}
	return repos, scanner.Err()
}

func cloneRepos(repos map[string]string) error {
	pathBase := path.Clean(viper.Get("path").(string))
	for name, url := range repos {
		fmt.Printf("Attempting to clone repository: %s\n", name)
		out := path.Join(pathBase, name)
		_, err := git.PlainClone(out, false, &git.CloneOptions{
			URL:        url,
			Progress:   os.Stdout,
			RemoteName: "snyk",
		})
		if err != nil && err != git.ErrRepositoryAlreadyExists {
			return err
		} else if err == git.ErrRepositoryAlreadyExists {
			fmt.Printf("%s already cloned, attempting to pull from upstream\n", name)
			r, err := git.PlainOpen("repos/" + name)
			if err != nil {
				return err
			}

			// Get the working directory for the repository
			w, err := r.Worktree()
			if err != nil {
				return err
			}

			// Pull the latest changes from the origin remote and merge into the current branch
			err = w.Pull(&git.PullOptions{RemoteName: "snyk"})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return err
			} else if err == git.NoErrAlreadyUpToDate {
				fmt.Printf("%s already up to date, nothing to pull\n", name)
			}
		}
	}
	return nil
}

func pushUpstream(name string, remote string, giturl string) error {
	pathBase := path.Clean(viper.Get("path").(string))
	in := path.Join(pathBase, name)

	switch remote {
	case "github":
		fmt.Printf("Pushing latest %s to remote \"%s\"\n", name, remote)
		r, err := git.PlainOpen(in)
		if err != nil {
			return err
		}

		r.CreateRemote(&config.RemoteConfig{
			Name: remote,
			URLs: []string{giturl},
		})

		err = r.Push(&git.PushOptions{
			RemoteName: remote,
		})

	default:
		return errors.New("pushUpstream: unknown upstream SCM")
	}

	return nil
}

func createRemoteRepos(repos map[string]string) error {
	for name := range repos {
		err := createGitHubRepo(name)
		if err != nil {
			return err
		}
		createGitLabRepo()
		createBitBucketRepo()
		createAzureRepo()
	}
	return nil
}

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

func createGitLabRepo() {

}

func createBitBucketRepo() {

}

func createAzureRepo() {

}
