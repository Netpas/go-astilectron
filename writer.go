package astilectron

import (
	"encoding/json"
	"io"

	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// writer represents an object capable of writing in the TCP server
type writer struct {
	wc io.WriteCloser
}

// newWriter creates a new writer
func newWriter(wc io.WriteCloser) *writer {
	return &writer{
		wc: wc,
	}
}

// close closes the writer properly
func (w *writer) close() error {
	return w.wc.Close()
}

// write writes to the stdin 把 event 转换成 json 字符串 用 writer 写到一个地方
func (w *writer) write(e Event) (err error) {
	// Marshal
	var b []byte
	if b, err = json.Marshal(e); err != nil {
		return errors.Wrapf(err, "Marshaling %+v failed", e)
	}

	// Write
	astilog.Debugf("Sending to Astilectron: %s", b)
	if _, err = w.wc.Write(append(b, '\n')); err != nil {
		return errors.Wrapf(err, "Writing %s failed", b)
	}
	return
}
