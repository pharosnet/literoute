package literoute

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/pharosnet/literoute/schema"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func newSimpleContext(rw http.ResponseWriter, r *http.Request) (ctx Context) {

	return
}

type Context interface {
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

	Push()

	Pusher() *ResponsePusher

	IsPushing() (*ResponsePusher, bool)
}

type JSON struct {
	StreamingJSON bool
	UnescapeHTML  bool
	Indent        string
	Prefix        string
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

var _ Context = (*context)(nil)

var LimitRequestBodySize = func(maxRequestBodySizeBytes int64) HandleFunc {
	return func(ctx Context) {
		ctx.SetMaxRequestBodySize(maxRequestBodySizeBytes)
	}
}

var Gzip = func(ctx Context) {
	ctx.Gzip(true)
}

type Map = map[string]interface{}

type context struct {
	id uint64
	writer ResponseWriter
	request *http.Request
	currentRouteName string

	params RequestParams  // url named parameters.
	values memstore.Store // generic storage, middleware communication.

	handlers Handlers
	// the current position of the handler's chain
	currentHandlerIndex int
}

func (ctx *context) ResponseWriter() ResponseWriter {
	return ctx.writer
}

func (ctx *context) Request() *http.Request {
	return ctx.request
}

func (ctx *context) Method() string {
	return ctx.request.Method
}

func (ctx *context) Path() string {
	return ctx.RequestPath(false)
}

func (ctx *context) RequestPath(escape bool) string {
	if escape {
		return ctx.request.URL.EscapedPath()
	}
	return ctx.request.URL.Path
}

func (ctx *context) Host() string {
	return GetHost(ctx.request)
}

func (ctx *context) FullRequestURI() string {
	return ctx.AbsoluteURI(ctx.Path())
}

func (ctx *context) RemoteAddr() string {
	remoteHeaders := ctx.Application().ConfigurationReadOnly().GetRemoteAddrHeaders()

	for headerName, enabled := range remoteHeaders {
		if enabled {
			headerValue := ctx.GetHeader(headerName)
			// exception needed for 'X-Forwarded-For' only , if enabled.
			if headerName == xForwardedForHeaderKey {
				idx := strings.IndexByte(headerValue, ',')
				if idx >= 0 {
					headerValue = headerValue[0:idx]
				}
			}

			realIP := strings.TrimSpace(headerValue)
			if realIP != "" {
				return realIP
			}
		}
	}

	addr := strings.TrimSpace(ctx.request.RemoteAddr)
	if addr != "" {
		// if addr has port use the net.SplitHostPort otherwise(error occurs) take as it is
		if ip, _, err := net.SplitHostPort(addr); err == nil {
			return ip
		}
	}

	return addr
}

func (ctx *context) AbsoluteURI(s string) string {
	if s == "" {
		return ""
	}

	if s[0] == '/' {
		scheme := ctx.request.URL.Scheme
		if scheme == "" {
			if ctx.request.TLS != nil {
				scheme = "https:"
			} else {
				scheme = "http:"
			}
		}

		host := ctx.Host()

		return scheme + "//" + host + path.Clean(s)
	}

	if u, err := url.Parse(s); err == nil {
		r := ctx.request

		if u.Scheme == "" && u.Host == "" {
			oldpath := r.URL.Path
			if oldpath == "" {
				oldpath = "/"
			}

			if s == "" || s[0] != '/' {
				olddir, _ := path.Split(oldpath)
				s = olddir + s
			}

			var query string
			if i := strings.Index(s, "?"); i != -1 {
				s, query = s[:i], s[i:]
			}

			// clean up but preserve trailing slash
			trailing := strings.HasSuffix(s, "/")
			s = path.Clean(s)
			if trailing && !strings.HasSuffix(s, "/") {
				s += "/"
			}
			s += query
		}
	}

	return s
}


func (ctx *context) IsAjax() bool {
	return ctx.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

func (ctx *context) IsMobile() bool {
	s := ctx.GetHeader("User-Agent")
	return isMobileRegex.MatchString(s)
}

func (ctx *context) Header(name string, value string) {
	if value == "" {
		ctx.writer.Header().Del(name)
		return
	}
	ctx.writer.Header().Add(name, value)
}

func (ctx *context) contentTypeOnce(cType string, charset string) {

	if cType != ContentBinaryHeaderValue {
		cType += "; charset=" + charset
	}
	ctx.writer.Header().Set(ContentTypeHeaderKey, cType)
}

func (ctx *context) ContentType(cType string) {
	if cType == "" {
		return
	}

	if strings.Contains(cType, ".") {
		ext := filepath.Ext(cType)
		cType = mime.TypeByExtension(ext)
	}
	if !strings.Contains(cType, "charset") {
		if cType != ContentBinaryHeaderValue {
			cType += "; charset=utf-8"
		}
	}

	ctx.writer.Header().Set(ContentTypeHeaderKey, cType)
}

func (ctx *context) GetContentType() string {
	return ctx.writer.Header().Get(ContentTypeHeaderKey)
}

func (ctx *context) GetContentTypeRequested() string {
	return ctx.GetHeader(ContentTypeHeaderKey)
}

func (ctx *context) GetContentLength() int64 {
	if v := ctx.GetHeader(ContentLengthHeaderKey); v != "" {
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	}
	return 0
}

func (ctx *context) StatusCode(statusCode int) {
	ctx.writer.WriteHeader(statusCode)
}

func (ctx *context) NotFound() {
	ctx.StatusCode(http.StatusNotFound)
}

func (ctx *context) GetStatusCode() int {
	return ctx.writer.StatusCode()
}

func (ctx *context) URLParamExists(name string) bool {
	if q := ctx.request.URL.Query(); q != nil {
		_, exists := q[name]
		return exists
	}

	return false
}

func (ctx *context) URLParamDefault(name string, def string) string {
	if v := ctx.request.URL.Query().Get(name); v != "" {
		return v
	}

	return def
}

func (ctx *context) URLParam(name string) string {
	return ctx.URLParamDefault(name, "")
}

func (ctx *context) URLParamTrim(name string) string {
	return strings.TrimSpace(ctx.URLParam(name))
}

func (ctx *context) URLParamEscape(name string) string {
	return DecodeQuery(ctx.URLParam(name))
}

func (ctx *context) URLParamInt(name string) (int, error) {
	if v := ctx.URLParam(name); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return -1, err
		}
		return n, nil
	}

	return -1, ErrNotFound
}

