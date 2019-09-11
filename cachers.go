package main

import (
	"os"
	"github.com/aniketalshi/go_rest_cache/db"
	"encoding/json"
	"fmt"
)

type Cacher struct {
	gitClient *GithubClient
	dbClient *db.DBClient
}

// Queries the github api to fetch all the repos for a given organization and caches 
// the response into the redis
func (cc *Cacher) cache_repos() {
	
	repos, err := cc.gitClient.GetRepositories()
	if err != nil {
		fmt.Println("Error getting the repositories", err.Error())
		os.Exit(1)
	}

	js, err := json.Marshal(repos)
	if err != nil {
		fmt.Println("Error trying to marshal repository struct", err.Error())
		os.Exit(1)
	}
	
	cc.dbClient.Set("/orgs/Netflix/repos", js)
}

func (cc *Cacher) cache_members() {

	users, err := cc.gitClient.GetMembers()
	if err != nil {
		fmt.Println("Error getting the users", err.Error())
		os.Exit(1)
	}

	js, err := json.Marshal(users)
	if err != nil {
		fmt.Println("Error trying to marshal users struct", err.Error())
		os.Exit(1)
	}
	
	cc.dbClient.Set("/orgs/Netflix/members", js)
}
