// 本文件用于描述基础对象 object, 注意：object 对象实现了 Listenable 接口

package astilectron

import (
	"context"
	"errors"

	"github.com/asticode/go-astitools/context"
)

// Object errors
var (
	ErrCancellerCancelled = errors.New("canceller.cancelled")
	ErrObjectDestroyed    = errors.New("object.destroyed")
)

// object represents a base object
type object struct {
	cancel context.CancelFunc
	ctx    context.Context
	c      *asticontext.Canceller
	d      *dispatcher
	i      *identifier
	id     string
	w      *writer
}

// newObject returns a new base object
func newObject(parentCtx context.Context, c *asticontext.Canceller, d *dispatcher, i *identifier, w *writer, id string) (o *object) {
	o = &object{
		c:  c,
		d:  d,
		i:  i,
		id: id,
		w:  w,
	}
	if parentCtx != nil {
		o.ctx, o.cancel = context.WithCancel(parentCtx)
	} else {
		o.ctx, o.cancel = c.NewContext()    // 等同于 context.WithCancel(c.ctx)  （c.ctx 是 parentCtx）
	}
	return
}

// isActionable checks whether any type of action is allowed on the window
func (o *object) isActionable() error {
	if o.c.Cancelled() {
		return ErrCancellerCancelled
	} else if o.IsDestroyed() {
		return ErrObjectDestroyed
	}
	return nil
}

// IsDestroyed checks whether the window has been destroyed
func (o *object) IsDestroyed() bool {
	return o.ctx.Err() != nil
}

// On implements the Listenable interface
func (o *object) On(eventName string, l Listener) {
	o.d.addListener(o.id, eventName, l)
}
