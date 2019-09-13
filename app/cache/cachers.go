package cache

import (
	"time"
	"encoding/json"
	"sort"
	"context"
	"strconv"

	"go.uber.org/zap"

	"github.com/aniketalshi/go_rest_cache/app/model"
	"github.com/aniketalshi/go_rest_cache/config"
	"github.com/google/go-github/v28/github"
	"github.com/aniketalshi/go_rest_cache/app/logging"
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
func (cc *Cacher) CacheRepos(isCached chan<- bool, url string) {

	cc.schedule(func() {
	        repos, err := cc.GitClient.GetRepositories()
	        if err != nil {
				logging.Logger(context.Background()).Fatal("Error getting the repositories",
					      									zap.String("msg", err.Error()))
	        }

	        js, err := json.Marshal(repos)
	        if err != nil {
				logging.Logger(context.Background()).Fatal("Error trying to marshal repository struct",
															zap.String("msg", err.Error()))
	        }
	    	
	        cc.DBClient.Set(url, js)
	
			// Nofity the go routine populating views that we have cached new repository data into redis
			isCached <- true
	})
}

func (cc *Cacher) SortAndSetView(repos []github.Repository, key string, comparator func(int, int)bool) {

	// sort repos by comparator
	sort.Slice(repos, comparator)

	js, err := json.Marshal(repos)
	if err != nil {
		logging.Logger(context.Background()).Fatal("Error trying to marshal repository struct",
												   zap.String("msg", err.Error()))
	}

	cc.DBClient.Set(key, js)
}

func (cc *Cacher) PopulateViews(isCached <-chan bool, url string) {
	
	cc.schedule(func() {
	
		// waiting on signal from go routine which has cached new repository data into redis
		<-isCached
	
		resp := cc.GetCachedEndpoint(url)
	    var repos []github.Repository

	    if err := json.Unmarshal(resp, &repos); err != nil {
	    	logging.Logger(context.Background()).Error("Error unmarshalling repository struct",
	    							  zap.String("msg", err.Error()))
        }	

	    
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

func (cc *Cacher) GetView(ctx context.Context, key string, limit int) ([]ViewResult, error) {

	serializedRepos := cc.DBClient.Get(key)
	
	var repos []github.Repository
	if err := json.Unmarshal(serializedRepos, &repos); err != nil {

		logging.Logger(ctx).Error("Error unmarshalling repository struct",
								  zap.String("msg", err.Error()))
		return nil, err
    }	

	logging.Logger(ctx).Info("Repositories fetched for view",
							zap.Int("Num Repos", len(repos)))
	
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
	return result, nil
}

// CacheMembers caches data related to member of org into redis
func (cc *Cacher) CacheMembers(url string) {

	cc.schedule(func() {
	    users, err := cc.GitClient.GetMembers()

	    if err != nil {
	    	logging.Logger(context.Background()).Error("Error fetching members of org",
											         zap.String("msg", err.Error()))
	    } else {

	        js, err := json.Marshal(users)
	        if err != nil {
	        	logging.Logger(context.Background()).Fatal("Error trying to marshal users struct",
		    									         zap.String("msg", err.Error()))
	        }
	        
	        cc.DBClient.Set(url, js)
		}
	})
}


// CacheOrgDetails caches data from org endpoint into redis
func (cc *Cacher) CacheOrgDetails(url string) {

	cc.schedule(func() {
		orgInfo, err := cc.GitClient.GetOrgDetails(url)
	    if err != nil {
	    	logging.Logger(context.Background()).Fatal("Error getting the org info",
											         zap.String("msg", err.Error()))
	    }

	    cc.DBClient.Set(url, orgInfo)
	})
}

// CacheRootEndpoint caches the info from root endpoint into redis
func (cc *Cacher) CacheRootEndpoint(url string) {
	
	cc.schedule (func() {
	    resp, err := cc.GitClient.GetRootInfo()
	    if err != nil {
	    	logging.Logger(context.Background()).Fatal("Error getting the root node",
											         zap.String("msg", err.Error()))
	    }
	    cc.DBClient.Set(url, resp)
	})
}

// GetCachedEndpoint fetches the data from redis and serves response back to handler
func (cc *Cacher) GetCachedEndpoint(path string) []byte {
	return cc.DBClient.Get(path)
}