func (ctx *context) URLParamIntDefault(name string, def int) int {
	v, err := ctx.URLParamInt(name)
	if err != nil {
		return def
	}

	return v
}

func (ctx *context) URLParamInt32Default(name string, def int32) int32 {
	if v := ctx.URLParam(name); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return def
		}

		return int32(n)
	}

	return def
}

func (ctx *context) URLParamInt64(name string) (int64, error) {
	if v := ctx.URLParam(name); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return -1, err
		}
		return n, nil
	}

	return -1, ErrNotFound
}

func (ctx *context) URLParamInt64Default(name string, def int64) int64 {
	v, err := ctx.URLParamInt64(name)
	if err != nil {
		return def
	}

	return v
}

func (ctx *context) URLParamFloat64(name string) (float64, error) {
	if v := ctx.URLParam(name); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return -1, err
		}
		return n, nil
	}

	return -1, ErrNotFound
}

func (ctx *context) URLParamFloat64Default(name string, def float64) float64 {
	v, err := ctx.URLParamFloat64(name)
	if err != nil {
		return def
	}

	return v
}

func (ctx *context) URLParamBool(name string) (bool, error) {
	return strconv.ParseBool(ctx.URLParam(name))
}

func (ctx *context) URLParams() map[string]string {
	values := map[string]string{}

	q := ctx.request.URL.Query()
	if q != nil {
		for k, v := range q {
			values[k] = strings.Join(v, ",")
		}
	}

	return values
}



func (ctx *context) FormValue(name string) string {
	return ctx.FormValueDefault(name, "")
}

func (ctx *context) FormValues() map[string][]string {
	form, _ := ctx.form()
	return form
}

func (ctx *context) form() (form map[string][]string, found bool) {
	return GetForm(ctx.request, ctx.Application().ConfigurationReadOnly().GetPostMaxMemory(), ctx.Application().ConfigurationReadOnly().GetDisableBodyConsumptionOnUnmarshal())
}

func (ctx *context) PostValueDefault(name string, def string) string {
	ctx.form()
	if v := ctx.request.PostForm[name]; len(v) > 0 {
		return v[0]
	}
	return def
}

func (ctx *context) PostValue(name string) string {
	return ctx.PostValueDefault(name, "")
}

func (ctx *context) PostValueTrim(name string) string {
	return strings.TrimSpace(ctx.PostValue(name))
}

func (ctx *context) PostValueInt(name string) (int, error) {
	v := ctx.PostValue(name)
	if v == "" {
		return -1, ErrNotFound
	}
	return strconv.Atoi(v)
}

func (ctx *context) PostValueIntDefault(name string, def int) int {
	if v, err := ctx.PostValueInt(name); err == nil {
		return v
	}

	return def
}

func (ctx *context) PostValueInt64(name string) (int64, error) {
	v := ctx.PostValue(name)
	if v == "" {
		return -1, ErrNotFound
	}
	return strconv.ParseInt(v, 10, 64)
}

func (ctx *context) PostValueInt64Default(name string, def int64) int64 {
	if v, err := ctx.PostValueInt64(name); err == nil {
		return v
	}

	return def
}

func (ctx *context) PostValueFloat64(name string) (float64, error) {
	v := ctx.PostValue(name)
	if v == "" {
		return -1, ErrNotFound
	}
	return strconv.ParseFloat(v, 64)
}

