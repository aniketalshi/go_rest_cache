package app

import (
	"context"
	"net/http"
	"log"

	"github.com/aniketalshi/go_rest_cache/app/logging"
	"github.com/aniketalshi/go_rest_cache/app/cache"
	"github.com/aniketalshi/go_rest_cache/app/model"
)

// App is high level server struct that acts as entrypoint
type App struct {
	DBClient *model.DBClient
	Cacher   *cache.Cacher
	Handler  http.Handler
}

// Initialize initializes all high level datastructures
func (aa *App) Initialize() {

	// initialize the logger
	logging.InitLogger()

	// setup client to talk to redis instance
	aa.DBClient = model.SetupDBClient()
	
	// setup cacher which maintains go routines to periodically cache data
	aa.Cacher = &cache.Cacher {
		GitClient: cache.GetNewGithubClient(context.Background()),
		DBClient: aa.DBClient,
	}
	
	// set up the mux router and handlers
	aa.Handler = cache.SetupHandlers(aa.Cacher)

	log.Print("Server Initialized. Starting up...")
}

// Run runs the go routines which will start caching the data periodically
func (aa *App) Run() {

	// channel used for synchronizing two different go routines so that one can 
	// let the other know when data is cached in redis.
	isCached := make(chan bool)
	
	go aa.Cacher.CacheRepos(isCached)
	go aa.Cacher.CacheMembers()
	go aa.Cacher.CacheOrgDetails()
	go aa.Cacher.CacheRootEndpoint()
	
	go aa.Cacher.PopulateViews(isCached)
}
	
