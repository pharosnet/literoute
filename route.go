package literoute

import (
	context0 "context"
	"net/http"
	"net/url"
	"strings"
)

const (
	tokenParam = 2
	tokenSub   = 4
	contextKey = "a_lite_route"
	matchOk    = 1
	matchNon   = 0
	matchFail  = -1
)

type token struct {
	raw    []int
	Tokens []string
	Size   int
}

func newRoute(mux *LiteMux, url string, h HandleFunc) *route {
	r := &route{Path: url, Handle: h, mux: mux}
	r.save()
	return r
}

type route struct {
	Path       string
	Method     string
	Size       int
	Attrs      int
	Token      token
	Pattern    map[int]string
	Handle     HandleFunc
	mux        *LiteMux
	validators map[string][]string
}

func (r *route) handle(rw http.ResponseWriter, req *http.Request) {
	ctx := acquireContext(r.mux, rw, req)
	if !r.mux.handleMiddleware(ctx) {
		releaseContext(ctx)
		return
	}
	r.Handle(ctx)
	releaseContext(ctx)
}

func (r *route) save() {
	r.Size = len(r.Path)
	r.Token.Tokens = strings.Split(r.Path, "/")
	for i, s := range r.Token.Tokens {
		if len(s) >= 1 {
			if s[:1] == ":" {
				s = s[1:]
				if r.Pattern == nil {
					r.Pattern = make(map[int]string)
				}
				if validators := containsValidators(s); validators != nil {
					if r.validators == nil {
						r.validators = make(map[string][]string)
					}
					for _, vali := range validators {
						s = s[:validators[0].start]
						r.validators[s] = append(r.validators[s], vali.name[1:])
					}
				}
				r.Pattern[i] = s
				r.Attrs |= tokenParam
			} else {
				r.Token.raw = append(r.Token.raw, i)
			}
		}
		r.Token.Size++
	}
}

func (r *route) match(rw http.ResponseWriter, req *http.Request) int {
	mr, _ := r.matchAndParse(rw, req)
	return mr
}

func (r *route) matchAndParse(rw http.ResponseWriter, req *http.Request) (int, map[string]string) {
	ss := strings.Split(req.URL.EscapedPath(), "/")
	if r.matchRawTokens(&ss) {
		if len(ss) == r.Token.Size {
			totalSize := len(r.Pattern)
			vars := make(map[string]string, totalSize)
			for k, v := range r.Pattern {
				if validators := r.validators[v]; validators != nil {
					for _, validatorName := range validators {
						validator := (*r.mux).validators[validatorName]
						if !validator.Validate(ss[k]) {
							ctx := acquireContext(r.mux, rw, req)
							validator.OnFail(ctx)
							releaseContext(ctx)
							return matchFail, nil
						}
					}
				}
				vars[v], _ = url.QueryUnescape(ss[k])
			}
			return matchOk, vars
		}
	}
	return matchNon, nil
}

func (r *route) parse(rw http.ResponseWriter, req *http.Request) (bool, int) {
	if r.Attrs != 0 {
		if r.Attrs&tokenSub != 0 {
			if len(req.URL.Path) >= r.Size {
				if req.URL.Path[:r.Size] == r.Path {
					req.URL.Path = req.URL.Path[r.Size:]
					r.handle(rw, req)
					return true, matchOk
				}
			}
		}

		if mr, vars := r.matchAndParse(rw, req); mr == matchOk {
			ctx0 := context0.WithValue(req.Context(), contextKey, vars)
			newReq := req.WithContext(ctx0)
			r.handle(rw, newReq)
			return true, matchOk
		} else if mr == matchFail {
			return true, matchFail
		} else if mr == matchNon {
			return false, matchNon
		}
	}
	if req.URL.Path == r.Path {
		r.handle(rw, req)
		return true, matchOk
	}
	return false, matchNon
}

func (r *route) matchRawTokens(ss *[]string) bool {
	if len(*ss) >= r.Token.Size {
		for _, v := range r.Token.raw {
			return (*ss)[v] == r.Token.Tokens[v]
		}
		return true
	}
	return false
}

func (r *route) exists(rw http.ResponseWriter, req *http.Request) bool {
	if r.Attrs != 0 {
		if r.Attrs&tokenSub != 0 {
			if len(req.URL.Path) >= r.Size {
				if req.URL.Path[:r.Size] == r.Path {
					return true
				}
			}
		}

		if mr, _ := r.matchAndParse(rw, req); mr == matchOk || mr == matchFail {
			return true
		}
	}
	if req.URL.Path == r.Path {
		return true
	}
	return false
}
