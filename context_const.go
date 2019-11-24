package literoute

import (
	"errors"
	"regexp"
)

const (
	ContentTypeHeaderKey            = "Content-Type"
	LastModifiedHeaderKey           = "Last-Modified"
	IfModifiedSinceHeaderKey        = "If-Modified-Since"
	CacheControlHeaderKey           = "Cache-Control"
	ETagHeaderKey                   = "ETag"
	ContentDispositionHeaderKey     = "Content-Disposition"
	ContentLengthHeaderKey          = "Content-Length"
	ContentEncodingHeaderKey        = "Content-Encoding"
	GzipHeaderValue                 = "gzip"
	AcceptEncodingHeaderKey         = "Accept-Encoding"
	VaryHeaderKey                   = "Vary"
	ContentBinaryHeaderValue        = "application/octet-stream"
	ContentHTMLHeaderValue          = "text/html"
	ContentJSONHeaderValue          = "application/json"
	ContentJavascriptHeaderValue    = "application/javascript"
	ContentTextHeaderValue          = "text/plain"
	ContentXMLHeaderValue           = "text/xml"
	ContentXMLUnreadableHeaderValue = "application/xml"
	ContentFormHeaderValue          = "application/x-www-form-urlencoded"
	ContentFormMultipartHeaderValue = "multipart/form-data"

	xForwardedForHeaderKey = "X-Forwarded-For"
	XRealIp                = "X-Real-Ip"
	CfConnectingIp         = "CF-Connecting-IP"

	DefaultPostMaxMemory = 32 << 20 // 32MB
)

var (
	newLineB = []byte("\n")
	ltHex    = []byte("\\u003c")
	lt       = []byte("<")

	gtHex = []byte("\\u003e")
	gt    = []byte(">")

	andHex = []byte("\\u0026")
	and    = []byte("&")

	finishCallbackB = []byte(");")

	isMobileRegex = regexp.MustCompile(`(?i)(android|avantgo|blackberry|bolt|boost|cricket|docomo|fone|hiptop|mini|mobi|palm|phone|pie|tablet|up\.browser|up\.link|webos|wos)`)
	maxAgeExp     = regexp.MustCompile(`maxage=(\d+)`)
)

var (
	ErrNotFound           = errors.New("not found")
	ErrPreconditionFailed = errors.New("precondition failed")
	ErrGzipNotSupported   = errors.New("client does not support gzip compression")
)
