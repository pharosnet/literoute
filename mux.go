package literoute

import (
	"net/http"
	"strings"
)

func New() (mux *LiteMux) {

	return
}

var (
	static   = "static"
	location = "Location"
	methods  = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodPatch, http.MethodOptions, http.MethodConnect, http.MethodTrace}
)

type LiteMux struct {
	routes     map[string][]*route
	prefix     string
	notFound   http.Handler
	validators map[string]Validator
	Serve      func(rw http.ResponseWriter, req *http.Request)
}

func (m *LiteMux) Prefix(p string) *LiteMux {
	m.prefix = strings.TrimSuffix(p, "/")
	return m
}

func (m *LiteMux) RegisterValidator(name string, validator Validator) {
	if m.validators == nil {
		m.validators = make(map[string]Validator)
	}
	m.validators[name] = validator
}

func (m *LiteMux) parse(rw http.ResponseWriter, req *http.Request) bool {
	for _, r := range m.routes[req.Method] {
		ok := r.parse(rw, req)
		if ok {
			return true
		}
	}

	if req.Method == "HEAD" {
		for _, r := range m.routes["GET"] {
			ok := r.parse(rw, req)
			if ok {
				return true
			}
		}
	}

	return false
}

func (m *LiteMux) staticRoute(rw http.ResponseWriter, req *http.Request) bool {
	for _, s := range m.routes[static] {
		if len(req.URL.Path) >= s.Size {
			if req.URL.Path[:s.Size] == s.Path {
				s.Handler.ServeHTTP(rw, req)
				return true
			}
		}
	}
	return false
}

func (m *LiteMux) validate(rw http.ResponseWriter, req *http.Request) bool {
	pathLength := len(req.URL.Path)
	if pathLength > 1 && req.URL.Path[pathLength-1:] == "/" {
		cleanURL(&req.URL.Path)
		rw.Header().Set(location, req.URL.String())
		rw.WriteHeader(http.StatusFound)
		return true
	}

	return m.parse(rw, req)
}

func (m *LiteMux) otherMethods(rw http.ResponseWriter, req *http.Request) bool {
	for _, method := range methods {
		if method != req.Method {
			for _, r := range m.routes[method] {
				ok := r.exists(rw, req)
				if ok {
					rw.WriteHeader(http.StatusMethodNotAllowed)
					return true
				}
			}
		}
	}
	return false
}

func (m *LiteMux) HandleNotFound(rw http.ResponseWriter, req *http.Request) {
	if m.notFound != nil {
		m.notFound.ServeHTTP(rw, req)
	} else {
		http.NotFound(rw, req)
	}
}

func (m *LiteMux) serve(rw http.ResponseWriter, req *http.Request) {
	if !m.parse(rw, req) {
		if !m.staticRoute(rw, req) {
			if !m.validate(rw, req) {
				if !m.otherMethods(rw, req) {
					m.HandleNotFound(rw, req)
				}
			}
		}
	}
}
