package literoute

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
	ContentJSONProblemHeaderValue   = "application/problem+json"
	ContentXMLProblemHeaderValue    = "application/problem+xml"
	ContentJavascriptHeaderValue    = "application/javascript"
	ContentTextHeaderValue          = "text/plain"
	ContentXMLHeaderValue           = "text/xml"
	ContentXMLUnreadableHeaderValue = "application/xml"
	ContentMarkdownHeaderValue      = "text/markdown"
	ContentYAMLHeaderValue          = "application/x-yaml"
	ContentFormHeaderValue          = "application/x-www-form-urlencoded"
	ContentFormMultipartHeaderValue = "multipart/form-data"
)

var (
	newLineB = []byte("\n")
	ltHex = []byte("\\u003c")
	lt    = []byte("<")

	gtHex = []byte("\\u003e")
	gt    = []byte(">")

	andHex = []byte("\\u0026")
	and    = []byte("&")
)