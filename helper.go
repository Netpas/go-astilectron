package astilectron

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mholt/archiver"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/archive"
	"github.com/asticode/go-astitools/context"
	"github.com/asticode/go-astitools/http"
	"github.com/asticode/go-astitools/io"
	"github.com/pkg/errors"
)

// Download is a cancellable function that downloads a src into a dst using a specific *http.Client and cleans up on
// failed downloads
func Download(ctx context.Context, c *http.Client, src, dst string) (err error) {
	// Log
	astilog.Debugf("Downloading %s into %s", src, dst)

	// Destination already exists
	if _, err = os.Stat(dst); err == nil {
		astilog.Debugf("%s already exists, skipping download...", dst)
		return
	} else if !os.IsNotExist(err) {
		return errors.Wrapf(err, "stating %s failed", dst)
	}
	err = nil

	// Clean up on error
	defer func(err *error) {
		if *err != nil || ctx.Err() != nil {
			astilog.Debugf("Removing %s...", dst)
			os.Remove(dst)
		}
	}(&err)

	// Make sure the dst directory  exists
	if err = os.MkdirAll(filepath.Dir(dst), 0775); err != nil {
		return errors.Wrapf(err, "mkdirall %s failed", filepath.Dir(dst))
	}

	// Download
	if err = astihttp.Download(ctx, c, src, dst); err != nil {
		return errors.Wrap(err, "astihttp.Download failed")
	}
	return
}

// Disembed is a cancellable disembed of an src to a dst using a custom Disembedder
// 该函数不是用下载来获取zip包，而是用已有的zip包拷贝一下就可以了
func Disembed(ctx context.Context, d Disembedder, src, dst string) (err error) {
	// Log
	astilog.Debugf("Disembedding %s into %s...", src, dst)

	// No need to disembed
	if _, err = os.Stat(dst); err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "stating %s failed", dst)
	} else if err == nil {
		astilog.Debugf("%s already exists, skipping disembed...", dst)
		return
	}
	err = nil

	// Clean up on error
	defer func(err *error) {
		if *err != nil || ctx.Err() != nil {
			astilog.Debugf("Removing %s...", dst)
			os.Remove(dst)
		}
	}(&err)

	// Make sure directory exists
	var dirPath = filepath.Dir(dst)
	astilog.Debugf("Creating %s", dirPath)
	if err = os.MkdirAll(dirPath, 0755); err != nil {
		return errors.Wrapf(err, "mkdirall %s failed", dirPath)
	}

	// Create dst
	var f *os.File
	astilog.Debugf("Creating %s", dst)
	if f, err = os.Create(dst); err != nil {
		return errors.Wrapf(err, "creating %s failed", dst)
	}
	defer f.Close()

	// Disembed
	var b []byte
	astilog.Debugf("Disembedding %s", src)
	if b, err = d(src); err != nil {
		return errors.Wrapf(err, "disembedding %s failed", src)
	}

	// Copy
	astilog.Debugf("Copying disembedded data to %s", dst)
	if _, err = astiio.Copy(ctx, bytes.NewReader(b), f); err != nil {
		return errors.Wrapf(err, "copying disembedded data into %s failed", dst)
	}
	return
}

// Unzip unzips a src into a dst.
// Possible src formats are /path/to/zip.zip or /path/to/zip.zip/internal/path.
func Unzip(ctx context.Context, src, dst string) (err error) {
	// Clean up on error
	defer func(err *error) {
		if *err != nil || ctx.Err() != nil {
			astilog.Debugf("Removing %s...", dst)
			os.RemoveAll(dst)
		}
	}(&err)

	// Unzipping
	/* eg : Unzipping
	   C:\\Users\\Caleb\\go\\src\\lets-civet.windows\\output\\windows-386\\Lets\\vendor\\electron-windows-386-v1.8.1.zip
	   into
	   C:\\Users\\Caleb\\go\\src\\lets-civet.windows\\output\\windows-386\\Lets\\vendor\\electron-windows-386  */
	astilog.Debugf("Unzipping %s into %s", src, dst)
	if err = astiarchive.Unzip(ctx, src, dst); err != nil {
		err = errors.Wrapf(err, "unzipping %s into %s failed", src, dst)
		return
	}
	return
}

// unzipForAstilectron 解压Astilectron
func unzipForAstilectron(dir string) (err error) {
	zipPath := fmt.Sprintf("%s\\astilectron-v%s.zip", dir, VersionAstilectron)
	if err = archiver.Unarchive(zipPath, dir); err != nil {
		return
	}
	os.RemoveAll(zipPath)
	time.Sleep(time.Second)
	if err = os.Rename(fmt.Sprintf("%s\\astilectron-%s", dir, VersionAstilectron), dir+`\astilectron`); err != nil {
		return
	}
	return
}

// unzipForElectron 解压Electron
func unzipForElectron(dir string) (err error) {
	zipPath := fmt.Sprintf("%s\\electron-%s-%s-v%s.zip", dir, runtime.GOOS, runtime.GOARCH, VersionElectron)
	if err = archiver.Unarchive(zipPath,
		fmt.Sprintf("%s\\electron-%s-%s", dir, runtime.GOOS, runtime.GOARCH)); err != nil {
		return
	}
	os.RemoveAll(zipPath)
	return
}

// PtrBool transforms a bool into a *bool
func PtrBool(i bool) *bool {
	return &i
}

// PtrFloat transforms a float64 into a *float64
func PtrFloat(i float64) *float64 {
	return &i
}

// PtrInt transforms an int into an *int
func PtrInt(i int) *int {
	return &i
}

// PtrInt64 transforms an int64 into an *int64
func PtrInt64(i int64) *int64 {
	return &i
}

// PtrStr transforms a string into a *string
func PtrStr(i string) *string {
	return &i
}

// synchronousFunc 该函数执行fn()，然后 <作为监听者一直等到收到一个eventNameDone事件> 或 <被cancelled> 导致退出函数
func synchronousFunc(c *asticontext.Canceller, l listenable, fn func(), eventNameDone string) (e Event) {
	var ctx, cancel = c.NewContext()
	defer cancel()
	// 这个监听者收到一个事件只是将其作为返回值返回
	l.On(eventNameDone, func(i Event) (deleteListener bool) {
		e = i
		cancel()
		return true
	})
	fn()
	<-ctx.Done()
	return
}

// synchronousEvent 该函数发送一个 event，然后 <监听等待收到一个eventNameDone事件> 或 <被cancelled>
func synchronousEvent(c *asticontext.Canceller, l listenable, w *writer, i Event, eventNameDone string) (o Event, err error) {
	o = synchronousFunc(c, l, func() {
		if err = w.write(i); err != nil {
			err = errors.Wrapf(err, "writing %+v event failed", i)
			return
		}
		return
	}, eventNameDone)
	return
}
