package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

func handleWithBasicAuth(pattern string, user, pwd string, handlerFunc http.HandlerFunc) {
	if len(user) > 0 {
		http.HandleFunc(pattern, basicAuth(handlerFunc, []byte(user), []byte(pwd)))
		return
	}
	http.HandleFunc(pattern, handlerFunc)
}

func basicAuth(f http.HandlerFunc, user, pwd []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		basicAuthPrefix := "Basic "
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			payload, err := base64.StdEncoding.DecodeString(
				auth[len(basicAuthPrefix):],
			)
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 && bytes.Equal(pair[0], user) &&
					bytes.Equal(pair[1], pwd) {
					f(w, r)
					return
				}
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func debounce(f func(), d time.Duration) func() {
	mu := sync.Mutex{}
	var timer *time.Timer
	return func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(d, f)
	}
}
