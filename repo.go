package main

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/spf13/viper"
)

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

func createRemoteRepos(repos map[string]string) error {
	for name := range repos {
		err := createGitHubRepo(name)
		if err != nil {
			return err
		}
		err = createGitLabRepo(name)
		if err != nil {
			return err
		}
		err = createBitBucketRepo(name)
		if err != nil {
			return err
		}
		err = createAzureRepo(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func pushUpstream(name string, remote string, giturl string, u string, p string) error {
	pathBase := path.Clean(viper.Get("path").(string))
	in := path.Join(pathBase, name)

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
		Progress:   os.Stdout,
		Auth: &http.BasicAuth{
			Username: u,
			Password: p,
		},
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	fmt.Printf("%s on remote \"%s\" up to date\n", name, remote)

	return nil
}
