package literoute

import (
	"log"
	"net/http"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

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

	mux.AppendMiddleware(&LogMid{})

	mux.Get("/", Root)

	party := mux.Party("/todo")

	party.Get("/:name", Todo)

	log.Fatalln(http.ListenAndServe(":8080", mux))

}

func Root(ctx Context) {
	log.Println(ctx.GzipResponseWriter().WriteString(FormatTimeRFC3339Nano(time.Now().UTC())))
}

type TodoV struct {
	Id   string
	Name string
	Time time.Time
}

func Todo(ctx Context) {
	log.Println(ctx.JSON(TodoV{
		Id:   "1",
		Name: ctx.Param("name"),
		Time: UnixEpochTime,
	}))
}

type LogMid struct {
}

func (m *LogMid) Handle(ctx Context) {
	log.Println(ctx)
}
