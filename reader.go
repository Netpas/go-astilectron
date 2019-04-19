package astilectron

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/asticode/go-astilog"
)

// reader represents an object capable of reading in the TCP server
type reader struct {
	ctx context.Context
	d   *dispatcher
	rc  io.ReadCloser
}

// newReader creates a new reader
func newReader(ctx context.Context, d *dispatcher, rc io.ReadCloser) *reader {
	return &reader{
		ctx: ctx,
		d:   d,
		rc:  rc,
	}
}

// close closes the reader properly
func (r *reader) close() error {
	return r.rc.Close()
}

// isEOFErr checks whether the error is an EOF error
// wsarecv is the error sent on Windows when the client closes its connection
func (r *reader) isEOFErr(err error) bool {
	return err == io.EOF || strings.Contains(strings.ToLower(err.Error()), "wsarecv:")
}

// read reads from stdout 从一个地方把 json 字符串的 event 读出来，unmarshal 成为 event，分发给 r.d 的监听者们
//                        可见事件的传递是通过 json 字符串进行的
func (r *reader) read() {
	var reader = bufio.NewReader(r.rc)
	for {
		// Check context error
		if r.ctx.Err() != nil {
			return
		}

		// Read next line
		var b []byte
		var err error
		if b, err = reader.ReadBytes('\n'); err != nil {
			if !r.isEOFErr(err) {
				astilog.Errorf("%s while reading", err)
				continue
			}
			return
		}
		b = bytes.TrimSpace(b)
		astilog.Debugf("Astilectron says: %s", b)

		// Unmarshal
		var e Event
		if err = json.Unmarshal(b, &e); err != nil {
			astilog.Errorf("%s while unmarshaling %s", err, b)
			continue
		}

		// Dispatch 把事件 e 发送给 r.d 的监听者们
		r.d.dispatch(e)
	}
}
