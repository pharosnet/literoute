package literoute

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	tokenParam = 2
	tokenSub   = 4
	contextKey = "a_lite_route"
)

type token struct {
	raw    []int
	Tokens []string
	Size   int
}

func newRoute(mux *LiteMux, url string, h http.Handler) *route {
	r := &route{Path: url, Handler: h, mux: mux}
	r.save()
	return r
}

type route struct {
	Path   string
	Method string
	Size   int
	Attrs  int
	//wildPos    int
	Token      token
	Pattern    map[int]string
	Compile    map[int]*regexp.Regexp
	Tag        map[int]string
	Handler    http.Handler
	mux        *LiteMux
	validators map[string][]string
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

func (r *route) Match(req *http.Request) bool {
	ok, _ := r.matchAndParse(req)
	return ok
}

func (r *route) matchAndParse(req *http.Request) (bool, map[string]string) {
	ss := strings.Split(req.URL.EscapedPath(), "/")
	if r.matchRawTokens(&ss) {
		if len(ss) == r.Token.Size {
			totalSize := len(r.Pattern)
			vars := make(map[string]string, totalSize)
			for k, v := range r.Pattern {
				if validators := r.validators[v]; validators != nil {
					for _, validatorName := range validators {
						if !(*r.mux).Validators[validatorName].Validate(ss[k]) {
							return false, nil
						}
					}
				}
				vars[v], _ = url.QueryUnescape(ss[k])
			}
			return true, vars
		}
	}
	return false, nil
}

func (r *route) parse(rw http.ResponseWriter, req *http.Request) bool {
	if r.Attrs != 0 {
		if r.Attrs&tokenSub != 0 {
			if len(req.URL.Path) >= r.Size {
				if req.URL.Path[:r.Size] == r.Path {
					req.URL.Path = req.URL.Path[r.Size:]
					r.Handler.ServeHTTP(rw, req)
					return true
				}
			}
		}

		if ok, vars := r.matchAndParse(req); ok {
			ctx := context.WithValue(req.Context(), contextKey, vars)
			newReq := req.WithContext(ctx)
			r.Handler.ServeHTTP(rw, newReq)
			return true
		}
	}
	if req.URL.Path == r.Path {
		r.Handler.ServeHTTP(rw, req)
		return true
	}
	return false
}

func (r *route) matchRawTokens(ss *[]string) bool {
	if len(*ss) >= r.Token.Size {
		for _, v := range r.Token.raw {
			return (*ss)[v] == r.Token.Tokens[v]
			//if (*ss)[v] != r.Token.Tokens[v] {
			//	if r.Atts&WC != 0 && r.wildPos == i {
			//		return true
			//	}
			//	return false
			//}
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

		if ok, _ := r.matchAndParse(req); ok {
			return true
		}
	}
	if req.URL.Path == r.Path {
		return true
	}
	return false
}
