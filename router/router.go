package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// TODO: implement radix tree for better performance.
// TODO: has to return a http.Handler instead of router.HandlerFunc

type HandlerFunc http.HandlerFunc

// Middleware
type MiddlewareFunc func(h http.Handler) http.Handler

// Router is a simple HTTP request router tha uses a simple
// map of routerEntry array to store the handlers.
type Router struct {
	mu      sync.RWMutex
	entries map[string][]*routerEntry

	// Called when no matching route is found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Function to handle panics recovered from http handlers.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

// Returns a new Router
func New() *Router {
	return &Router{}
}

// represents a router entry
type routerEntry struct {
	HasParams bool
	//params    map[string]string
	Handler  HandlerFunc
	Pattern  string
	segments []string
}

// add a new router entry for a given path, method and handler
func (r *Router) add(method, pattern string, handler HandlerFunc, middleware ...MiddlewareFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if pattern == "" {
		return fmt.Errorf("router - invalid pattern: %s", pattern)
	}

	if pattern[0] != '/' {
		return fmt.Errorf("router - path must begin with '/': %s", pattern)
	}

	if handler == nil {
		return fmt.Errorf("router - nil handler for pattern %s", pattern)
	}

	if r.entries == nil {
		r.entries = make(map[string][]*routerEntry)
	}

	// remove the last /
	n := len(pattern)
	if pattern != "/" && n > 0 && pattern[n-1] == '/' {
		pattern = pattern[:n-1]
	}

	segments := strings.Split(pattern[1:], "/")

	// Checks if this router was already added
	if _, exists := r.entries[method]; exists {
		for _, entry := range r.entries[method] {
			if entry.Pattern == pattern {
				return fmt.Errorf("router - multiple registrations for pattern %s", pattern)
			}
		}
	}

	// Check if there's any conflicts
	hasParam := strings.Contains(pattern, ":")
	if h, _ := r.find(method, pattern, hasParam); h != nil {
		return fmt.Errorf("router - the pattern %s matched with %s", pattern, h.Pattern)
	}

	hasParams := strings.Contains(pattern, ":")

	r.entries[method] = append(r.entries[method], &routerEntry{HasParams: hasParams, Handler: handler, Pattern: pattern, segments: segments})
	return nil
}

// Returns the router entry and params to use for the given request.
// Returns nil if not match was found.
func (r *Router) find(method, url string, hasParam bool) (*routerEntry, map[string]string) {
	for _, entry := range r.entries[method] {
		// If entry has param, we will compare each segment
		if entry.HasParams || hasParam {
			params := map[string]string{}
			segments := strings.Split(url[1:], "/")
			segmentLen := len(segments)

			// Should have the same length
			if segmentLen != len(entry.segments) {
				continue
			}

			match := true
			for i, value := range segments {
				if entry.segments[i][0] == ':' {
					params[entry.segments[i]] = value
					continue
				}

				if value[0] == ':' || entry.segments[i] == value {
					continue
				}

				match = false
				break
			}

			if match {
				return entry, params
			}
		} else if entry.Pattern == url {
			return entry, nil
		}

	}

	return nil, nil
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle("GET", path, handle, middleware...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle("HEAD", path, handle, middleware...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle("OPTIONS", path, handle, middleware...)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle("POST", path, handle, middleware...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle("PUT", path, handle, middleware...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle("PATCH", path, handle, middleware...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	r.Handle("DELETE", path, handle, middleware...)
}

// Handle registers a new request handle for the given path and method.
// If a handler already exists for pattern and method, it will panics.
func (r *Router) Handle(method string, path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
	if err := r.add(strings.ToUpper(method), path, handle, middleware...); err != nil {
		panic(err)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if r.PanicHandler != nil {
		defer func(w http.ResponseWriter, req *http.Request) {
			if rcv := recover(); rcv != nil {
				r.PanicHandler(w, req, rcv)
			}
		}(w, req)
	}

	entry, params := r.find(req.Method, req.URL.Path, false)

	// if nothing was found show the noFound or methodNotAllowed
	if entry == nil {
		r.noFound(w, req)
		return
	}

	for k, v := range params {
		if req.Form == nil {
			req.Form = make(url.Values)
		}
		req.Form.Add(k, v)
	}

	entry.Handler(w, req)
}

// Internal no found handler.
func (r *Router) noFound(w http.ResponseWriter, req *http.Request) {
	// HandlerFunc 404
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}
