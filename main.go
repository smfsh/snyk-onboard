// Clone down and replicate goof-type repositories to create a
// valid demo ecosystem for Snyk. This application will recursively
// download repositories defined in repolist.txt and automatically
// create personal clones of each repository in each of Snyk's
// supported upstream SCMs: GitHub, GitLab, Bitbucket, and Azure.
// This application will also keep each repository up to date and
// synced with the upstream original.

package main

import (
	"log"
)

// Main function to execute the processes sub-processes of the app.
func main() {
	// Create a map of the repository name and location to be utilized
	// by the rest of the app functions.
	repos, err := parseRepoList("repolist.txt")
	if err != nil {
		log.Panic(err)
	}
	// Clone or pull original remote repositories.
	err = cloneRepos(repos)
	if err != nil {
		log.Panic(err)
	}
	// Create or update personal remote repositories.
	err = createRemoteRepos(repos)
	if err != nil {
		log.Panic(err)
	}
}
