package literoute

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	http.CloseNotifier
	http.Pusher

	Naive() http.ResponseWriter

	IsHijacked() bool

	Writef(format string, a ...interface{}) (n int, err error)

	WriteString(s string) (n int, err error)

	StatusCode() int

	Written() int

	SetWritten(int)

	SetBeforeFlush(cb func())
	GetBeforeFlush() func()
	FlushResponse()

	BeginResponse(underline http.ResponseWriter)
	EndResponse()

	Clone() ResponseWriter

	WriteTo(ResponseWriter)

	Flusher() (http.Flusher, bool)

	CloseNotifier() (http.CloseNotifier, bool)
}

var rpool = sync.Pool{New: func() interface{} { return &responseWriter{} }}

func acquireResponseWriter() ResponseWriter {
	return rpool.Get().(*responseWriter)
}

func releaseResponseWriter(w ResponseWriter) {
	rpool.Put(w)
}

//func asResponseWriter(w http.ResponseWriter) ResponseWriter {
//	return &responseWriter{
//		ResponseWriter: w,
//		statusCode:     defaultStatusCode,
//		written:        NoWritten,
//		beforeFlush:    nil,
//	}
//}

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	written     int
	beforeFlush func()
}

var _ ResponseWriter = (*responseWriter)(nil)

const (
	defaultStatusCode = http.StatusOK
	NoWritten         = -1
	StatusCodeWritten = 0
)

func (w *responseWriter) Naive() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriter) BeginResponse(underline http.ResponseWriter) {
	w.beforeFlush = nil
	w.written = NoWritten
	w.statusCode = defaultStatusCode
	w.ResponseWriter = underline
}

func (w *responseWriter) EndResponse() {
	releaseResponseWriter(w)
}

func (w *responseWriter) SetWritten(n int) {
	if n >= NoWritten && n <= StatusCodeWritten {
		w.written = n
	}
}

func (w *responseWriter) Written() int {
	return w.written
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *responseWriter) tryWriteHeader() {
	if w.written == NoWritten {
		w.written = StatusCodeWritten
		w.ResponseWriter.WriteHeader(w.statusCode)
	}
}

func (w *responseWriter) IsHijacked() bool {
	_, err := w.ResponseWriter.Write(nil)
	return err == http.ErrHijacked
}

func (w *responseWriter) Write(contents []byte) (int, error) {
	w.tryWriteHeader()
	n, err := w.ResponseWriter.Write(contents)
	w.written += n
	return n, err
}

func (w *responseWriter) Writef(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w, format, a...)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.tryWriteHeader()
	n, err := io.WriteString(w.ResponseWriter, s)
	w.written += n
	return n, err
}

func (w *responseWriter) StatusCode() int {
	return w.statusCode
}

func (w *responseWriter) GetBeforeFlush() func() {
	return w.beforeFlush
}

func (w *responseWriter) SetBeforeFlush(cb func()) {
	w.beforeFlush = cb
}

func (w *responseWriter) FlushResponse() {
	if w.beforeFlush != nil {
		w.beforeFlush()
	}

	w.tryWriteHeader()
}

func (w *responseWriter) Clone() ResponseWriter {
	wc := &responseWriter{}
	wc.ResponseWriter = w.ResponseWriter
	wc.statusCode = w.statusCode
	wc.beforeFlush = w.beforeFlush
	wc.written = w.written
	return wc
}

func (w *responseWriter) WriteTo(to ResponseWriter) {
	if w.statusCode >= 400 {
		to.WriteHeader(w.statusCode)
	}
	if w.Header() != nil {
		for k, values := range w.Header() {
			for _, v := range values {
				if to.Header().Get(v) == "" {
					to.Header().Add(k, v)
				}
			}
		}
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, isHijacker := w.ResponseWriter.(http.Hijacker); isHijacker {
		w.written = StatusCodeWritten
		return h.Hijack()
	}

	return nil, nil, errors.New("hijack is not supported by this ResponseWriter")
}

func (w *responseWriter) Flusher() (http.Flusher, bool) {
	flusher, canFlush := w.ResponseWriter.(http.Flusher)
	return flusher, canFlush
}

func (w *responseWriter) Flush() {
	if flusher, ok := w.Flusher(); ok {
		flusher.Flush()
	}
}

var ErrPushNotSupported = errors.New("push feature is not supported by this ResponseWriter")

func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, isPusher := w.ResponseWriter.(http.Pusher); isPusher {
		err := pusher.Push(target, opts)
		if err != nil && err.Error() == http.ErrNotSupported.ErrorString {
			return ErrPushNotSupported
		}
		return err
	}
	return ErrPushNotSupported
}

func (w *responseWriter) CloseNotifier() (http.CloseNotifier, bool) {
	notifier, supportsCloseNotify := w.ResponseWriter.(http.CloseNotifier)
	return notifier, supportsCloseNotify
}

func (w *responseWriter) CloseNotify() <-chan bool {
	if notifier, ok := w.CloseNotifier(); ok {
		return notifier.CloseNotify()
	}

	ch := make(chan bool, 1)
	return ch
}
