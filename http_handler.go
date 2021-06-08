package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mailgun/groupcache"
)

// SimpleHandler just a smiple handler that gives cached value
type SimpleHandler struct {
	IP         string
	IPExternal string
}

func (s *SimpleHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")
	if name == "" {
		fmt.Fprintf(w, "Missing parameter `name` :(\n")
		return
	}

	cacheGroup := groupcache.GetGroup("generator")
	var cacheValue string
	err := cacheGroup.Get(context.Background(), name, groupcache.StringSink(&cacheValue))
	if err != nil {
		fmt.Fprintf(w, "Failed to get cache :(\n")
		return
	}

	fmt.Fprintf(w, "You are served by %v (%v).\nCache data: `%v`.\n", s.IP, s.IPExternal, cacheValue)
}
