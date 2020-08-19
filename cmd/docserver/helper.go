package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type basicAuthHandler struct {
	user, pwd string
	handler   http.Handler
}

func newBasicAuthHandler(user string, pwd string, handler http.Handler) *basicAuthHandler {
	return &basicAuthHandler{user: user, pwd: pwd, handler: handler}
}

func (b *basicAuthHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(b.user) == 0 {
		b.handler.ServeHTTP(rw, req)
		return
	}
	basicAuthPrefix := "Basic "
	auth := req.Header.Get("Authorization")
	if strings.HasPrefix(auth, basicAuthPrefix) {
		payload, err := base64.StdEncoding.DecodeString(
			auth[len(basicAuthPrefix):],
		)
		if err == nil {
			pair := bytes.SplitN(payload, []byte(":"), 2)
			if len(pair) == 2 && bytes.Equal(pair[0], []byte(b.user)) &&
				bytes.Equal(pair[1], []byte(b.pwd)) {
				b.handler.ServeHTTP(rw, req)
				return
			}
		}
	}
	rw.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	rw.WriteHeader(http.StatusUnauthorized)
}

func handleWithBasicAuth(pattern string, user, pwd string, handlerFunc http.HandlerFunc) {
	if len(user) == 0 {
		http.HandleFunc(pattern, handlerFunc)
		return
	}
	http.HandleFunc(pattern, basicAuth(handlerFunc, []byte(user), []byte(pwd)))
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

func createSrcDirLink(dir string) error {
	srcPath := "/mkdoc/src"
	envSrcPath := os.Getenv("SRCPATH")
	if len(envSrcPath) > 0 {
		srcPath = envSrcPath
	}
	linkName := filepath.Join(dir, "src")
	os.Remove(linkName)
	return os.Symlink(srcPath, linkName)
}
