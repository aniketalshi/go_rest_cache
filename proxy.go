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
	"github.com/aniketalshi/go_rest_cache/config"
)

// refers to context of underlying process
var httpContext = context.Background()

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

func GenerateProxy() *httputil.ReverseProxy {
	
	// get the configuration parameters about the upstream target 
	origin := config.GetConfig().GetTarget()

	proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Url)
		req.Header.Add("Authorization", origin.Token)
		req.Host = origin.Url
		req.URL.Host = origin.Url
		req.URL.Scheme = "https"

	}, Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(origin.Timeout) * time.Second,
		}).Dial,
	}}

	return proxy
}
