// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package module

import (
	"context"
	"net/http"
	"runtime/pprof"
	"sync"

	"github.com/gorilla/mux"
)

// Router provides a wrapper around mux.Router so routes can be re-registered
// This is needed to support the module-restart feature
type Router struct {
	mux            sync.Mutex
	handlerByRoute map[string]func(http.ResponseWriter, *http.Request)
	registered     map[string]bool
	router         *mux.Router
	labels         pprof.LabelSet
}

// NewRouter returns a new Router
func NewRouter(namespace string, parent *mux.Router) *Router {
	return &Router{
		handlerByRoute: make(map[string]func(http.ResponseWriter, *http.Request)),
		registered:     make(map[string]bool),
		router:         parent.PathPrefix("/" + namespace).Subrouter(),
		labels:         pprof.Labels("module", namespace),
	}
}

// HandleFunc registers a HandleFunc in such a way that routes can be registered multiple times
func (r *Router) HandleFunc(path string, responseWriter func(http.ResponseWriter, *http.Request)) *mux.Route {
	r.mux.Lock()
	_, registered := r.registered[path]
	r.registered[path] = true
	// overwrite the handler regardless if it was registered before
	r.handlerByRoute[path] = responseWriter
	r.mux.Unlock()

	if registered {
		// If this route was previously registered there is nothing left to do.
		// The return value serves as a stub to support modules that are (re)registering routes
		// chaining calls like HandleFunc(path, handler).Method("POST")
		return new(mux.Route)
	}

	return r.router.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		r.mux.Lock()
		// obtain the current handler inline, which allows module restart
		handlerFn, ok := r.handlerByRoute[path]
		r.mux.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		pprof.Do(req.Context(), r.labels, func(_ context.Context) {
			handlerFn(w, req)
		})
	})
}

// Unregister removes the registered handler functions
func (r *Router) Unregister() {
	r.mux.Lock()
	defer r.mux.Unlock()
	clear(r.handlerByRoute)
}
