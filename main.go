package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/aniketalshi/go_rest_cache/config"
	"github.com/aniketalshi/go_rest_cache/app"
)

func main() {
	httpPort := flag.String("http_port", "3000", "Port to listen for HTTP traffic")
	
	// parse all cmdline flags
	flag.Parse()

	// initialize the config
	_, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	// initialize the application
	app := &app.App{}
	app.Initialize()
	app.Run()
	
	// launch the server 
	log.Fatal(http.ListenAndServe(":" + *httpPort, app.Handler))
}

