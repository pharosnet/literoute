package literoute

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)


func newSimpleContext(rw http.ResponseWriter, r *http.Request) (ctx Context) {

	return
}

type Context interface {
	Context() context.Context
	Request() (r *http.Request)
	ResponseWriter() (w http.ResponseWriter)
	Method() string
	Path() string
	RequestPath(escape bool) string
	Host() string
	FullRequestURI() string
	RemoteAddr() string
	GetHeader(name string) string

	IsAjax() bool

	IsMobile() bool
	Header(name string, value string)

	ContentType(cType string)
	GetContentType() string
	GetContentTypeRequested() string
	GetContentLength() int64
	StatusCode(statusCode int)
	GetStatusCode() int

	AbsoluteURI(s string) string

	Redirect(urlToRedirect string, statusHeader ...int)

	URLParamExists(name string) bool
	URLParamDefault(name string, def string) string
	URLParam(name string) string
	URLParamTrim(name string) string
	URLParamEscape(name string) string
	URLParamInt(name string) (int, error)
	URLParamIntDefault(name string, def int) int
	URLParamInt32Default(name string, def int32) int32
	URLParamInt64(name string) (int64, error)
	URLParamInt64Default(name string, def int64) int64
	URLParamFloat64(name string) (float64, error)
	URLParamFloat64Default(name string, def float64) float64
	URLParamBool(name string) (bool, error)
	URLParams() map[string]string

	FormValueDefault(name string, def string) string
	FormValue(name string) string
	FormValues() map[string][]string
	PostValueDefault(name string, def string) string
	PostValue(name string) string
	PostValueTrim(name string) string
	PostValueInt(name string) (int, error)
	PostValueIntDefault(name string, def int) int
	PostValueInt64(name string) (int64, error)
	PostValueInt64Default(name string, def int64) int64
	PostValueFloat64(name string) (float64, error)
	PostValueFloat64Default(name string, def float64) float64
	PostValueBool(name string) (bool, error)
	PostValues(name string) []string
	FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	UploadFormFiles(destDirectory string, before ...func(Context, *multipart.FileHeader)) (n int64, err error)


	NotFound()


	SetMaxRequestBodySize(limitOverBytes int64)


	GetBody() ([]byte, error)

	UnmarshalBody(outPtr interface{}, unMarshaller Unmarshaller) error
	ReadJSON(jsonObjectPtr interface{}) error

	ReadForm(formObject interface{}) error

	ReadQuery(ptr interface{}) error

	Write(body []byte) (int, error)

	Writef(format string, args ...interface{}) (int, error)

	WriteString(body string) (int, error)


	SetLastModified(modifyTime time.Time)

	CheckIfModifiedSince(modifyTime time.Time) (bool, error)

	WriteNotModified()

	WriteWithExpiration(body []byte, modifyTime time.Time) (int, error)

	StreamWriter(writer func(w io.Writer) bool)


	ClientSupportsGzip() bool

	WriteGzip(b []byte) (int, error)

	TryWriteGzip(b []byte) (int, error)

	GzipResponseWriter() *GzipResponseWriter

	Gzip(enable bool)



	Binary(data []byte) (int, error)
	Text(format string, args ...interface{}) (int, error)
	HTML(format string, args ...interface{}) (int, error)
	JSON(v interface{}, options ...JSON) (int, error)
	JSONP(v interface{}, options ...JSONP) (int, error)
	XML(v interface{}, options ...XML) (int, error)





	ServeContent(content io.ReadSeeker, filename string, modifyTime time.Time, gzipCompression bool) error

	ServeFile(filename string, gzipCompression bool) error

	SendFile(filename string, destinationName string) error


	SetCookie(cookie *http.Cookie, options ...CookieOption)

	SetCookieKV(name, value string, options ...CookieOption)

	GetCookie(name string, options ...CookieOption) string

	RemoveCookie(name string, options ...CookieOption)
	VisitAllCookies(visitor func(name string, value string))


	MaxAge() int64


	Record()

	Recorder() *ResponseRecorder

	IsRecording() (*ResponseRecorder, bool)

}

type JSON struct {
	StreamingJSON bool
	UnescapeHTML bool
	Indent       string
	Prefix       string
}

type JSONP struct {
	Indent   string
	Callback string
}

type XML struct {
	Indent string
	Prefix string
}

type Unmarshaller interface {
Unmarshal(data []byte, outPtr interface{}) error
}