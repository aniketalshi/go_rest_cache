package db

import (
	"github.com/go-redis/redis"
	"github.com/aniketalshi/go_rest_cache/config"
)

type DBClient struct {
	client *redis.Client
}

// initializes client for redis 
func SetupDBClient() *DBClient {
	
	db := &DBClient{
		client: redis.NewClient(&redis.Options{
		    Addr: config.GetConfig().GetRedisURL(),
		    Password: "",
		    DB:       0,
		}),
	}

	return db
}

func (db *DBClient) Set (key string, data []byte) {
	db.client.Set(key, data, 0)
}

func (db *DBClient) Get (key string) []byte {
	content, _ := db.client.Get(key).Bytes()
	return content
}
