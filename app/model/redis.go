package model

import (
	"github.com/go-redis/redis"
	"github.com/aniketalshi/go_rest_cache/config"
)

// DBClient maintains a redis connection and a shim layer on top of redis library
type DBClient struct {
	client *redis.Client
}

// SetupDBCLient initializes client to talk to redis
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

// Set sets the key in redis with provided data
func (db *DBClient) Set (key string, data []byte) {
	db.client.Set(key, data, 0)
}

// Get retrieves the value corresponding to key in redis
func (db *DBClient) Get (key string) []byte {
	content, _ := db.client.Get(key).Bytes()
	return content
}
