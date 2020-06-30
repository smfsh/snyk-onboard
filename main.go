package main

import (
	"log"
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
