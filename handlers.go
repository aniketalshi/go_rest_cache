package main

import(
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
			Path: "/orgs/Netflix",
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
