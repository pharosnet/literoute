package literoute

type Middleware interface {
	Handle(ctx Context)
}


