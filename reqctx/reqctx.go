package reqctx

import (
	"context"
	"net/http"

	"github.com/anacrolix/missinggo/assert"
)

func SetNewValue(r *http.Request, key, value interface{}) *http.Request {
	assert.Nil(r.Context().Value(key))
	assert.NotNil(value)
	return r.WithContext(context.WithValue(r.Context(), key, value))
}
