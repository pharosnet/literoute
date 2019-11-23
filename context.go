package literoute

import (
	"net/http"
)

type Context interface {
}

func newSimpleContext(rw http.ResponseWriter, r *http.Request) (ctx Context) {

	return
}
