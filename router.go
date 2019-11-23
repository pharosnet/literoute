package literoute

import (
	"net/http"
	"strings"
)

type Router struct {
	prefix string
	mux    *LiteMux
}

func newRouter(path string, mux *LiteMux) *Router {
	return &Router{
		prefix: strings.TrimSuffix(path, "/"),
		mux:    mux,
	}
}

func (r *Router) register(method string, path string, handle HandleFunc) {
	route := newRoute(r.mux, r.prefix+path, handle)
	route.Method = method
	if valid(path) {
		r.mux.routes[method] = append(r.mux.routes[method], route)
		return
	}
	r.mux.routes[static] = append(r.mux.routes[static], route)
}

func (r *Router) Get(path string, handle HandleFunc) {
	r.register(http.MethodGet, path, handle)
}

func (r *Router) Post(path string, handle HandleFunc) {
	r.register(http.MethodPost, path, handle)
}

func (r *Router) Put(path string, handle HandleFunc) {
	r.register(http.MethodPut, path, handle)
}

func (r *Router) Delete(path string, handle HandleFunc) {
	r.register(http.MethodDelete, path, handle)
}

func (r *Router) Head(path string, handle HandleFunc) {
	r.register(http.MethodHead, path, handle)
}

func (r *Router) Patch(path string, handle HandleFunc) {
	r.register(http.MethodPatch, path, handle)
}

func (r *Router) Options(path string, handle HandleFunc) {
	r.register(http.MethodOptions, path, handle)
}

func (r *Router) Trace(path string, handle HandleFunc) {
	r.register(http.MethodTrace, path, handle)
}

func (r *Router) Connect(path string, handle HandleFunc) {
	r.register(http.MethodConnect, path, handle)
}