func (ctx *context) PostValueFloat64Default(name string, def float64) float64 {
	if v, err := ctx.PostValueFloat64(name); err == nil {
		return v
	}

	return def
}

func (ctx *context) PostValueBool(name string) (bool, error) {
	v := ctx.PostValue(name)
	if v == "" {
		return false, ErrNotFound
	}

	return strconv.ParseBool(v)
}

func (ctx *context) PostValues(name string) []string {
	ctx.form()
	return ctx.request.PostForm[name]
}

func (ctx *context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	// we don't have access to see if the request is body stream
	// and then the ParseMultipartForm can be useless
	// here but do it in order to apply the post limit,
	// the internal request.FormFile will not do it if that's filled
	// and it's not a stream body.
	if err := ctx.request.ParseMultipartForm(ctx.Application().ConfigurationReadOnly().GetPostMaxMemory()); err != nil {
		return nil, nil, err
	}

	return ctx.request.FormFile(key)
}

func (ctx *context) UploadFormFiles(destDirectory string, before ...func(Context, *multipart.FileHeader)) (n int64, err error) {
	err = ctx.request.ParseMultipartForm(ctx.Application().ConfigurationReadOnly().GetPostMaxMemory())
	if err != nil {
		return 0, err
	}

	if ctx.request.MultipartForm != nil {
		if fhs := ctx.request.MultipartForm.File; fhs != nil {
			for _, files := range fhs {
				for _, file := range files {

					for _, b := range before {
						b(ctx, file)
					}

					n0, err0 := uploadTo(file, destDirectory)
					if err0 != nil {
						return 0, err0
					}
					n += n0
				}
			}
			return n, nil
		}
	}

	return 0, http.ErrMissingFile
}

func uploadTo(fh *multipart.FileHeader, destDirectory string) (int64, error) {
	src, err := fh.Open()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = src.Close()
	}()

	out, err := os.OpenFile(filepath.Join(destDirectory, fh.Filename),
		os.O_WRONLY|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = out.Close()
	}()

	return io.Copy(out, src)
}

func (ctx *context) Redirect(urlToRedirect string, statusHeader ...int) {
	status := ctx.GetStatusCode()
	if status < 300 {
		status = 0
	}

	if len(statusHeader) > 0 {
		if s := statusHeader[0]; s > 0 {
			status = s
		}
	}
	if status == 0 {
		status = http.StatusFound
	}

	http.Redirect(ctx.writer, ctx.request, urlToRedirect, status)
}

func (ctx *context) SetMaxRequestBodySize(limitOverBytes int64) {
	ctx.request.Body = http.MaxBytesReader(ctx.writer, ctx.request.Body, limitOverBytes)
}

func (ctx *context) GetBody() ([]byte, error) {
	return GetBody(ctx.request, ctx.Application().ConfigurationReadOnly().GetDisableBodyConsumptionOnUnmarshal())
}

func (ctx *context) UnmarshalBody(outPtr interface{}, unMarshaller Unmarshaller) error {
	if ctx.request.Body == nil {
		return fmt.Errorf("unmarshal: empty body: %w", ErrNotFound)
	}

	rawData, err := ctx.GetBody()
	if err != nil {
		return err
	}

	if decoder, isDecoder := outPtr.(BodyDecoder); isDecoder {
		return decoder.Decode(rawData)
	}

	return unMarshaller.Unmarshal(rawData, outPtr)
}

//func (ctx *context) shouldOptimize() bool {
//	return ctx.Application().ConfigurationReadOnly().GetEnableOptimizations()
//}

func (ctx *context) ReadJSON(outPtr interface{}) error {
	return ctx.UnmarshalBody(outPtr, UnMarshallerFunc(json.Unmarshal))
}

func (ctx *context) ReadXML(outPtr interface{}) error {
	return ctx.UnmarshalBody(outPtr, UnMarshallerFunc(xml.Unmarshal))
}


func (ctx *context) ReadForm(formObject interface{}) error {
	values := ctx.FormValues()
	if len(values) == 0 {
		return nil
	}

	return schema.DecodeForm(values, formObject)
}

func (ctx *context) ReadQuery(ptr interface{}) error {
	values := ctx.request.URL.Query()
	if len(values) == 0 {
		return nil
	}

	return schema.DecodeQuery(values, ptr)
}

func (ctx *context) Write(rawBody []byte) (int, error) {
	return ctx.writer.Write(rawBody)
}

func (ctx *context) WriteString(body string) (n int, err error) {
	return ctx.writer.WriteString(body)
}

func (ctx *context) SetLastModified(modtime time.Time) {
	if !IsZeroTime(modtime) {
		ctx.Header(LastModifiedHeaderKey, FormatTimeRFC3339Nano(ctx, modtime.UTC()))
	}
}

