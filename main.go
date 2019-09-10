package main

import (
	"fmt"
	//"github.com/go-redis/redis"
	"log"
	"go.uber.org/zap"
	"github.com/aniketalshi/go_rest_cache/config"
	"github.com/aniketalshi/go_rest_cache/logging"
)

func main() {
	fmt.Println("Hello World!")
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
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Port : ", cfg.GetServerPort())
	
	logging.InitLogger()
	
	logging.GetLogger().Info("This is info msg", zap.String("msg", "path"))


}

//func getEnv(key, defaultValue string) string {
//	value := os.Getenv(key)
//	if value == "" {
//		return defaultValue
//	}
//	return value
//}
