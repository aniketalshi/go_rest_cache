package main

import (
	"time"
	"net"
	"net/http"
	"context"
	"net/http/httputil"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"github.com/aniketalshi/go_rest_cache/logging"
)

// refers to context of underlying process
var httpContext = context.Background()

// ReverseProxy is a container for proxy stub and corresponding handler interfaces
//type ReverseProxy struct 
//{
//	stub *httputil.ReverseProxy
//}
//
//// HandleCachedAPI handles the api responses for path which are pre-cached in redis
//func (rp *ReverseProxy) HandleCachedAPI(w http.ResponseWriter, r *http.Request) {
//	rp.stub.ServeHTTP(w, r)	
//}
//
//// HandleDefaults is the default http handler
//func (rp *ReverseProxy) HandleDefaults (w http.ResponseWriter, r *http.Request) {
//	rp.stub.ServeHTTP(w, r)	
//}

// SetupInterceptor is a middleware function which acts like a wrapper over handler
// Generates request id, associate it with the context which is passed along api calls
func SetupInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	
		// TODO: add metrics 

		// generate a unique request id
		requestID := uuid.New()
	
		// tag the context with newly generated req id
		requestCtx := logging.WithRequestID(r.Context(), requestID.String())
	
		// retreive the logger with this particular context settings
		logger := logging.Logger(requestCtx)
	
		logger.Info("Request received",
					 zap.String("method", r.Method),
					 zap.String("uri", r.RequestURI))

		next.ServeHTTP(w, r.WithContext(requestCtx))
	})
}

type ProxyConfig struct {
	Path     string
	Host     string
}

//func GenerateProxy(conf ProxyConfig) http.Handler {
func GenerateProxy(conf ProxyConfig) *httputil.ReverseProxy {

	proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {

		originHost := conf.Host
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originHost)
		req.Host = originHost
		req.URL.Host = originHost
		req.URL.Scheme = "https"

	}, Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
	}}

	return proxy
}
