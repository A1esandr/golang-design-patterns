// Пакет gorilla предоставляет реализацию go.net/context.Context
// чей метод Value возвращает значения, связанные с определенным HTTP запросом
// в github.com/gorilla/context пакете.
package gorilla

import (
	"net/http"

	"context"
	gcontext "github.com/gorilla/context"
)

// NewContext возвращает Context чей метод Value возвращает значения, связанные
// с req, использующим Gorilla context пакет:
// http://www.gorillatoolkit.org/pkg/context
func NewContext(parent context.Context, req *http.Request) context.Context {
	return &wrapper{parent, req}
}

type wrapper struct {
	context.Context
	req *http.Request
}

type key int

const reqKey key = 0

// Value возвращает значение пакет context Gorilla для данного запроса Context и ключа.
// Он передает родительский Context если нет такого значения.
func (ctx *wrapper) Value(key interface{}) interface{} {
	if key == reqKey {
		return ctx.req
	}
	if val, ok := gcontext.GetOk(ctx.req, key); ok {
		return val
	}
	return ctx.Context.Value(key)
}

// HTTPRequest возвращает *http.Request связанный с ctx использующий NewContext, если присуствует
func HTTPRequest(ctx context.Context) (*http.Request, bool) {
	// Не можем использовать ctx.(*wrapper).req чтобы получить запрос потому что ctx может
	// быть производным Context из *wrapper. Вместо этого используем Value
	// для получения доступа к запросу если он где-то выше в дереве Context.
	req, ok := ctx.Value(reqKey).(*http.Request)
	return req, ok
}
