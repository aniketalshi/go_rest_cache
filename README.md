## GO REST API Cache

This is a project to demonstrate building api cache in golang. Responses from upstream server is periodically cached in Redis. This allows us to build custom aggregated views on cached data and allow clients to query those views.

For this demo, we have built api cache on top of github apis for an organization. 

### Building and Running


Clone this repository and cd into go_rest_cache and run :

```bash
git clone https://github.com/aniketalshi/go_rest_cache.git
cd go_rest_cache

docker-compose build

GITHUB_API_TOKEN=<token> docker-compose up
```
pre-requisties : 
 - Docker Engine >= 1.13 
 - Docker Compose >= 1.11

Make sure you have GITHUB_API_TOKEN set in your environment which helps us overcome api rate limit.

To stop the containers, run
```bash
docker-compose down
```


### Testing

While the container is running, run this script 

Run 
```bash
test/api_suite.sh
```

### Configuration

`config/config.yaml` script contains all configuration parameters. It is loaded at runtime by the program although environment variables take precedence over values set in config.yaml.



### Architecture

As the server starts, it starts bunch of go routines. These go routines run periodically and fetch the response from upstream api.github.com - refresh interval can be set in configuration script.
Requests from users are examined if they are cached, if yes then we lookup in redis and serve those. For all non_cached requests, they are queried from upstream. We also build views to get aggregated response on top of get_repository api. Views are:

- /view/top/N/forks
- /view/top/N/last_updated
- /view/top/N/open_issues
-/view/top/N/stars

For computing these views, go routine fetch our cached response for repository from redis. Sort the repository struct according to respective parameter and cache it in its own key in redis. Thread which caches repository and thread which computes view communicate and achieve synchronization using channels.


### Next Steps

- Add unit-tests and test-coverage
- Observability is missing - integration with m3db, grafana and ELK to track metrics such as requests latencies, number of requests for each endpoint, http statuscodes, cpu/mem usage etc.
- Ratelimiting - will prevent us from storming redis instance with heavy load
- Redis - Master/Slave setup
