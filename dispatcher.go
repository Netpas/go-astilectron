// 本文件描述 dispatcher 分发器的功能，用来对 Listener 进行操作（增加/删除/遍历/发消息）
// Listener 本质上就是一个函数，可以对参数 Event 进行操作

package astilectron

import "sync"

// Listener represents a listener executed when an event is dispatched
type Listener func(e Event) (deleteListener bool)

// listenable represents an object that can listen
// 该接口在此文件中没有被用到
type listenable interface {
	On(eventName string, l Listener)
}

// dispatcher represents an object capable of dispatching events
type dispatcher struct {
	// id 以递增的方式帮助生成 listener id
	id int
	// Indexed by target ID then by event name then be listener id
	// We use a map[int]Listener so that deletion is as smooth as possible
	// It means it doesn't store listeners in order
	l map[string]map[string]map[int]Listener    // map[targetID]map[eventName]map[listenerID]Listener
	m sync.Mutex
}

// newDispatcher creates a new dispatcher
func newDispatcher() *dispatcher {
	return &dispatcher{
		l: make(map[string]map[string]map[int]Listener),
	}
}

// addListener adds a listener
func (d *dispatcher) addListener(targetID, eventName string, l Listener) {
	d.m.Lock()
	defer d.m.Unlock()
	if _, ok := d.l[targetID]; !ok {
		d.l[targetID] = make(map[string]map[int]Listener)
	}
	if _, ok := d.l[targetID][eventName]; !ok {
		d.l[targetID][eventName] = make(map[int]Listener)
	}
	d.id++
	d.l[targetID][eventName][d.id] = l
}

// delListener delete a specific listener
func (d *dispatcher) delListener(targetID, eventName string, id int) {
	d.m.Lock()
	defer d.m.Unlock()
	if _, ok := d.l[targetID]; !ok {
		return
	}
	if _, ok := d.l[targetID][eventName]; !ok {
		return
	}
	delete(d.l[targetID][eventName], id)
}

// Dispatch dispatches an event 把事件 e 发送给 d 的监听者们
func (d *dispatcher) dispatch(e Event) {
	go func() {
		for id, l := range d.listeners(e.TargetID, e.Name) {
			if l(e) {
				d.delListener(e.TargetID, e.Name, id)
			}
		}
	}()
}

// listeners returns the listeners for a target ID and an event name
func (d *dispatcher) listeners(targetID, eventName string) (l map[int]Listener) {
	d.m.Lock()
	defer d.m.Unlock()
	l = map[int]Listener{}
	if _, ok := d.l[targetID]; !ok {
		return
	}
	if _, ok := d.l[targetID][eventName]; !ok {
		return
	}
	for k, v := range d.l[targetID][eventName] {
		l[k] = v
	}
	return
}
