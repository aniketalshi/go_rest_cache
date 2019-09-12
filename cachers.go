package main

import (
	"os"
	"github.com/aniketalshi/go_rest_cache/db"
	"github.com/google/go-github/v28/github"
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

func (cc *Cacher) cache_org_details() {

	orgInfo, err := cc.gitClient.GetOrgDetails()
	if err != nil {
		fmt.Println("Error getting the org info", err.Error())
		os.Exit(1)
	}

	js, err := json.Marshal(orgInfo)
	if err != nil {
		fmt.Println("Error trying to marshal organization struct", err.Error())
		os.Exit(1)
	}
	
	cc.dbClient.Set("/orgs/Netflix", js)
}

func (cc *Cacher) cache_root_endpoint() {
	
	resp, err := cc.gitClient.GetRootInfo()
	if err != nil {
		fmt.Println("Error getting the root node", err.Error())
		os.Exit(1)
	}
	cc.dbClient.Set("/", resp)
}

func (cc *Cacher) get_repos() []github.Repository {
	
	serializedRepos := cc.dbClient.Get("/orgs/Netflix/repos")
	
	var repos []github.Repository
	if err := json.Unmarshal(serializedRepos, &repos); err != nil {
		fmt.Println("Error unmarshalling repository struct", err.Error())
    }	
	
	return repos
	//for _, repo := range repos {
	//	fmt.Println(*repo.FullName)
	//}
	//return serializedRepos
}

func (cc *Cacher) get_members() []github.User {
	
	serializedMembers := cc.dbClient.Get("/orgs/Netflix/members")

	var members []github.User
	if err := json.Unmarshal(serializedMembers, &members); err != nil {
		fmt.Println("Error unmarshalling repository struct", err.Error())
    }	
	
	return members
}

func (cc *Cacher) get_org() github.Organization {
	
	serializedOrgInfo := cc.dbClient.Get("/orgs/Netflix")

	var orgInfo github.Organization
	if err := json.Unmarshal(serializedOrgInfo, &orgInfo); err != nil {
		fmt.Println("Error unmarshalling org struct", err.Error())	
	}

	return orgInfo
}

func (cc *Cacher) get_root_endpoint() []byte {
	serializedRootInfo := cc.dbClient.Get("/")
	return serializedRootInfo
}
