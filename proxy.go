package main

import (
	"time"
	"net"
	"net/http"
	"fmt"
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
	
		// generate a unique request id
		requestID := uuid.New()

		reqCtx := logging.NewContext(httpContext, zap.Stringer("requestID", requestID))

		logger := logging.Logger(reqCtx)

		logger.Info("Request received",
					 zap.String("method", r.Method),
					 zap.String("uri", r.RequestURI))

		next.ServeHTTP(w, r.WithContext(reqCtx))
	})
}

func GenerateProxy() *httputil.ReverseProxy {
	
	// get the configuration parameters about the upstream target 
	origin := config.GetConfig().GetTarget()

	if origin.Token == "" {
		fmt.Println("Token is not set")
	}

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
