package cache

import (
	"fmt"
	"net/http"
	"encoding/json"
	"strconv"
	"strings"
	"net/http/httputil"

	"github.com/gorilla/mux"
)

// Handlers is a container for maintaining reference to all handlers and also maintains reference to reverse proxy stub 
type Handlers struct 
{
	stub *httputil.ReverseProxy
	cacher *Cacher
}

// HandleCachedAPI handles the api responses for path which are pre-cached in redis
func (hh *Handlers) HandleCachedAPI(w http.ResponseWriter, r *http.Request) {
	
	if r.URL.Path == "/orgs/Netflix/repos" {
		response := hh.cacher.GetRepos()
		json.NewEncoder(w).Encode(response)

	} else if r.URL.Path == "/orgs/Netflix/members" {
		response := hh.cacher.GetMembers()
		json.NewEncoder(w).Encode(response)

	} else if r.URL.Path == "/orgs/Netflix" {
		response := hh.cacher.GetOrg()
		//json.NewEncoder(w).Encode(response)
		w.WriteHeader(200)
		w.Write(response)

	} else if r.URL.Path == "/" {
		response := hh.cacher.GetRootEndpoint()
		w.WriteHeader(200)
		w.Write(response)

	} else {
		fmt.Println("The requested path is not supposed to be cached")
		hh.stub.ServeHTTP(w, r)	
	}
}

// HandleDefaults is the default http handler
func (hh *Handlers) HandleDefaults (w http.ResponseWriter, r *http.Request) {
	hh.stub.ServeHTTP(w, r)	
}

func (hh *Handlers) GetTopForkedRepos (w http.ResponseWriter, r *http.Request) {
	hh.HandleViews(w, r, "top-repo-by-forks")
}

func (hh *Handlers) GetLastUpdatedRepos (w http.ResponseWriter, r *http.Request) {
	hh.HandleViews(w, r, "top-repo-by-lastupdated")
}

func (hh *Handlers) GetTopOpenIssuesRepos (w http.ResponseWriter, r *http.Request) {
	hh.HandleViews(w, r, "top-repo-by-openissues")
}

func (hh *Handlers) GetTopStarredRepos (w http.ResponseWriter, r *http.Request) {
	hh.HandleViews(w, r, "top-repo-by-stars")
}


func (hh *Handlers) HandleViews(w http.ResponseWriter, r *http.Request, key string) {

	vars := mux.Vars(r)

	limit, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println("Wrong count specified", err.Error())
		w.WriteHeader(400)
		return
	}
   
	response := hh.cacher.GetView(key, limit)

	var craftedResp strings.Builder

	craftedResp.WriteString("[")
	for i, resp := range response {
		fmt.Fprintf(&craftedResp, "[%s, %s]", resp.Repo, resp.Count)
		if i != len(response)-1 {
			craftedResp.WriteString(",")
		}	
	}
	craftedResp.WriteString("]")

	fmt.Fprintf(w, craftedResp.String())
}

// a health check endpoint to let others know service is up and running
func (hh *Handlers) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)	
}

// Setuphandlers sets up the mux router with appropriate paths and handlers
func SetupHandlers(cacher *Cacher) http.Handler{
	r := mux.NewRouter()

	proxy := &Handlers{
		stub: GenerateProxy(),
		cacher: cacher,	
	}

	r.HandleFunc("/healthcheck", proxy.Healthcheck)

	cachedPaths := []string{"/",
			           "/orgs/Netflix",
			           "/orgs/Netflix/members",
			           "/orgs/Netflix/repos",
					 }

	for _, conf := range cachedPaths {
		r.HandleFunc(conf, proxy.HandleCachedAPI)
	}
	
	// handlers for views we have constructed over repository data
	viewr := r.PathPrefix("/view").Subrouter()
	viewr.HandleFunc("/top/{id}/forks", proxy.GetTopForkedRepos)
	viewr.HandleFunc("/top/{id}/last_updated", proxy.GetLastUpdatedRepos)
	viewr.HandleFunc("/top/{id}/open_issues", proxy.GetTopOpenIssuesRepos)
	viewr.HandleFunc("/top/{id}/stars", proxy.GetTopStarredRepos)

	// fallback to default handler for all the rest of paths
	r.PathPrefix("/").HandlerFunc(proxy.HandleDefaults)

	return SetupInterceptor(r)
}
