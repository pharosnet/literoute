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


	SetLastModified(modtime time.Time)

	CheckIfModifiedSince(modtime time.Time) (bool, error)

	WriteNotModified()

	WriteWithExpiration(body []byte, modtime time.Time) (int, error)

	StreamWriter(writer func(w io.Writer) bool)


	ClientSupportsGzip() bool

	WriteGzip(b []byte) (int, error)

	TryWriteGzip(b []byte) (int, error)

	GzipResponseWriter() *GzipResponseWriter

	Gzip(enable bool)

	//  +------------------------------------------------------------+
	//  | Rich Body Content Writers/Renderers                        |
	//  +------------------------------------------------------------+

	// ViewLayout sets the "layout" option if and when .View
	// is being called afterwards, in the same request.
	// Useful when need to set or/and change a layout based on the previous handlers in the chain.
	//
	// Note that the 'layoutTmplFile' argument can be set to iris.NoLayout || view.NoLayout
	// to disable the layout for a specific view render action,
	// it disables the engine's configuration's layout property.
	//
	// Look .ViewData and .View too.
	//
	// Example: https://github.com/kataras/iris/tree/master/_examples/view/context-view-data/
	ViewLayout(layoutTmplFile string)
	// ViewData saves one or more key-value pair in order to be passed if and when .View
	// is being called afterwards, in the same request.
	// Useful when need to set or/and change template data from previous hanadlers in the chain.
	//
	// If .View's "binding" argument is not nil and it's not a type of map
	// then these data are being ignored, binding has the priority, so the main route's handler can still decide.
	// If binding is a map or context.Map then these data are being added to the view data
	// and passed to the template.
	//
	// After .View, the data are not destroyed, in order to be re-used if needed (again, in the same request as everything else),
	// to clear the view data, developers can call:
	// ctx.Set(ctx.Application().ConfigurationReadOnly().GetViewDataContextKey(), nil)
	//
	// If 'key' is empty then the value is added as it's (struct or map) and developer is unable to add other value.
	//
	// Look .ViewLayout and .View too.
	//
	// Example: https://github.com/kataras/iris/tree/master/_examples/view/context-view-data/
	ViewData(key string, value interface{})
	// GetViewData returns the values registered by `context#ViewData`.
	// The return value is `map[string]interface{}`, this means that
	// if a custom struct registered to ViewData then this function
	// will try to parse it to map, if failed then the return value is nil
	// A check for nil is always a good practise if different
	// kind of values or no data are registered via `ViewData`.
	//
	// Similarly to `viewData := ctx.Values().Get("iris.viewData")` or
	// `viewData := ctx.Values().Get(ctx.Application().ConfigurationReadOnly().GetViewDataContextKey())`.
	GetViewData() map[string]interface{}
	// View renders a template based on the registered view engine(s).
	// First argument accepts the filename, relative to the view engine's Directory and Extension,
	// i.e: if directory is "./templates" and want to render the "./templates/users/index.html"
	// then you pass the "users/index.html" as the filename argument.
	//
	// The second optional argument can receive a single "view model"
	// that will be binded to the view template if it's not nil,
	// otherwise it will check for previous view data stored by the `ViewData`
	// even if stored at any previous handler(middleware) for the same request.
	//
	// Look .ViewData` and .ViewLayout too.
	//
	// Examples: https://github.com/kataras/iris/tree/master/_examples/view
	View(filename string, optionalViewModel ...interface{}) error

	// Binary writes out the raw bytes as binary data.
	Binary(data []byte) (int, error)
	// Text writes out a string as plain text.
	Text(format string, args ...interface{}) (int, error)
	// HTML writes out a string as text/html.
	HTML(format string, args ...interface{}) (int, error)
	// JSON marshals the given interface object and writes the JSON response.
	JSON(v interface{}, options ...JSON) (int, error)
	// JSONP marshals the given interface object and writes the JSON response.
	JSONP(v interface{}, options ...JSONP) (int, error)
	// XML marshals the given interface object and writes the XML response.
	// To render maps as XML see the `XMLMap` package-level function.
	XML(v interface{}, options ...XML) (int, error)
	// Problem writes a JSON or XML problem response.
	// Order of Problem fields are not always rendered the same.
	//
	// Behaves exactly like `Context.JSON`
	// but with default ProblemOptions.JSON indent of " " and
	// a response content type of "application/problem+json" instead.
	//
	// Use the options.RenderXML and XML fields to change this behavior and
	// send a response of content type "application/problem+xml" instead.
	//
	// Read more at: https://github.com/kataras/iris/wiki/Routing-error-handlers
	Problem(v interface{}, opts ...ProblemOptions) (int, error)
	// Markdown parses the markdown to html and renders its result to the client.
	Markdown(markdownB []byte, options ...Markdown) (int, error)
	// YAML parses the "v" using the yaml parser and renders its result to the client.
	YAML(v interface{}) (int, error)

	//  +-----------------------------------------------------------------------+
	//  | Content Νegotiation                                                   |
	//  | https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation |                                       |
	//  +-----------------------------------------------------------------------+

	// Negotiation creates once and returns the negotiation builder
	// to build server-side available content for specific mime type(s)
	// and charset(s).
	//
	// See `Negotiate` method too.
	Negotiation() *NegotiationBuilder
	// Negotiate used for serving different representations of a resource at the same URI.
	//
	// The "v" can be a single `N` struct value.
	// The "v" can be any value completes the `ContentSelector` interface.
	// The "v" can be any value completes the `ContentNegotiator` interface.
	// The "v" can be any value of struct(JSON, JSONP, XML, YAML) or
	// string(TEXT, HTML) or []byte(Markdown, Binary) or []byte with any matched mime type.
	//
	// If the "v" is nil, the `Context.Negotitation()` builder's
	// content will be used instead, otherwise "v" overrides builder's content
	// (server mime types are still retrieved by its registered, supported, mime list)
	//
	// Set mime type priorities by `Negotiation().JSON().XML().HTML()...`.
	// Set charset priorities by `Negotiation().Charset(...)`.
	// Set encoding algorithm priorities by `Negotiation().Encoding(...)`.
	// Modify the accepted by
	// `Negotiation().Accept./Override()/.XML().JSON().Charset(...).Encoding(...)...`.
	//
	// It returns `ErrContentNotSupported` when not matched mime type(s).
	//
	// Resources:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Charset
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Encoding
	//
	// Supports the above without quality values.
	//
	// Read more at: https://github.com/kataras/iris/wiki/Content-negotiation
	Negotiate(v interface{}) (int, error)

	//  +------------------------------------------------------------+
	//  | Serve files                                                |
	//  +------------------------------------------------------------+

	// ServeContent serves content, headers are autoset
	// receives three parameters, it's low-level function, instead you can use .ServeFile(string,bool)/SendFile(string,string)
	//
	//
	// You can define your own "Content-Type" with `context#ContentType`, before this function call.
	//
	// This function doesn't support resuming (by range),
	// use ctx.SendFile or router's `HandleDir` instead.
	ServeContent(content io.ReadSeeker, filename string, modtime time.Time, gzipCompression bool) error
	// ServeFile serves a file (to send a file, a zip for example to the client you should use the `SendFile` instead)
	// receives two parameters
	// filename/path (string)
	// gzipCompression (bool)
	//
	// You can define your own "Content-Type" with `context#ContentType`, before this function call.
	//
	// This function doesn't support resuming (by range),
	// use ctx.SendFile or router's `HandleDir` instead.
	//
	// Use it when you want to serve dynamic files to the client.
	ServeFile(filename string, gzipCompression bool) error
	// SendFile sends file for force-download to the client
	//
	// Use this instead of ServeFile to 'force-download' bigger files to the client.
	SendFile(filename string, destinationName string) error

	//  +------------------------------------------------------------+
	//  | Cookies                                                    |
	//  +------------------------------------------------------------+

	// SetCookie adds a cookie.
	// Use of the "options" is not required, they can be used to amend the "cookie".
	//
	// Example: https://github.com/kataras/iris/tree/master/_examples/cookies/basic
	SetCookie(cookie *http.Cookie, options ...CookieOption)
	// SetCookieKV adds a cookie, requires the name(string) and the value(string).
	//
	// By default it expires at 2 hours and it's added to the root path,
	// use the `CookieExpires` and `CookiePath` to modify them.
	// Alternatively: ctx.SetCookie(&http.Cookie{...})
	//
	// If you want to set custom the path:
	// ctx.SetCookieKV(name, value, iris.CookiePath("/custom/path/cookie/will/be/stored"))
	//
	// If you want to be visible only to current request path:
	// ctx.SetCookieKV(name, value, iris.CookieCleanPath/iris.CookiePath(""))
	// More:
	//                              iris.CookieExpires(time.Duration)
	//                              iris.CookieHTTPOnly(false)
	//
	// Example: https://github.com/kataras/iris/tree/master/_examples/cookies/basic
	SetCookieKV(name, value string, options ...CookieOption)
	// GetCookie returns cookie's value by its name
	// returns empty string if nothing was found.
	//
	// If you want more than the value then:
	// cookie, err := ctx.Request().Cookie("name")
	//
	// Example: https://github.com/kataras/iris/tree/master/_examples/cookies/basic
	GetCookie(name string, options ...CookieOption) string
	// RemoveCookie deletes a cookie by its name and path = "/".
	// Tip: change the cookie's path to the current one by: RemoveCookie("name", iris.CookieCleanPath)
	//
	// Example: https://github.com/kataras/iris/tree/master/_examples/cookies/basic
	RemoveCookie(name string, options ...CookieOption)
	// VisitAllCookies accepts a visitor function which is called
	// on each (request's) cookies' name and value.
	VisitAllCookies(visitor func(name string, value string))

	// MaxAge returns the "cache-control" request header's value
	// seconds as int64
	// if header not found or parse failed then it returns -1.
	MaxAge() int64

	//  +------------------------------------------------------------+
	//  | Advanced: Response Recorder and Transactions               |
	//  +------------------------------------------------------------+

	// Record transforms the context's basic and direct responseWriter to a ResponseRecorder
	// which can be used to reset the body, reset headers, get the body,
	// get & set the status code at any time and more.
	Record()
	// Recorder returns the context's ResponseRecorder
	// if not recording then it starts recording and returns the new context's ResponseRecorder
	Recorder() *ResponseRecorder
	// IsRecording returns the response recorder and a true value
	// when the response writer is recording the status code, body, headers and so on,
	// else returns nil and false.
	IsRecording() (*ResponseRecorder, bool)

}


type Unmarshaller interface {
Unmarshal(data []byte, outPtr interface{}) error
}