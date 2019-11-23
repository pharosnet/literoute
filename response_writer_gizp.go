package literoute

import (
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

type compressionPool struct {
	sync.Pool
	Level int
}

var gzipPool = &compressionPool{Level: -1}

func acquireGzipWriter(w io.Writer) *gzip.Writer {
	v := gzipPool.Get()
	if v == nil {
		gzipWriter, err := gzip.NewWriterLevel(w, gzipPool.Level)
		if err != nil {
			return nil
		}
		return gzipWriter
	}
	gzipWriter := v.(*gzip.Writer)
	gzipWriter.Reset(w)
	return gzipWriter
}

func releaseGzipWriter(gzipWriter *gzip.Writer) {
	_ = gzipWriter.Close()
	gzipPool.Put(gzipWriter)
}

func writeGzip(w io.Writer, b []byte) (int, error) {
	gzipWriter := acquireGzipWriter(w)
	n, err := gzipWriter.Write(b)
	if err != nil {
		releaseGzipWriter(gzipWriter)
		return -1, err
	}
	err = gzipWriter.Flush()
	releaseGzipWriter(gzipWriter)
	return n, err
}

var gzpool = sync.Pool{New: func() interface{} { return &GzipResponseWriter{} }}

func acquireGzipResponseWriter() *GzipResponseWriter {
	w := gzpool.Get().(*GzipResponseWriter)
	return w
}

func releaseGzipResponseWriter(w *GzipResponseWriter) {
	gzpool.Put(w)
}

type GzipResponseWriter struct {
	ResponseWriter
	chunks   []byte
	disabled bool
}

var _ ResponseWriter = (*GzipResponseWriter)(nil)

func (w *GzipResponseWriter) BeginGzipResponse(underline ResponseWriter) {
	w.ResponseWriter = underline

	w.chunks = w.chunks[0:0]
	w.disabled = false
}

func (w *GzipResponseWriter) EndResponse() {
	releaseGzipResponseWriter(w)
	w.ResponseWriter.EndResponse()
}

func (w *GzipResponseWriter) Write(contents []byte) (int, error) {
	w.chunks = append(w.chunks, contents...)
	return len(contents), nil
}

func (w *GzipResponseWriter) Writef(format string, a ...interface{}) (n int, err error) {
	n, err = fmt.Fprintf(w, format, a...)
	if err == nil {
		if w.ResponseWriter.Header()[ContentTypeHeaderKey] == nil {
			w.ResponseWriter.Header().Set(ContentTypeHeaderKey, ContentTextHeaderValue)
		}
	}

	return
}

func (w *GzipResponseWriter) WriteString(s string) (n int, err error) {
	n, err = w.Write([]byte(s))
	if err == nil {
		if w.ResponseWriter.Header()[ContentTypeHeaderKey] == nil {
			w.ResponseWriter.Header().Set(ContentTypeHeaderKey, ContentTextHeaderValue)
		}
	}
	return
}

func (w *GzipResponseWriter) WriteNow(contents []byte) (int, error) {
	if w.disabled {
		return w.ResponseWriter.Write(contents)
	}

	addGzipHeaders(w.ResponseWriter)

	return writeGzip(w.ResponseWriter, contents)
}

func addGzipHeaders(w ResponseWriter) {
	w.Header().Add(VaryHeaderKey, AcceptEncodingHeaderKey)
	w.Header().Add(ContentEncodingHeaderKey, GzipHeaderValue)
}

func (w *GzipResponseWriter) FlushResponse() {
	_, _ = w.WriteNow(w.chunks)
	w.ResponseWriter.FlushResponse()
}

func (w *GzipResponseWriter) ResetBody() {
	w.chunks = w.chunks[0:0]
}

func (w *GzipResponseWriter) Disable() {
	w.disabled = true
}
