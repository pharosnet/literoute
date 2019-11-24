# Lite Route
One lite http Route For GoLang. It supports:

- URL Parameters
- Party (Sub Route)
- Http Middleware
- Customize Http Status Code
- Customize Body Encoder and Decoder
- Route Parameters Validators
- Lite and Fast
- No dependency libs

### Example

Install

```
go get -u github.com/pharosnet/literoute
```

New Mux

```go
mux := literoute.New(Config{
		BodyEncoder: JsonBodyEncode,
		Status: CustomizeStatus{
			Succeed:        200,
			Fail:           555,
			NotFound:       444,
			InvalidRequest: 440,
		},
		PostMaxMemory: DefaultPostMaxMemory,
	})
```

Append Middleware

```go
type LogMid struct {}

func (m *LogMid) Handle(ctx Context) bool {
	log.Println(ctx)
	return true
}

mux.AppendMiddleware(&LogMid{})
```

Register Handlers

```go
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

mux.Get("/", Root)

party := mux.Party("/todo")

party.Get("/:name|IsInt", TodoGet)

party.Get("/:name/:var", TodoVar)

```

Register Validator

```go
type IsIntValidator struct {}

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

mux.RegisterValidator("IsInt", &IsIntValidator{})
```

Http Listen And Serve

```go
log.Fatalln(http.ListenAndServe(":8080", mux))
```

### Speed

```
goos: windows
goarch: amd64
pkg: github.com/pharosnet/literoute
BenchmarkLiteMux-16    	10000000	       148 ns/op
PASS
```

