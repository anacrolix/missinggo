package reqctx

import (
	"context"
	"net/http"

	"github.com/anacrolix/missinggo/expect"
)

func SetNewValue(r *http.Request, key, value interface{}) *http.Request {
	expect.Nil(r.Context().Value(key))
	expect.NotNil(value)
	return r.WithContext(context.WithValue(r.Context(), key, value))
}
