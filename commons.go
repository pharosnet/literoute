package literoute

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func cleanURL(url *string) {
	_url := *url
	urlLength := len(_url)
	if urlLength > 1 {
		if (*url)[urlLength-1:] == "/" {
			*url = (*url)[:urlLength-1]
			cleanURL(url)
		}
	}
}

func valid(path string) bool {
	pathLength := len(path)
	if pathLength > 1 && path[pathLength-1:] == "/" {
		return false
	}
	return true
}

func DecodeQuery(path string) string {
	if path == "" {
		return ""
	}
	encodedPath, err := url.QueryUnescape(path)
	if err != nil {
		return path
	}
	return encodedPath
}

func DecodeURL(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	return u.String()
}

func GetHost(r *http.Request) string {
	if host := r.Host; host != "" {
		return host
	}

	return r.URL.Host
}

func (ctx *context) GetHeader(name string) string {
	return ctx.request.Header.Get(name)
}

func GetForm(r *http.Request, postMaxMemory int64, resetBody bool) (form map[string][]string, found bool) {

	if form := r.Form; len(form) > 0 {
		return form, true
	}

	if form := r.PostForm; len(form) > 0 {
		return form, true
	}

	if m := r.MultipartForm; m != nil {
		if len(m.Value) > 0 {
			return m.Value, true
		}
	}

	var bodyCopy []byte

	if resetBody {
		if m := r.Method; m == "POST" || m == "PUT" || m == "PATCH" {
			bodyCopy, _ = GetBody(r, resetBody)
			if len(bodyCopy) == 0 {
				return nil, false
			}
		} else {
			resetBody = false
		}
	}

	err := r.ParseMultipartForm(postMaxMemory)
	if resetBody {
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyCopy))
	}
	if err != nil && err != http.ErrNotMultipart {
		return nil, false
	}

	if form := r.Form; len(form) > 0 {
		return form, true
	}

	if form := r.PostForm; len(form) > 0 {
		return form, true
	}

	if m := r.MultipartForm; m != nil {
		if len(m.Value) > 0 {
			return m.Value, true
		}
	}

	return nil, false
}

func (ctx *context) FormValueDefault(name string, def string) string {
	if form, has := ctx.form(); has {
		if v := form[name]; len(v) > 0 {
			return v[0]
		}
	}
	return def
}

func FormValueDefault(r *http.Request, name string, def string, postMaxMemory int64, resetBody bool) string {
	if form, has := GetForm(r, postMaxMemory, resetBody); has {
		if v := form[name]; len(v) > 0 {
			return v[0]
		}
	}
	return def
}

func GetBody(r *http.Request, resetBody bool) ([]byte, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if resetBody {
		r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}

	return data, nil
}

var UnixEpochTime = time.Unix(0, 0)

func IsZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(UnixEpochTime)
}

var ParseTimeRFC3339 = func(ctx Context, text string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC3339, text)
	if err != nil {
		return http.ParseTime(text)
	}

	return
}

var ParseTimeRFC3339Nano = func(ctx Context, text string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return http.ParseTime(text)
	}

	return
}

func FormatTimeRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func FormatTimeRFC3339Nano(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func parseHeader(headerValue string) []string {
	in := strings.Split(headerValue, ",")
	out := make([]string, 0, len(in))

	for _, value := range in {
		v := strings.TrimSpace(strings.Split(value, ";")[0])
		if v != "" {
			out = append(out, v)
		}
	}

	return out
}
