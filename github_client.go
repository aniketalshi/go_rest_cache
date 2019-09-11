package main

import(
	"os"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"context"
	"github.com/aniketalshi/go_rest_cache/logging"
)

type GithubClient struct
{
	Stub *github.Client	
	ctx context.Context
}

func GetNewGithubClient(ctx context.Context) *GithubClient {

	// GITHUB api token is required for overcoming ratelimit while querying the apis
	apiToken := os.Getenv("GITHUB_API_TOKEN")
	
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
    	repos, resp, err := gc.Stub.Repositories.ListByOrg(gc.ctx, "Netflix", opt)
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
		users, resp, err := gc.Stub.Organizations.ListMembers(gc.ctx, "Netflix", opt)
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
