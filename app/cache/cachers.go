package cache

import (
	"os"
	"time"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/aniketalshi/go_rest_cache/app/model"
	"github.com/aniketalshi/go_rest_cache/config"
	"github.com/google/go-github/v28/github"
)

// Cacher is responsible for maintaining go routines which peridoically cache data in redis 
// and apis to get data back from redis
type Cacher struct {
	GitClient *GithubClient
	DBClient *model.DBClient
}

// ViewResult is a structure for extracting data into custom views we serve to clients
// for viewing repository by top N parameters
type ViewResult struct {
	Repo string	 `json:"repo"`
	Count string `json:"count"`
}

// schedules the go routine to run at periodic intervals
func (cc *Cacher) schedule(cachingFunc func()) {
	
	// Get the refresh rate from config
	refreshDuration := time.Duration(config.GetConfig().GetCacheConfig().RefreshInterval)
	
	// ticker goes of at fixed intervals
	ticker := time.NewTicker(refreshDuration * time.Second)
	defer ticker.Stop()
	
	for ; true; <- ticker.C{
		cachingFunc()
	}
}

// Queries the github api to fetch all the repos for a given organization and caches 
// the response into the redis
func (cc *Cacher) CacheRepos(isCached chan<- bool) {

	cc.schedule(func() {
	        repos, err := cc.GitClient.GetRepositories()
	        if err != nil {
	        	fmt.Println("Error getting the repositories", err.Error())
	        	os.Exit(1)
	        }

	        js, err := json.Marshal(repos)
	        if err != nil {
	        	fmt.Println("Error trying to marshal repository struct", err.Error())
	        	os.Exit(1)
	        }
	    	
	        cc.DBClient.Set("/orgs/Netflix/repos", js)
	
			// Nofity the go routine populating views that we have cached new repository data into redis
			isCached <- true
	})
}

func (cc *Cacher) SortAndSetView(repos []github.Repository, key string, comparator func(int, int)bool) {

	// sort repos by comparator
	sort.Slice(repos, comparator)

	js, err := json.Marshal(repos)
	if err != nil {
		fmt.Println("Error trying to marshal repository struct", err.Error())
		os.Exit(1)
	}

	cc.DBClient.Set(key, js)
}

func (cc *Cacher) PopulateViews(isCached <-chan bool) {
	
	cc.schedule(func() {
	
		// waiting on signal from go routine which has cached new repository data into redis
		<-isCached
	
	    repos := cc.GetRepos()
	    
	    // sort repo by forks and insert the sorted list in redis
	    cc.SortAndSetView(repos, "top-repo-by-forks", func(i, j int) bool {
	    	return *(repos[i].ForksCount) > *(repos[j].ForksCount)
	    })

	    // sort by last updated 	
	    cc.SortAndSetView(repos, "top-repo-by-lastupdated", func(i, j int) bool {
	    	return repos[i].UpdatedAt.Time.Sub(repos[j].UpdatedAt.Time) > 0
	    })
	    
	    // sort by number of open issues
	    cc.SortAndSetView(repos, "top-repo-by-openissues", func(i, j int) bool {
	    	return *(repos[i].OpenIssuesCount) > *(repos[j].OpenIssuesCount)
	    })

	    // sort by number of stars
	    cc.SortAndSetView(repos, "top-repo-by-stars", func(i, j int) bool {
	    	return *(repos[i].StargazersCount) > *(repos[j].StargazersCount)
	    })
	})
}

func (cc *Cacher) GetView(key string, limit int) []ViewResult {

	serializedRepos := cc.DBClient.Get(key)
	
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
			count = repo.UpdatedAt.Time.Format(time.RFC3339)
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

func (cc *Cacher) CacheMembers() {

	cc.schedule(func() {
	    users, err := cc.GitClient.GetMembers()
	    if err != nil {
	    	fmt.Println("Error getting the users", err.Error())
	    	os.Exit(1)
	    }

	    js, err := json.Marshal(users)
	    if err != nil {
	    	fmt.Println("Error trying to marshal users struct", err.Error())
	    	os.Exit(1)
	    }
	    
	    cc.DBClient.Set("/orgs/Netflix/members", js)
	})
}

func (cc *Cacher) CacheOrgDetails() {

	cc.schedule(func() {
		orgInfo, err := cc.GitClient.GetOrgDetails()
	    if err != nil {
	    	fmt.Println("Error getting the org info", err.Error())
	    	os.Exit(1)
	    }

	    cc.DBClient.Set("/orgs/Netflix", orgInfo)
	})
}

func (cc *Cacher) CacheRootEndpoint() {
	
	cc.schedule (func() {
	    resp, err := cc.GitClient.GetRootInfo()
	    if err != nil {
	    	fmt.Println("Error getting the root node", err.Error())
	    	os.Exit(1)
	    }
	    cc.DBClient.Set("/", resp)
	})
}

func (cc *Cacher) GetRepos() []github.Repository {
	
	serializedRepos := cc.DBClient.Get("/orgs/Netflix/repos")
	
	var repos []github.Repository
	if err := json.Unmarshal(serializedRepos, &repos); err != nil {
		fmt.Println("Error unmarshalling repository struct", err.Error())
    }	
	
	return repos
}

func (cc *Cacher) GetMembers() []github.User {
	
	serializedMembers := cc.DBClient.Get("/orgs/Netflix/members")

	var members []github.User
	if err := json.Unmarshal(serializedMembers, &members); err != nil {
		fmt.Println("Error unmarshalling member struct", err.Error())
    }	
	
	return members
}

func (cc *Cacher) GetOrg() []byte {
	
	serializedOrgInfo := cc.DBClient.Get("/orgs/Netflix")

	return serializedOrgInfo
}

func (cc *Cacher) GetRootEndpoint() []byte {
	serializedRootInfo := cc.DBClient.Get("/")
	return serializedRootInfo
}
