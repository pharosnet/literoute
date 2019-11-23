package literoute

import (
	"net/http"
	"time"
)

type CookieOption func(*http.Cookie)

func CookiePath(path string) CookieOption {
	return func(c *http.Cookie) {
		c.Path = path
	}
}

func CookieCleanPath(c *http.Cookie) {
	c.Path = ""
}

func CookieExpires(durFromNow time.Duration) CookieOption {
	return func(c *http.Cookie) {
		c.Expires = time.Now().Add(durFromNow)
		c.MaxAge = int(durFromNow.Seconds())
	}
}

func CookieHTTPOnly(httpOnly bool) CookieOption {
	return func(c *http.Cookie) {
		c.HttpOnly = httpOnly
	}
}

type (
	CookieEncoder func(cookieName string, value interface{}) (string, error)
	CookieDecoder func(cookieName string, cookieValue string, v interface{}) error
)

func CookieEncode(encode CookieEncoder) CookieOption {
	return func(c *http.Cookie) {
		newVal, err := encode(c.Name, c.Value)
		if err != nil {
			c.Value = ""
		} else {
			c.Value = newVal
		}
	}
}

func CookieDecode(decode CookieDecoder) CookieOption {
	return func(c *http.Cookie) {
		if err := decode(c.Name, c.Value, &c.Value); err != nil {
			c.Value = ""
		}
	}
}
