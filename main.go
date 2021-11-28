package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	localCfg = viper.New()
)

func getReqIp(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

func printAccessLog(r *http.Request, statusCode int) {
	fmt.Printf("Current request ip: %s\n", getReqIp(r))
	fmt.Printf("Current response status code: %d\n", statusCode)
}

func envCheckMiddleware(next http.Handler) http.Handler {
	checkSysEnv := func(w http.ResponseWriter) http.ResponseWriter {
		if os.Getenv("VERSION") != "" {
			w.Header().Set("VERSION", os.Getenv("VERSION"))
		}
		return w
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Info("Executing middleware of checking environment variables...")
		next.ServeHTTP(checkSysEnv(w), r)
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	for k, v := range r.Header {
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}
	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		glog.Error(err)
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)
	w.Write(resp)
	printAccessLog(r, statusCode)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	w.Header().Add("Content-Type", "application/json")
	resp, err := json.Marshal(map[string]string{})
	if err != nil {
		glog.Error(err)
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)
	w.Write(resp)
	printAccessLog(r, statusCode)
}

func loadConfigFile() {
	glog.Info("Loading configuration...")
	localCfg.SetConfigFile("config.yaml")
	localCfg.AddConfigPath("./conf/")
	localCfg.SetConfigName("config")
	localCfg.SetConfigType("yaml")
	if err := localCfg.ReadInConfig(); err != nil {
		glog.Fatalf("Failed to load initial configuration to run..., error: %v", err)
	}
}

func main() {
	flag.Parse()
	defer glog.Flush()
	loadConfigFile()
	glog.Info("Starting http server...")

	mux := http.NewServeMux()
	mux.Handle("/", envCheckMiddleware(http.HandlerFunc(rootHandler)))
	mux.HandleFunc("/healthz", healthzHandler)
	srv := http.Server{
		Addr:    fmt.Sprint(":", localCfg.GetInt("portNum")),
		Handler: mux,
	}
	processed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); nil != err {
			glog.Fatalf("Server shutdown failed, error: %v\n", err)
		}
		glog.Info("Server gracefully shutdown")

		close(processed)
	}()
	if err := srv.ListenAndServe(); http.ErrServerClosed != err {
		glog.Fatalf("Server not gracefully shutdown, err :%v\n", err)
	}

	<-processed
}
