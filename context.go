package literoute

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
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
	"sync"
	"sync/atomic"
	"time"
)

type Context interface {
	Mux() *LiteMux
	Request() (r *http.Request)
	ResponseWriter() ResponseWriter
	ResetResponseWriter(ResponseWriter)
	Method() string
	Path() string
	RequestPath(escape bool) string
	Host() string
	FullRequestURI() string
	RemoteAddr(headerNames ...string) string
	GetHeader(name string) string

	IsAjax() bool

	IsMobile() bool
	Header(name string, value string) Context

	ContentType(cType string) Context
	GetContentType() string
	GetContentTypeRequested() string
	GetContentLength() int64
	StatusCode(statusCode int) Context
	GetStatusCode() int

	AbsoluteURI(s string) string

	Redirect(urlToRedirect string, statusHeader ...int)

	Param(key string) string
	ParamInt(key string) (int, error)
	ParamInt32(key string) (int32, error)
	ParamInt64(key string) (int64, error)

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

	Succeed(v interface{})
	NotFound()
	Fail(v interface{})
	Invalid(v interface{})

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

	BeginRequest(w http.ResponseWriter, r *http.Request)
	End()
}

var DefaultJSONOptions = JSON{}

type JSON struct {
	StreamingJSON bool
	UnescapeHTML  bool
	Indent        string
	Prefix        string
}

var DefaultJSONPOptions = JSONP{}

type JSONP struct {
	Indent   string
	Callback string
}

var DefaultXMLOptions = XML{}

type XML struct {
	Indent string
	Prefix string
}

type Unmarshaller interface {
	Unmarshal(data []byte, outPtr interface{}) error
}

var _ Context = (*context)(nil)

var contextPool = sync.Pool{}

func acquireContext(mux *LiteMux, w http.ResponseWriter, r *http.Request) Context {
	var ctx Context
	v := contextPool.Get()
	if v == nil {
		ctx = newContext(mux)
	} else {
		ctx, _ = v.(Context)
	}
	ctx.BeginRequest(w, r)
	return ctx
}

func releaseContext(ctx Context) {
	ctx.End()
	contextPool.Put(ctx)
}

func newContext(mux *LiteMux) (ctx Context) {
	ctx = &context{
		id:  LastCapturedContextID(),
		mux: mux,
	}
	return
}

type context struct {
	id      uint64
	mux     *LiteMux
	writer  ResponseWriter
	request *http.Request
}

func (ctx *context) String() string {
	if ctx.id == 0 {
		forward := atomic.AddUint64(&lastCapturedContextID, 1)
		ctx.id = forward
	}

	return fmt.Sprintf("[%d] %s ▶ %s:%s",
		ctx.id, ctx.RemoteAddr(), ctx.Method(), ctx.Request().RequestURI)
}

func (ctx *context) Mux() *LiteMux {
	return ctx.mux
}

func (ctx *context) BeginRequest(w http.ResponseWriter, r *http.Request) () {
	ctx.writer = asResponseWriter(w)
	ctx.request = r
}

func (ctx *context) End() {
	ctx.writer.FlushResponse()
	ctx.writer.EndResponse()
}

var lastCapturedContextID uint64

