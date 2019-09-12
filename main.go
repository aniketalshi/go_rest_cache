package main

import (
	"flag"
	//"github.com/go-redis/redis"
	"log"
	//"go.uber.org/zap"
	"context"
	"github.com/aniketalshi/go_rest_cache/config"
	"github.com/aniketalshi/go_rest_cache/logging"
	"github.com/aniketalshi/go_rest_cache/db"
	"net/http"
	"fmt"
)



//func main() {
	// Create Redis Client
	//client := redis.NewClient(&redis.Options{
	//	Addr:     getEnv("REDIS_URL", "localhost:6379"),
	//	Password: getEnv("REDIS_PASSWORD", ""),
	//	DB:       0,
	//})

	//_, err := client.Ping().Result()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//cfg, err := config.InitConfig()
	//if err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Println("Port : ", cfg.GetServerPort())
	//
	//logging.InitLogger()
	//
	//logging.GetLogger().Info("This is info msg", zap.String("msg", "path"))


//}

//func getEnv(key, defaultValue string) string {
//	value := os.Getenv(key)
//	if value == "" {
//		return defaultValue
//	}
//	return value
//}

func main() {
	httpPort := flag.Uint("http_port", 3000, "Port to listen for HTTP traffic")
	
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

	cacher.cache_repos()
	cacher.cache_members()
	cacher.cache_org_details()
	cacher.cache_root_endpoint()
	cacher.populate_views()


	cacher.get_repos()
	log.Fatal(http.ListenAndServe(":3000", contexedHandler))
}

