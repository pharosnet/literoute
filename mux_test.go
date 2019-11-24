package literoute

import (
	"log"
	"net/http"
	"strconv"
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

	mux.RegisterValidator("IsInt", &IsIntValidator{})

	mux.Get("/", Root)

	party := mux.Party("/todo")

	party.Get("/:name|IsInt", TodoGet)

	party.Get("/:name/:var", TodoVar)

	log.Fatalln(http.ListenAndServe(":8080", mux))

}

type IsIntValidator struct{}

func (v *IsIntValidator) Validate(param string) bool {
	if _, err := strconv.Atoi(param); err != nil {
		return false
	}
	return true
}

func (v *IsIntValidator) OnFail(ctx Context) {
	ctx.Fail(map[string]string{
		"msg": "param is not int",
	})
}

func Root(ctx Context) {
	log.Println(
		ctx.GzipResponseWriter().WriteString(
			FormatTimeRFC3339Nano(time.Now().UTC()),
		),
	)
}

type Todo struct {
	Id   string
	Name string
	Var  string
	Time time.Time
}

func TodoGet(ctx Context) {
	ctx.Fail(Todo{
		Id:   "1",
		Name: ctx.Param("name"),
		Time: time.Now(),
	})
}

func TodoVar(ctx Context) {
	ctx.Fail(Todo{
		Id:   "1",
		Name: ctx.Param("name"),
		Var:  ctx.Param("var"),
		Time: time.Now(),
	})
}

type LogMid struct {
}

func (m *LogMid) Handle(ctx Context) bool {
	log.Println(ctx)
	//ctx.Fail(map[string]time.Time{
	//	"fail":time.Now(),
	//})
	return true
}
