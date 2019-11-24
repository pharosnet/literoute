package literoute

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkLiteMux(b *testing.B) {
	b.N = 10000000
	request, _ := http.NewRequest("GET", "/a", nil)
	response := httptest.NewRecorder()
	mux := New(Config{
		BodyEncoder: JsonBodyEncode,
		Status: CustomizeStatus{
			Succeed:        200,
			Fail:           555,
			NotFound:       444,
			InvalidRequest: 440,
		},
		PostMaxMemory: DefaultPostMaxMemory,
	})

	mux.Get("/", Bench)
	mux.Get("/a", Bench)
	mux.Get("/ab", Bench)
	mux.Get("/abc", Bench)

	for n := 0; n < b.N; n++ {

		mux.ServeHTTP(response, request)
	}
}

func Bench(ctx Context) {
	_, _ = ctx.Write([]byte("b"))
}
