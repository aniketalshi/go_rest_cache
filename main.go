package main

import (
	"flag"
	//"github.com/go-redis/redis"
	"log"
	//"go.uber.org/zap"
	"time"
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
	
	done := make(chan bool)	

	go cacher.cache_repos(done)
	go cacher.cache_members(done)
	go cacher.cache_org_details(done)
	go cacher.cache_root_endpoint(done)
	
	time.Sleep(10 * time.Second)
	go cacher.populate_views(done)

	err = http.ListenAndServe(":" + *httpPort, contexedHandler)
	
	if err != nil {
		// close all goroutines gracefully
		done <- true
		log.Fatal(err)
	}
}

