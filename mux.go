package literoute

import (
	"net/http"
)

func New(config Config) (mux *LiteMux) {
	mux = &LiteMux{
		config:           config,
		routes:           make(map[string][]*route),
		validators:       make(map[string]Validator),
		middlewareList:   make([]Middleware, 0, 1),
		middlewareNum:    0,
		extraBodyEncoder: nil,
	}
	mux.rootRouter = newRouter("/", mux)
	return
}

func Default() (mux *LiteMux) {
	return New(DefaultConfig)
}

var (
	static   = "static"
	location = "Location"
	methods  = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodPatch, http.MethodOptions, http.MethodConnect, http.MethodTrace}
)

const (
	JsonBodyEncode = iota
	XmlBodyEncode
)

type Config struct {
	BodyEncoder   int
	Status        CustomizeStatus
	PostMaxMemory int64
}

type CustomizeStatus struct {
	Succeed        int
	Fail           int
	NotFound       int
	InvalidRequest int
}

var DefaultConfig = Config{
	BodyEncoder: JsonBodyEncode,
	Status: CustomizeStatus{
		Succeed:        200,
		Fail:           555,
		NotFound:       444,
		InvalidRequest: 440,
	},
	PostMaxMemory: DefaultPostMaxMemory,
}

type LiteMux struct {
	config           Config
	rootRouter       *Router
	routes           map[string][]*route
	notFound         HandleFunc
	validators       map[string]Validator
	middlewareNum    int
	middlewareList   []Middleware
	extraBodyEncoder BodyEncoder
}

func (m *LiteMux) AppendMiddleware(mid Middleware) {
	m.middlewareList = append(m.middlewareList, mid)
	m.middlewareNum = len(m.middlewareList)
}

func (m *LiteMux) RegisterValidator(name string, validator Validator) {
	if m.validators == nil {
		m.validators = make(map[string]Validator)
	}
	m.validators[name] = validator
}

func (m *LiteMux) RegisterBodyEncode(encode BodyEncoder) {
	m.extraBodyEncoder = encode
}

func (m *LiteMux) getExtraBodyEncoder() BodyEncoder {
	return m.extraBodyEncoder
}

func (m *LiteMux) getConfig() Config {
	return m.config
}

func (m *LiteMux) parse(rw http.ResponseWriter, req *http.Request) (bool, int) {
	for _, r := range m.routes[req.Method] {
		ok, match := r.parse(rw, req)
		if ok {
			return true, match
		}
	}

	if req.Method == http.MethodHead {
		for _, r := range m.routes[http.MethodGet] {
			ok, match := r.parse(rw, req)
			if ok {
				return true, match
			}
		}
	}

	return false, matchNon
}

func (m *LiteMux) staticRoute(rw http.ResponseWriter, req *http.Request) bool {
	for _, s := range m.routes[static] {
		if len(req.URL.Path) >= s.Size {
			if req.URL.Path[:s.Size] == s.Path {
				ctx := acquireContext(m, rw, req)
				s.Handle(ctx)
				releaseContext(ctx)
				return true
			}
		}
	}
	return false
}

func (m *LiteMux) validate(rw http.ResponseWriter, req *http.Request) (bool, int) {
	pathLength := len(req.URL.Path)
	if pathLength > 1 && req.URL.Path[pathLength-1:] == "/" {
		cleanURL(&req.URL.Path)
		rw.Header().Set(location, req.URL.String())
		rw.WriteHeader(http.StatusFound)
		return true, matchOk
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

func (m *LiteMux) handleNotFound(rw http.ResponseWriter, req *http.Request) {
	if m.notFound != nil {
		ctx := acquireContext(m, rw, req)
		m.notFound(ctx)
		releaseContext(ctx)
	} else {
		http.NotFound(rw, req)
	}
}

func (m *LiteMux) handleMiddleware(ctx Context) bool {
	if m.middlewareNum > 0 {
		for _, mid := range m.middlewareList {
			if !mid.Handle(ctx) {
				return false
			}
		}
	}
	return true
}

func (m *LiteMux) serve(rw http.ResponseWriter, req *http.Request) {
	if _, match := m.parse(rw, req); match != matchOk && match != matchFail {
		if !m.staticRoute(rw, req) {
			if _, match := m.validate(rw, req); match != matchOk && match != matchFail {
				if !m.otherMethods(rw, req) {
					m.handleNotFound(rw, req)
				}
			}
		}
	}
}

func (m *LiteMux) Party(path string) *Router {
	return newRouter(path, m)
}

func (m *LiteMux) Get(path string, handle HandleFunc) {
	m.rootRouter.Get(path, handle)
}

func (m *LiteMux) Post(path string, handle HandleFunc) {
	m.rootRouter.Post(path, handle)
}

func (m *LiteMux) Put(path string, handle HandleFunc) {
	m.rootRouter.Put(path, handle)
}

func (m *LiteMux) Delete(path string, handle HandleFunc) {
	m.rootRouter.Delete(path, handle)
}

func (m *LiteMux) Head(path string, handle HandleFunc) {
	m.rootRouter.Head(path, handle)
}

func (m *LiteMux) Patch(path string, handle HandleFunc) {
	m.rootRouter.Patch(path, handle)
}

func (m *LiteMux) Options(path string, handle HandleFunc) {
	m.rootRouter.Options(path, handle)
}

func (m *LiteMux) Trace(path string, handle HandleFunc) {
	m.rootRouter.Trace(path, handle)
}

func (m *LiteMux) Connect(path string, handle HandleFunc) {
	m.rootRouter.Connect(path, handle)
}

func (m *LiteMux) NotFound(handle HandleFunc) {
	m.notFound = handle
}

func (m *LiteMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.serve(rw, req)
}
