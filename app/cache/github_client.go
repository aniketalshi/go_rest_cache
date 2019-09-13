package cache

import (
	"golang.org/x/oauth2"
	"context"
	"net/http"
	"io/ioutil"
	"github.com/aniketalshi/go_rest_cache/app/logging"
	"github.com/aniketalshi/go_rest_cache/config"
	"github.com/google/go-github/v28/github"
)

// GithubClient maintains instance of golang github client class and associated context
type GithubClient struct
{
	Stub *github.Client	
	ctx	 context.Context
}

// GetNewGithubClient will setup access tokens and setup a new github client 
func GetNewGithubClient(ctx context.Context) *GithubClient {

	// GITHUB api token is required for overcoming ratelimit while querying the apis
	apiToken := config.GetConfig().GetTargetToken()

	var client *github.Client	

	// check if token is set
	if apiToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: apiToken},	
		)

		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {

		logging.Logger(ctx).Error("GITHUB API TOKEN is not set")
		client = github.NewClient(nil)
	}

	return &GithubClient{
		Stub: client,
		ctx: ctx,
	}
}

// GetRepositories calls the get repo for a given organization api on github and
// paginates through all  the response pages adding them to result set
func (gc *GithubClient) GetRepositories() ([]*github.Repository, error) {

    opt := &github.RepositoryListByOrgOptions{
    	ListOptions: github.ListOptions{PerPage: 10},
    }
    // get all pages of results
    var allRepos []*github.Repository
    for {
    	repos, resp, err := gc.Stub.Repositories.ListByOrg(gc.ctx, config.GetConfig().GetOrg(), opt)
    	if err != nil {
    		return nil, err
    	}
    	allRepos = append(allRepos, repos...)

		// check if this is the last page
    	if resp.NextPage == 0 {
    		break
    	}
    	opt.Page = resp.NextPage
    }	

	return allRepos, nil
}

func (gc *GithubClient) GetMembers() ([]*github.User, error) {
	
	opt := &github.ListMembersOptions {
		ListOptions: github.ListOptions{PerPage: 10},
	}

	var allMembers []*github.User
	for {
		users, resp, err := gc.Stub.Organizations.ListMembers(gc.ctx, config.GetConfig().GetOrg(), opt)
		if err != nil {
			return nil, err
		}
	
		allMembers = append(allMembers, users...)

		// check if this is the last page
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allMembers, nil
}

// GetOrgDetails query /org/{org} endpoint 
// We deliberately bypass github client here because it returns a repostiory struct
// that is different than what we observe by curling endpoint directly - some fields are skipped.
// That caused our unit-test to failed. So hitting endpoint directly.
func (gc *GithubClient) GetOrgDetails(url string) ([]byte, error) {
	return gc.queryUpstream(url)
}

// GetRootInfo queries the root endpoint 
// github client api does not method to query root endpoint. 
// So querying manually bypassing the client
func (gc *GithubClient) GetRootInfo() ([]byte, error) {
	return gc.queryUpstream("")
}

// queryUpstream is internal function which is used for querying endpoint directly 
// bypassing github client library
func (gc *GithubClient) queryUpstream (path string) ([]byte, error) {

	token := config.GetConfig().GetTargetToken()
	
	url := config.GetConfig().GetTargetScheme() + "://" + config.GetConfig().GetTargetUrl()

	if path != "" {
		url = url + path
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// API Token to overcome ratelimit	
	req.Header.Set("Authorization", "token " + token)
	client := &http.Client{}

	// issue the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// this is crucial to make sure resp.Body is closed after 
	// we open it for reading it into buffer
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