func LastCapturedContextID() uint64 {
	return atomic.LoadUint64(&lastCapturedContextID)
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

func (ctx *context) RemoteAddr(headerNames ...string) string {

	for _, headerName := range headerNames {
		headerValue := ctx.GetHeader(headerName)
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

	addr := strings.TrimSpace(ctx.request.RemoteAddr)
	if addr != "" {
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

func (ctx *context) Header(name string, value string) Context {
	if value == "" {
		ctx.writer.Header().Del(name)
		return ctx
	}
	ctx.writer.Header().Add(name, value)
	return ctx
}

func (ctx *context) contentTypeOnce(cType string, charset string) {

	if cType != ContentBinaryHeaderValue {
		cType += "; charset=" + charset
	}
	ctx.writer.Header().Set(ContentTypeHeaderKey, cType)
}

func (ctx *context) ContentType(cType string) Context {
	if cType == "" {
		return ctx
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

	return ctx
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

func (ctx *context) StatusCode(statusCode int) Context {
	ctx.writer.WriteHeader(statusCode)
	return ctx
}

func (ctx *context) Succeed(v interface{}) {
	ctx.op(func() int {
		if status := ctx.Mux().getConfig().Status.Succeed; status == 0 {
			return http.StatusOK
		} else {
			return status
		}
	}(), v)
}

func (ctx *context) Fail(v interface{}) {
	ctx.op(func() int {
		if status := ctx.Mux().getConfig().Status.Fail; status == 0 {
			return http.StatusInternalServerError
		} else {
			return status
		}
	}(), v)
}

func (ctx *context) Invalid(v interface{}) {
	ctx.op(func() int {
		if status := ctx.Mux().getConfig().Status.InvalidRequest; status == 0 {
			return http.StatusBadRequest
		} else {
			return status
		}
	}(), v)
}

func (ctx *context) op(status int, v interface{}) {
	ctx.StatusCode(status)
	switch ctx.Mux().getConfig().BodyEncoder {
	case JsonBodyEncode:
		_, _ = ctx.JSON(v)
	case XmlBodyEncode:
		_, _ = ctx.XML(v)
	}
}

func (ctx *context) NotFound() {
	ctx.StatusCode(func() int {
		if status := ctx.Mux().getConfig().Status.NotFound; status == 0 {
			return http.StatusNotFound
		} else {
			return status
		}
	}())
}

func (ctx *context) GetStatusCode() int {
	return ctx.writer.StatusCode()
}

func (ctx *context) getAllParams() map[string]string {
	values, ok := ctx.request.Context().Value(contextKey).(map[string]string)
	if ok {
		return values
	}

	return map[string]string{}
}

func (ctx *context) Param(key string) string {
	return ctx.getAllParams()[key]
}

func (ctx *context) ParamInt(key string) (int, error) {
	n, err := strconv.Atoi(ctx.Param(key))
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (ctx *context) ParamInt32(key string) (int32, error) {
	n, err := strconv.ParseInt(ctx.Param(key), 10, 32)
	if err != nil {
		return -1, err
	}
	return int32(n), nil
}

func (ctx *context) ParamInt64(key string) (int64, error) {
	n, err := strconv.ParseInt(ctx.Param(key), 10, 64)
	if err != nil {
		return -1, err
	}
	return n, nil
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
	return GetForm(ctx.request, ctx.Mux().getConfig().PostMaxMemory, false)
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
	if err := ctx.request.ParseMultipartForm(ctx.Mux().getConfig().PostMaxMemory); err != nil {
		return nil, nil, err
	}

	return ctx.request.FormFile(key)
}

func (ctx *context) UploadFormFiles(destDirectory string, before ...func(Context, *multipart.FileHeader)) (n int64, err error) {
	err = ctx.request.ParseMultipartForm(ctx.Mux().getConfig().PostMaxMemory)
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
	return GetBody(ctx.request, false)
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

func (ctx *context) Writef(format string, a ...interface{}) (n int, err error) {
	return ctx.writer.Writef(format, a...)
}

func (ctx *context) WriteString(body string) (n int, err error) {
	return ctx.writer.WriteString(body)
}

func (ctx *context) SetLastModified(modifyTime time.Time) {
	if !IsZeroTime(modifyTime) {
		ctx.Header(LastModifiedHeaderKey, FormatTimeRFC3339Nano(modifyTime.UTC()))
	}
}

func (ctx *context) CheckIfModifiedSince(modifyTime time.Time) (bool, error) {
	if method := ctx.Method(); method != http.MethodGet && method != http.MethodHead {
		return false, fmt.Errorf("method: %w", ErrPreconditionFailed)
	}
	ims := ctx.GetHeader(IfModifiedSinceHeaderKey)
	if ims == "" || IsZeroTime(modifyTime) {
		return false, fmt.Errorf("zero time: %w", ErrPreconditionFailed)
	}
	t, err := ParseTimeRFC3339Nano(ctx, ims)
	if err != nil {
		return false, err
	}
	// sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if modifyTime.UTC().Before(t.Add(1 * time.Second)) {
		return false, nil
	}
	return true, nil
}

func (ctx *context) WriteNotModified() {
	h := ctx.ResponseWriter().Header()
	delete(h, ContentTypeHeaderKey)
	delete(h, ContentLengthHeaderKey)
	if h.Get(ETagHeaderKey) != "" {
		delete(h, LastModifiedHeaderKey)
	}
	ctx.StatusCode(http.StatusNotModified)
}

func (ctx *context) WriteWithExpiration(body []byte, modifyTime time.Time) (int, error) {
	if modified, err := ctx.CheckIfModifiedSince(modifyTime); !modified && err == nil {
		ctx.WriteNotModified()
		return 0, nil
	}

	ctx.SetLastModified(modifyTime)
	return ctx.writer.Write(body)
}

func (ctx *context) StreamWriter(writer func(w io.Writer) bool) {
	w := ctx.writer
	notifyClosed := w.CloseNotify()
	for {
		select {
		case <-notifyClosed:
			return
		default:
			shouldContinue := writer(w)
			w.Flush()
			if !shouldContinue {
				return
			}
		}
	}
}

func (ctx *context) ClientSupportsGzip() bool {
	if h := ctx.GetHeader(AcceptEncodingHeaderKey); h != "" {
		for _, v := range strings.Split(h, ";") {
			if strings.Contains(v, GzipHeaderValue) {
				return true
			}
		}
	}
	return false
}

func (ctx *context) WriteGzip(b []byte) (int, error) {
	if !ctx.ClientSupportsGzip() {
		return 0, ErrGzipNotSupported
	}

	return ctx.GzipResponseWriter().Write(b)
}

func (ctx *context) TryWriteGzip(b []byte) (int, error) {
	n, err := ctx.WriteGzip(b)
	if err != nil {
		if errors.Is(err, ErrGzipNotSupported) {
			return ctx.writer.Write(b)
		}
	}
	return n, err
}

func (ctx *context) GzipResponseWriter() *GzipResponseWriter {
	if gzipResWriter, ok := ctx.writer.(*GzipResponseWriter); ok {
		return gzipResWriter
	}
	gzipResWriter := AsGzipResponseWriter(ctx.writer)
	ctx.ResetResponseWriter(gzipResWriter)
	return gzipResWriter
}

func (ctx *context) ResetResponseWriter(newResponseWriter ResponseWriter) {
	ctx.writer = newResponseWriter
}

func (ctx *context) Gzip(enable bool) {
	if enable {
		if ctx.ClientSupportsGzip() {
			_ = ctx.GzipResponseWriter()
		}
	} else {
		if gzipResWriter, ok := ctx.writer.(*GzipResponseWriter); ok {
			gzipResWriter.Disable()
		}
	}
}

func (ctx *context) Binary(data []byte) (int, error) {
	ctx.ContentType(ContentBinaryHeaderValue)
	return ctx.Write(data)
}

func (ctx *context) Text(format string, args ...interface{}) (int, error) {
	ctx.ContentType(ContentTextHeaderValue)
	return ctx.Writef(format, args...)
}

func (ctx *context) HTML(format string, args ...interface{}) (int, error) {
	ctx.ContentType(ContentHTMLHeaderValue)
	return ctx.Writef(format, args...)
}

func WriteJSON(writer io.Writer, v interface{}, options JSON) (int, error) {
	var (
		result []byte
		err    error
	)

	if indent := options.Indent; indent != "" {
		marshalIndent := json.MarshalIndent

		result, err = marshalIndent(v, "", indent)
		result = append(result, newLineB...)
	} else {
		marshal := json.Marshal

		result, err = marshal(v)
	}

	if err != nil {
		return 0, err
	}

	if options.UnescapeHTML {
		result = bytes.Replace(result, ltHex, lt, -1)
		result = bytes.Replace(result, gtHex, gt, -1)
		result = bytes.Replace(result, andHex, and, -1)
	}

	if prefix := options.Prefix; prefix != "" {
		result = append([]byte(prefix), result...)
	}

	return writer.Write(result)
}

func (ctx *context) JSON(v interface{}, opts ...JSON) (n int, err error) {
	options := DefaultJSONOptions

	if len(opts) > 0 {
		options = opts[0]
	}

	ctx.ContentType(ContentJSONHeaderValue)

	if options.StreamingJSON {

		enc := json.NewEncoder(ctx.writer)
		enc.SetEscapeHTML(!options.UnescapeHTML)
		enc.SetIndent(options.Prefix, options.Indent)
		err = enc.Encode(v)

		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			return 0, err
		}
		return ctx.writer.Written(), err
	}

	n, err = WriteJSON(ctx.writer, v, options)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return 0, err
	}

	return n, err
}

func WriteJSONP(writer io.Writer, v interface{}, options JSONP) (int, error) {
	if callback := options.Callback; callback != "" {
		_, _ = writer.Write([]byte(callback + "("))
		defer func() {
			_, _ = writer.Write(finishCallbackB)
		}()
	}

	if indent := options.Indent; indent != "" {

		result, err := json.MarshalIndent(v, "", indent)
		if err != nil {
			return 0, err
		}
		result = append(result, newLineB...)
		return writer.Write(result)
	}

	result, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	return writer.Write(result)
}

func (ctx *context) JSONP(v interface{}, opts ...JSONP) (int, error) {
	options := DefaultJSONPOptions

	if len(opts) > 0 {
		options = opts[0]
	}

	ctx.ContentType(ContentJavascriptHeaderValue)

	n, err := WriteJSONP(ctx.writer, v, options)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return 0, err
	}

	return n, err
}

func (ctx *context) XML(v interface{}, opts ...XML) (int, error) {
	options := DefaultXMLOptions

	if len(opts) > 0 {
		options = opts[0]
	}

	ctx.ContentType(ContentXMLHeaderValue)

	n, err := WriteXML(ctx.writer, v, options)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return 0, err
	}

	return n, err
}

func (ctx *context) ServeContent(content io.ReadSeeker, filename string, modifyTime time.Time, gzipCompression bool) error {
	if modified, err := ctx.CheckIfModifiedSince(modifyTime); !modified && err == nil {
		ctx.WriteNotModified()
		return nil
	}

	if ctx.GetContentType() == "" {
		ctx.ContentType(filename)
	}

	ctx.SetLastModified(modifyTime)
	var out io.Writer
	if gzipCompression && ctx.ClientSupportsGzip() {
		addGzipHeaders(ctx.writer)

		gzipWriter := acquireGzipWriter(ctx.writer)
		defer releaseGzipWriter(gzipWriter)
		out = gzipWriter
	} else {
		out = ctx.writer
	}
	_, err := io.Copy(out, content)
	return err
}

func (ctx *context) ServeFile(filename string, gzipCompression bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("%d", http.StatusNotFound)
	}
	defer func() {
		_ = f.Close()
	}()
	fi, _ := f.Stat()
	if fi.IsDir() {
		return ctx.ServeFile(path.Join(filename, "index.html"), gzipCompression)
	}

	return ctx.ServeContent(f, fi.Name(), fi.ModTime(), gzipCompression)
}

func (ctx *context) SendFile(filename string, destinationName string) error {
	ctx.writer.Header().Set(ContentDispositionHeaderKey, "attachment;filename="+destinationName)
	return ctx.ServeFile(filename, false)
}

func (ctx *context) SetCookie(cookie *http.Cookie, options ...CookieOption) {
	for _, opt := range options {
		opt(cookie)
	}

	http.SetCookie(ctx.writer, cookie)
}

func (ctx *context) SetCookieKV(name, value string, options ...CookieOption) {
	c := &http.Cookie{}
	c.Path = "/"
	c.Name = name
	c.Value = url.QueryEscape(value)
	c.HttpOnly = true
	c.Expires = time.Now().Add(setCookieKVExpiration)
	c.MaxAge = int(setCookieKVExpiration.Seconds())
	ctx.SetCookie(c, options...)
}

func (ctx *context) GetCookie(name string, options ...CookieOption) string {
	cookie, err := ctx.request.Cookie(name)
	if err != nil {
		return ""
	}

	for _, opt := range options {
		opt(cookie)
	}

	value, _ := url.QueryUnescape(cookie.Value)
	return value
}

func (ctx *context) RemoveCookie(name string, options ...CookieOption) {
	c := &http.Cookie{}
	c.Name = name
	c.Value = ""
	c.Path = "/"
	c.HttpOnly = true
	exp := time.Now().Add(-time.Duration(1) * time.Minute)
	c.Expires = exp
	c.MaxAge = -1
	ctx.SetCookie(c, options...)
	ctx.request.Header.Set("Cookie", "")
}

func (ctx *context) VisitAllCookies(visitor func(name string, value string)) {
	for _, cookie := range ctx.request.Cookies() {
		visitor(cookie.Name, cookie.Value)
	}
}

func (ctx *context) MaxAge() int64 {
	header := ctx.GetHeader(CacheControlHeaderKey)
	if header == "" {
		return -1
	}
	m := maxAgeExp.FindStringSubmatch(header)
	if len(m) == 2 {
		if v, err := strconv.Atoi(m[1]); err == nil {
			return int64(v)
		}
	}
	return -1
}

func (ctx *context) Push() {
	if w, ok := ctx.writer.(*responseWriter); ok {
		recorder := acquireResponsePusher()
		recorder.Begin(w)
		ctx.ResetResponseWriter(recorder)
	}
}

func (ctx *context) Pusher() *ResponsePusher {
	ctx.Push()
	return ctx.writer.(*ResponsePusher)
}

func (ctx *context) IsPushing() (*ResponsePusher, bool) {
	rr, ok := ctx.writer.(*ResponsePusher)
	return rr, ok
}

type BodyDecoder interface {
	Decode(data []byte) error
}

type UnMarshallerFunc func(data []byte, outPtr interface{}) error

func (u UnMarshallerFunc) Unmarshal(data []byte, v interface{}) error {
	return u(data, v)
}
