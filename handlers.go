package main

import(
	//"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"net/http/httputil"
)

// Handlers is a container for maintaining reference to all handlers and also maintains reference to reverse proxy stub 
type Handlers struct 
{
	stub *httputil.ReverseProxy
}

// HandleCachedAPI handles the api responses for path which are pre-cached in redis
func (hh *Handlers) HandleCachedAPI(w http.ResponseWriter, r *http.Request) {

	//githubClient := GetNewGithubClient(r.Context())
	//orgs, _, err := githubClient.Stub.Repositories.ListCommits(r.Context(), "google", "go-github", nil)
	//
	//if err != nil {
	//	fmt.Println(err.Error())
	//} else {
	//	fmt.Println(orgs)
	//}

	hh.stub.ServeHTTP(w, r)	
}

// HandleDefaults is the default http handler
func (hh *Handlers) HandleDefaults (w http.ResponseWriter, r *http.Request) {
	hh.stub.ServeHTTP(w, r)	
}

func SetupHandlers() http.Handler{
	r := mux.NewRouter()

	configuration := []ProxyConfig{
		ProxyConfig{
			Path: "/",
			Host: "api.github.com",
		},
		ProxyConfig{
			Path: "/orgs/Netflix",
			Host: "api.github.com",
		},
		ProxyConfig{
			Path: "/orgs/Netflix/members",
			Host: "api.github.com",
		},
		ProxyConfig{
			Path: "/orgs/Netflix/repos",
			Host: "api.github.com",
		},
	}

	for _, conf := range configuration {
		proxy := &Handlers{
			stub: GenerateProxy(conf),
		}
		r.HandleFunc(conf.Path, proxy.HandleCachedAPI)
	}
	
	proxy := &Handlers{
		stub: GenerateProxy(configuration[0]),
	}
	// fallback to default handler for all the rest of paths
	r.PathPrefix("/").HandlerFunc(proxy.HandleDefaults)
	
	return SetupInterceptor(r)	
}
