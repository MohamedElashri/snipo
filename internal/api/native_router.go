package api

import (
	"context"
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

type route struct {
	method      string
	pattern     string // e.g. /api/v1/snippets/{id}
	handler     http.Handler
	pathSegs    []string
}

// Global routes list. To avoid complex reference passing, we can pass a pointer to a shared slice.
type RouterCore struct {
	routes []route
}

type Router struct {
	core        *RouterCore
	middlewares []Middleware
	basePath    string
}

func NewNativeRouter() *Router {
	return &Router{
		core:        &RouterCore{},
		middlewares: []Middleware{},
		basePath:    "",
	}
}

func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) With(middlewares ...Middleware) *Router {
	newMws := make([]Middleware, len(r.middlewares)+len(middlewares))
	copy(newMws, r.middlewares)
	copy(newMws[len(r.middlewares):], middlewares)
	return &Router{
		core:        r.core,
		middlewares: newMws,
		basePath:    r.basePath,
	}
}

func (r *Router) Group(fn func(r *Router)) {
	subRouter := r.With()
	fn(subRouter)
}

func (r *Router) Route(pattern string, fn func(r *Router)) {
	subRouter := r.With()
	pattern = strings.TrimSuffix(pattern, "/")
	if r.basePath == "/" {
		subRouter.basePath = pattern
	} else {
		subRouter.basePath = r.basePath + pattern
	}
	fn(subRouter)
}

func (r *Router) handle(method, pattern string, handler http.Handler) {
	fullPattern := pattern
	if r.basePath != "" {
		if pattern == "/" {
			fullPattern = r.basePath
		} else {
			fullPattern = r.basePath + pattern
		}
	}
	
	finalHandler := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		finalHandler = r.middlewares[i](finalHandler)
	}
	
	r.core.routes = append(r.core.routes, route{
		method:   method,
		pattern:  fullPattern,
		handler:  finalHandler,
		pathSegs: strings.Split(strings.Trim(fullPattern, "/"), "/"),
	})
	println("ROUTER TRACE: Registered route", method, fullPattern)
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodGet, pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodPost, pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodPut, pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodDelete, pattern, handler)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.handle("", pattern, handler)
}

func (r *Router) Mount(pattern string, handler http.Handler) {
    pattern = strings.TrimSuffix(pattern, "/")
	r.handle("MOUNT", pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	println("ROUTER TRACE: ServeHTTP called for", req.URL.Path)
	reqPath := strings.Trim(req.URL.Path, "/")
	reqSegs := strings.Split(reqPath, "/")

	for _, route := range r.core.routes {
		if route.method != "MOUNT" && route.method != "" && route.method != req.Method {
			continue
		}

		if route.method == "MOUNT" {
			if strings.HasPrefix(req.URL.Path, route.pattern) {
				http.StripPrefix(route.pattern, route.handler).ServeHTTP(w, req)
				return
			}
			continue
		}

		if len(reqSegs) != len(route.pathSegs) && !strings.HasSuffix(route.pattern, "/*") {
            if strings.HasSuffix(route.pattern, "/*") {
                prefix := strings.TrimSuffix(route.pattern, "/*")
                if strings.HasPrefix(req.URL.Path, prefix) {
                    route.handler.ServeHTTP(w, req)
                    return
                }
            }
			continue
		}

		match := true
		params := make(map[string]string)

		for i, seg := range route.pathSegs {
            if i >= len(reqSegs) && seg == "*" {
                break
            }
            if i >= len(reqSegs) {
                match = false
                break
            }
			if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
				paramName := seg[1 : len(seg)-1]
				params[paramName] = reqSegs[i]
			} else if seg == "*" {
                break
			} else if seg != reqSegs[i] {
				match = false
				break
			}
		}

		if match {
			ctx := req.Context()
			if len(params) > 0 {
				ctx = context.WithValue(ctx, "path_params", params)
			}
			route.handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}

	http.NotFound(w, req)
}
