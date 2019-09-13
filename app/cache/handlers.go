package cache

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"net/http/httputil"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
	"github.com/aniketalshi/go_rest_cache/app/logging"
	"github.com/aniketalshi/go_rest_cache/config"
)

// Handlers is a container for maintaining reference to all handlers and also maintains reference to reverse proxy stub 
type Handlers struct 
{
	stub *httputil.ReverseProxy
	cacher *Cacher
}

// HandleCachedAPI handles the api responses for path which are pre-cached in redis
func (hh *Handlers) HandleCachedAPI(w http.ResponseWriter, r *http.Request) {

	for _, url := range config.GetConfig().GetCachedURLs() {
		if r.URL.Path == url {
			response := hh.cacher.GetCachedEndpoint(url)

			w.WriteHeader(200)
			w.Write(response)

			return
		}
	}
	
    logging.Logger(r.Context()).Info("The requested path is not supposed to be cached",
    								 zap.String("path", r.URL.Path))
    hh.stub.ServeHTTP(w, r)	
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
	if err != nil || limit < 1 {

		logging.Logger(r.Context()).Error("Wrong count specified", zap.String("msg", err.Error()))

		w.WriteHeader(400)
		w.Write([]byte("count incorrect in request"))
		return
	}
	
	// fetch the vewi from redis
	response, err := hh.cacher.GetView(r.Context(), key, limit)

	if err != nil {
		w.WriteHeader(400)		
		w.Write([]byte(err.Error()))
		return
	}

	var craftedResp strings.Builder
	
	// craft the response as flattened list of lists
	craftedResp.WriteString("[")
	for i, resp := range response {
		fmt.Fprintf(&craftedResp, "[%s, %s]", resp.Repo, resp.Count)
		if i != len(response)-1 {
			craftedResp.WriteString(",")
		}	
	}
	craftedResp.WriteString("]")

	logging.Logger(r.Context()).Info("custom view response", 
									  zap.Int("len", len(response)))

	w.WriteHeader(200)
	w.Write([]byte(craftedResp.String()))
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

	for _, url := range config.GetConfig().GetCachedURLs() {
		r.HandleFunc(url, proxy.HandleCachedAPI)
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
