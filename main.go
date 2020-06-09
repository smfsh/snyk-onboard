package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path"

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
	for name, url := range repos {
		fmt.Printf("Attempting to clone repository: %s\n", name)
		_, err := git.PlainClone("repos/"+name, false, &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout,
		})
		if err != nil && err != git.ErrRepositoryAlreadyExists {
			return err
		} else if err == git.ErrRepositoryAlreadyExists {
			fmt.Printf("%s already cloned, attempting to pull from origin\n", name)
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
			err = w.Pull(&git.PullOptions{RemoteName: "origin"})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return err
			} else if err == git.NoErrAlreadyUpToDate {
				fmt.Printf("%s already up to date, nothing to pull\n", name)
			}
		}
	}
	return nil
}

func pushUpstream(name string, origin string) error {
	// TODO: add authenticated push code
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
	fmt.Printf("Attempting to create %s in GitHub\n", name)

	gh_org := ""
	gh_token := ""
	private := false
	description := ""

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh_token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repo, resp, err := client.Repositories.Create(ctx, gh_org, &github.Repository{
		Name:        &name,
		Private:     &private,
		Description: &description,
	})
	// TODO: add proper error message checking for existence
	if err != nil && resp.StatusCode != 422 {
		return err
	} else if resp.StatusCode == 422 {
		fmt.Printf("%s already exists in this GitHub account or org\n", name)
		err = pushUpstream(name, "github")
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("%s created successfully at %s\n", repo.GetName(), repo.GetURL())
	}

	return nil
}

func createGitLabRepo() {

}

func createBitBucketRepo() {

}

func createAzureRepo() {

}
