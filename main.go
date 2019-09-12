package main

import (
	"flag"
	//"github.com/go-redis/redis"
	"log"
	//"go.uber.org/zap"
	//"time"
	//"strconv"
	"context"
	"github.com/aniketalshi/go_rest_cache/config"
	"github.com/aniketalshi/go_rest_cache/logging"
	"github.com/aniketalshi/go_rest_cache/db"
	"net/http"
	"fmt"
)

func main() {
	httpPort := flag.String("http_port", "3000", "Port to listen for HTTP traffic")
	
	// parse all cmdline flags
	flag.Parse()

	fmt.Println("Http port :", *httpPort)
	
	// initialize the config
	_, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	
	// initialize the logger
	logging.InitLogger()
	
	cacher := &Cacher {
		gitClient: GetNewGithubClient(context.Background()),
		dbClient: db.SetupDBClient(),
	}

	contexedHandler := SetupHandlers(cacher)
	
	// used for synchronizing two different go routines so that one can 
	// let the other know when data is cached
	isCached := make(chan bool)
	
	go cacher.cache_repos(isCached)
	go cacher.cache_members()
	go cacher.cache_org_details()
	go cacher.cache_root_endpoint()
	
	//time.Sleep(10 * time.Second)
	go cacher.populate_views(isCached)

	log.Fatal(http.ListenAndServe(":" + *httpPort, contexedHandler))
	
}

