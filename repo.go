// File contains functions related to handling repository
// clone, push, pull, and remote-setting events.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
)

// Recursively scan a file and create a map of the contents.
// The file should contain one public git URL per line.
func parseRepoList(list string) (map[string]string, error) {
	file, err := os.Open(list)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var repos = make(map[string]string)
	scanner := bufio.NewScanner(file)
	// Go through the file line by line.
	for scanner.Scan() {
		repo := scanner.Text()
		// For the key/value in the map, set the key
		// to the base of the repo (usually the repo
		// name) and the value to the URL.
		repos[path.Base(repo)] = scanner.Text()
	}
	return repos, scanner.Err()
}

// Clone or pull repositories defined in the passed in map.
func cloneRepos(repos map[string]string) error {
	// Build the local repository path base used for output.
	pathBase := path.Clean(viper.Get("path").(string))
	// Loop through the repository map.
	for name, url := range repos {
		fmt.Printf("Attempting to clone repository: %s\n", name)
		// Create the full output path for this repo.
		out := filepath.Join(pathBase, name)
		// Clone the repo to the output path.
		_, err := git.PlainClone(out, false, &git.CloneOptions{
			URL:        url,
			Progress:   os.Stdout,
			RemoteName: "snyk",
		})
		// Check if the repo was already cloned. If it was, attempt
		// to fetch and pull it instead of clone.
		if err != nil && err != git.ErrRepositoryAlreadyExists {
			return err
		} else if err == git.ErrRepositoryAlreadyExists {
			fmt.Printf("%s already cloned, attempting to pull from upstream\n", name)
			// Open previously cloned, local repo.
			r, err := git.PlainOpen(out)
			if err != nil {
				return err
			}

			// Get the working directory for the repository
			w, err := r.Worktree()
			if err != nil {
				return err
			}

			// Perform a fetch on upstream "snyk" remote.
			err = r.Fetch(&git.FetchOptions{
				RemoteName: "snyk",
				RefSpecs: []config.RefSpec{
					"refs/*:refs/*",
				},
				Progress: os.Stdout,
			})

			// Pull the latest changes from the origin remote and merge into the current branch.
			err = w.Pull(&git.PullOptions{RemoteName: "snyk"})
			if err != nil && err != git.NoErrAlreadyUpToDate && err != git.ErrNonFastForwardUpdate {
				return err
			} else if err == git.NoErrAlreadyUpToDate {
				fmt.Printf("%s already up to date, nothing to pull\n", name)
			} else if err == git.ErrNonFastForwardUpdate {
				fmt.Printf("%s has been modified locally, cannot merge from upstream", name)
			}
		}
	}
	return nil
}

// Control function to call each individual SCM creation
// function located in their respective files.
func createRemoteRepos(repos map[string]string) error {
	// Loop through repo list and call function to create
	// remote repository for each SCM.
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

// Create local git remote entries and push upstream.
func pushUpstream(name string, remote string, giturl string, u interface{}, p interface{}) error {
	// Build the local repository path base used for input.
	pathBase := path.Clean(viper.Get("path").(string))
	in := path.Join(pathBase, name)

	fmt.Printf("Pushing latest %s to remote \"%s\"\n", name, remote)
	// Open local git repository.
	r, err := git.PlainOpen(in)
	if err != nil {
		return err
	}

	// Create a named remote (github, gitlab, etc.)
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: remote,
		URLs: []string{giturl},
	})
	// Check if the remote was already present. If it was
	// remove it and recreate it to ensure integrity.
	if err != nil && !strings.Contains(err.Error(), "remote already exists") {
		return err
	} else if err != nil && strings.Contains(err.Error(), "remote already exists") {
		err = r.DeleteRemote(remote)
		if err != nil {
			return err
		}
		_, err = r.CreateRemote(&config.RemoteConfig{
			Name: remote,
			URLs: []string{giturl},
		})
		if err != nil {
			return err
		}
	}

	// Setup push options to be used in the actual push.
	pushOptions := git.PushOptions{
		RemoteName: remote,
		Progress:   os.Stdout,
		Force:      true,
	}

	// If credentials are necessary for the SCM push, add them
	// to the pushOptions variable.
	if u != nil && p != nil {
		pushOptions.Auth = &http.BasicAuth{
			Username: u.(string),
			Password: p.(string),
		}
	}

	// Push the code to the upstream SCM.
	err = r.Push(&pushOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	fmt.Printf("%s on remote \"%s\" up to date\n", name, remote)

	return nil
}
