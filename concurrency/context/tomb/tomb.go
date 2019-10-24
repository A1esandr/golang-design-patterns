// Пакет tomb предоставляет реализацию Context которая отменяется когда
// либо его родительский Context отменен или предоставленный Tomb уничтожен.
package tomb

import (
	"context"
	tomb "gopkg.in/tomb.v2"
)

// NewContext возвращает Context когда
// либо его родительский Context отменен или предоставленный t уничтожен.
func NewContext(parent context.Context, t *tomb.Tomb) context.Context {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		select {
		case <-t.Dying():
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx
}
