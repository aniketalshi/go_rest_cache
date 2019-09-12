package main

import(
	//"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"net/http/httputil"
	"encoding/json"
	//"github.com/google/go-github/v28/github"
)

// Handlers is a container for maintaining reference to all handlers and also maintains reference to reverse proxy stub 
type Handlers struct 
{
	stub *httputil.ReverseProxy
	cacher *Cacher
}

// HandleCachedAPI handles the api responses for path which are pre-cached in redis
func (hh *Handlers) HandleCachedAPI(w http.ResponseWriter, r *http.Request) {
	//hh.stub.ServeHTTP(w, r)	
	
	if r.URL.Path == "/orgs/Netflix/repos" {

		response := hh.cacher.get_repos()
		json.NewEncoder(w).Encode(response)
	}
}

// HandleDefaults is the default http handler
func (hh *Handlers) HandleDefaults (w http.ResponseWriter, r *http.Request) {
	hh.stub.ServeHTTP(w, r)	
}

func SetupHandlers(cacher *Cacher) http.Handler{
	r := mux.NewRouter()

	proxy := &Handlers{
		stub: GenerateProxy(),
		cacher: cacher,	
	}

	cachedPaths := []string{"/",
			           "/orgs/Netflix",
			           "/orgs/Netflix/members",
			           "/orgs/Netflix/repos",
					 }

	for _, conf := range cachedPaths {
		r.HandleFunc(conf, proxy.HandleCachedAPI)
	}
	
	// fallback to default handler for all the rest of paths
	r.PathPrefix("/").HandlerFunc(proxy.HandleDefaults)
	
	return SetupInterceptor(r)	
}
