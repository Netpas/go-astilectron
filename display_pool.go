package astilectron

import "sync"

// displayPool represents a display pool
type displayPool struct {
	d map[int64]*Display     // key：Unique identifier associated with the display
	m *sync.Mutex
}

// newDisplayPool creates a new display pool
func newDisplayPool() *displayPool {
	return &displayPool{
		d: make(map[int64]*Display),
		m: &sync.Mutex{},
	}
}

// all returns all the displays
func (p *displayPool) all() (ds []*Display) {
	p.m.Lock()
	defer p.m.Unlock()
	ds = []*Display{}
	for _, d := range p.d {
		ds = append(ds, d)
	}
	return
}

// primary returns the primary display, it defaults to the last display
func (p *displayPool) primary() (d *Display) {
	p.m.Lock()
	defer p.m.Unlock()
	for _, d = range p.d {
		if d.primary {
			return
		}
	}
	return
}

// update updates the pool based on event displays
func (p *displayPool) update(e *EventDisplays) {
	p.m.Lock()
	defer p.m.Unlock()

	// 用 e 来更新 p 中的 成员
	// 几个结构体之间的关系：displayPool 指向多个 Display, Display 指向 DisplayOptions；
	//                       EventDisplays 包含了多个 DisplayOptions；
	// displayPool 和 EventDisplays 之间的关联就在于：ID --- Unique identifier associated with the display

	var ids = make(map[int64]bool)
	// 对于参数中所有的 DisplayOptions，填充 p.d
	for _, o := range e.All {
		ids[*o.ID] = true
		var primary bool
		if *o.ID == *e.Primary.ID {
			primary = true
		}
		if d, ok := p.d[*o.ID]; ok {
			d.primary = primary
			*d.o = *o
		} else {
			p.d[*o.ID] = newDisplay(o, primary)
		}
	}
	// 检查所有 displayPool 中的成员，如果不在 ids 中存在（也就是旧的 display），则删除之
	for id := range p.d {
		if _, ok := ids[id]; !ok {
			delete(p.d, id)
		}
	}
}
