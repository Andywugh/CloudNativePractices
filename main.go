package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"log"
	"net/http"
	"os"
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
		log.Println("Executing middleware of checking environment variables...")
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
		log.Println(err)
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
		log.Println(err)
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)
	w.Write(resp)
	printAccessLog(r, statusCode)
}

func main() {
	glog.V(2).Info("Starting http server...")
	http.Handle("/", envCheckMiddleware(http.HandlerFunc(rootHandler)))
	http.HandleFunc("/healthz", healthzHandler)
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
