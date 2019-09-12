package main

import (
	"os"
	"github.com/aniketalshi/go_rest_cache/db"
	"github.com/google/go-github/v28/github"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type Cacher struct {
	gitClient *GithubClient
	dbClient *db.DBClient
}

type ViewResult struct {
	Repo string	 `json:"repo"`
	Count string `json:"count"`
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

func (cc *Cacher) sort_and_insert_view(repos []github.Repository, key string, comparator func(int, int)bool) {

	// sort repos by comparator
	sort.Slice(repos, comparator)

	js, err := json.Marshal(repos)
	if err != nil {
		fmt.Println("Error trying to marshal repository struct", err.Error())
		os.Exit(1)
	}

	cc.dbClient.Set(key, js)
}

func (cc *Cacher) populate_views() {
	
	repos := cc.get_repos()
	
	// sort repo by forks and insert the sorted list in redis
	cc.sort_and_insert_view(repos, "top-repo-by-forks", func(i, j int) bool {
		return *(repos[i].ForksCount) > *(repos[j].ForksCount)
	})

	// sort by last updated 	
	cc.sort_and_insert_view(repos, "top-repo-by-lastupdated", func(i, j int) bool {
		return repos[i].UpdatedAt.Time.Sub(repos[j].UpdatedAt.Time) > 0
	})
	
	// sort by number of open issues
	cc.sort_and_insert_view(repos, "top-repo-by-openissues", func(i, j int) bool {
		return *(repos[i].OpenIssuesCount) > *(repos[j].OpenIssuesCount)
	})

	// sort by number of stars
	cc.sort_and_insert_view(repos, "top-repo-by-stars", func(i, j int) bool {
		return *(repos[i].StargazersCount) > *(repos[j].StargazersCount)
	})
}

func (cc *Cacher) get_view(key string, limit int) []ViewResult {

	serializedRepos := cc.dbClient.Get(key)
	
	var repos []github.Repository
	if err := json.Unmarshal(serializedRepos, &repos); err != nil {
		fmt.Println("Error unmarshalling repository struct", err.Error())
    }	
	
	var result []ViewResult
	counter := 1

	for _, repo := range repos {
		
		var count string
		if key == "top-repo-by-forks" {
			count = strconv.Itoa(*repo.ForksCount)
		} else if key == "top-repo-by-lastupdated" {
			count = repo.UpdatedAt.Time.String()
		} else if key == "top-repo-by-openissues" {
			count = strconv.Itoa(*repo.OpenIssuesCount)
		} else if key == "top-repo-by-stars" {
			count = strconv.Itoa(*repo.StargazersCount)
		}

		res := ViewResult{
			Repo: *repo.FullName,
			Count: count,
		}
		
		result = append(result, res)

		if counter == limit {
			break
		}
		counter += 1
	}

	return result
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
