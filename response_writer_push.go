package literoute

import (
	"fmt"
	"net/http"
	"sync"
)

var Pusher = func(ctx Context) {
	ctx.Push()
}

var prpool = sync.Pool{New: func() interface{} { return &ResponsePusher{} }}

func acquireResponsePusher() *ResponsePusher {
	return prpool.Get().(*ResponsePusher)
}

func releaseResponsePusher(w *ResponsePusher) {
	prpool.Put(w)
}

type ResponsePusher struct {
	ResponseWriter
	chunks  []byte
	headers http.Header
}

var _ ResponseWriter = (*ResponsePusher)(nil)

func (w *ResponsePusher) Naive() http.ResponseWriter {
	return w.ResponseWriter.Naive()
}

func (w *ResponsePusher) Begin(underline ResponseWriter) {
	w.ResponseWriter = underline
	w.headers = underline.Header()
	w.ResetBody()
}

func (w *ResponsePusher) EndResponse() {
	releaseResponsePusher(w)
	w.ResponseWriter.EndResponse()
}

func (w *ResponsePusher) Write(contents []byte) (int, error) {
	w.chunks = append(w.chunks, contents...)
	return len(contents), nil
}

func (w *ResponsePusher) Writef(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w, format, a...)
}

func (w *ResponsePusher) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

func (w *ResponsePusher) SetBody(b []byte) {
	w.chunks = b
}

func (w *ResponsePusher) SetBodyString(s string) {
	w.SetBody([]byte(s))
}

func (w *ResponsePusher) Body() []byte {
	return w.chunks
}

func (w *ResponsePusher) ResetBody() {
	w.chunks = w.chunks[0:0]
}

func (w *ResponsePusher) ResetHeaders() {
	w.headers = w.ResponseWriter.Header()
}

func (w *ResponsePusher) ClearHeaders() {
	w.headers = http.Header{}
	h := w.ResponseWriter.Header()
	for k := range h {
		h[k] = nil
	}
}

func (w *ResponsePusher) Reset() {
	w.ClearHeaders()
	w.WriteHeader(defaultStatusCode)
	w.ResetBody()
}

func (w *ResponsePusher) FlushResponse() {
	if w.headers != nil {
		h := w.ResponseWriter.Header()

		for k, values := range w.headers {
			h[k] = nil
			for i := range values {
				h.Add(k, values[i])
			}
		}
	}

	w.ResponseWriter.FlushResponse()

	if len(w.chunks) > 0 {
		_, _ = w.ResponseWriter.Write(w.chunks)
	}
}

func (w *ResponsePusher) Clone() ResponseWriter {
	wc := &ResponsePusher{}
	wc.headers = w.headers
	wc.chunks = w.chunks[0:]
	if resW, ok := w.ResponseWriter.(*responseWriter); ok {
		wc.ResponseWriter = &(*resW)
	} else {
		wc.ResponseWriter = w.ResponseWriter
	}
	return wc
}

func (w *ResponsePusher) WriteTo(res ResponseWriter) {
	if to, ok := res.(*ResponsePusher); ok {

		if statusCode := w.ResponseWriter.StatusCode(); statusCode == defaultStatusCode {
			to.WriteHeader(statusCode)
		}

		if beforeFlush := w.ResponseWriter.GetBeforeFlush(); beforeFlush != nil {
			if to.GetBeforeFlush() != nil {
				nextBeforeFlush := beforeFlush
				prevBeforeFlush := to.GetBeforeFlush()
				to.SetBeforeFlush(func() {
					prevBeforeFlush()
					nextBeforeFlush()
				})
			} else {
				to.SetBeforeFlush(w.ResponseWriter.GetBeforeFlush())
			}
		}

		if resW, ok := to.ResponseWriter.(*responseWriter); ok {
			if resW.Written() != StatusCodeWritten {
				resW.written = w.ResponseWriter.Written()
			}
		}

		if w.headers != nil {
			for k, values := range w.headers {
				for _, v := range values {
					if to.headers.Get(v) == "" {
						to.headers.Add(k, v)
					}
				}
			}
		}

		if len(w.chunks) > 0 {
			_, _ = to.Write(w.chunks)
		}
	}
}

func (w *ResponsePusher) Flush() {
	w.ResponseWriter.Flush()
	w.ResetBody()
}

func (w *ResponsePusher) Push(target string, opts *http.PushOptions) error {
	w.FlushResponse()
	err := w.ResponseWriter.Push(target, opts)
	w.ResetBody()
	w.ResetHeaders()

	return err
}
