// 本文件就是对用户传入的 Options 参数做了整理后，初始化了一堆路径，方便后续代码执行使用

package astilectron

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Paths represents the set of paths needed by Astilectron
type Paths struct {
	appExecutable          string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\electron-windows-386\electron.exe
	appIconDarwinSrc       string // C:\Users\Caleb\AppData\Roaming\Lets\resources\dist\icon.icns
	appIconDefaultSrc      string // C:\Users\Caleb\AppData\Roaming\Lets\resources\dist\icon.png
	astilectronApplication string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\astilectron\main.js
	astilectronDirectory   string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\astilectron
	astilectronDownloadSrc string // https://github.com/Netpas/astilectron/archive/v0.27.1.zip
	astilectronDownloadDst string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\astilectron-v0.27.1.zip
	astilectronUnzipSrc    string // astilectron-v0.27.1.zip 里面的 astilectron-0.27.1 文件夹
	baseDirectory          string // C:\Program Files (x86)\letsvpn
	dataDirectory          string // C:\Users\Caleb\AppData\Roaming\Lets
	electronDirectory      string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\electron-windows-386
	electronDownloadSrc    string // https://github.com/electron/electron/releases/download/v1.8.1/electron-windows-386-v1.8.1.zip
	electronDownloadDst    string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\electron-windows-386-v1.8.1.zip
	electronUnzipSrc       string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\electron-windows-386-v1.8.1.zip
	provisionStatus        string // C:\Users\Caleb\AppData\Roaming\Lets\vendor\status.json
	vendorDirectory        string // C:\Users\Caleb\AppData\Roaming\Lets\vendor
}

// newPaths creates new paths
func newPaths(os, arch string, o Options) (p *Paths, err error) {
	// 初始化基础目录路径，如果用户不指定，则会默认使用 Lets.exe 所在目录路径
	p = &Paths{}
	if err = p.initBaseDirectory(o.BaseDirectoryPath); err != nil {
		err = errors.Wrap(err, "initializing base directory failed")
		return
	}

	// 初始化数据目录路径，如果用户不指定，则会默认使用例如"C:\Users\Caleb\AppData\Roaming\Lets"
	if err = p.initDataDirectory(o.DataDirectoryPath, o.AppName); err != nil {
		err = errors.Wrap(err, "initializing data directory failed")
		return
	}

	// Init other paths
	// C:\Users\Caleb\AppData\Roaming\Lets\resources\dist\icon.icns
	p.appIconDarwinSrc = o.AppIconDarwinPath
	if len(p.appIconDarwinSrc) > 0 && !filepath.IsAbs(p.appIconDarwinSrc) {
		p.appIconDarwinSrc = filepath.Join(p.dataDirectory, p.appIconDarwinSrc)
	}
	// C:\Users\Caleb\AppData\Roaming\Lets\resources\dist\icon.png
	p.appIconDefaultSrc = o.AppIconDefaultPath
	if len(p.appIconDefaultSrc) > 0 && !filepath.IsAbs(p.appIconDefaultSrc) {
		p.appIconDefaultSrc = filepath.Join(p.dataDirectory, p.appIconDefaultSrc)
	}
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor
	p.vendorDirectory = filepath.Join(p.dataDirectory, "vendor")
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor\status.json
	// 该文件存放的内容如：{"astilectron":{"version":"0.27.1"},"electron":{"windows-386":{"version":"1.8.1"}}}
	p.provisionStatus = filepath.Join(p.vendorDirectory, "status.json")
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor\astilectron
	p.astilectronDirectory = filepath.Join(p.vendorDirectory, "astilectron")
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor\astilectron\main.js
	p.astilectronApplication = filepath.Join(p.astilectronDirectory, "main.js")
	// https://github.com/Netpas/astilectron/archive/v0.27.1.zip
	p.astilectronDownloadSrc = AstilectronDownloadSrc()
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor\astilectron-v0.27.1.zip
	p.astilectronDownloadDst = filepath.Join(p.vendorDirectory, fmt.Sprintf("astilectron-v%s.zip", VersionAstilectron))
	// astilectron-v0.27.1.zip 里面的 astilectron-0.27.1 文件夹
	p.astilectronUnzipSrc = filepath.Join(p.astilectronDownloadDst, fmt.Sprintf("astilectron-%s", VersionAstilectron))
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor\electron-windows-386
	p.electronDirectory = filepath.Join(p.vendorDirectory, fmt.Sprintf("electron-%s-%s", os, arch))
	// https://github.com/electron/electron/releases/download/v1.8.1/electron-windows-386-v1.8.1.zip
	p.electronDownloadSrc = ElectronDownloadSrc(os, arch)
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor\electron-windows-386-v1.8.1.zip
	p.electronDownloadDst = filepath.Join(p.vendorDirectory, fmt.Sprintf("electron-%s-%s-v%s.zip", os, arch, VersionElectron))
	p.electronUnzipSrc = p.electronDownloadDst
	// C:\Users\Caleb\AppData\Roaming\Lets\vendor\electron-windows-386\electron.exe
	p.initAppExecutable(os, o.AppName)
	return
}

// initBaseDirectory 最终得到一个目录绝对路径
func (p *Paths) initBaseDirectory(baseDirectoryPath string) (err error) {
	// 获取 p.baseDirectory
	p.baseDirectory = baseDirectoryPath
	if len(p.baseDirectory) == 0 {
		var ep string
		if ep, err = os.Executable(); err != nil {
			err = errors.Wrap(err, "retrieving executable path failed")
			return
		}
		p.baseDirectory = filepath.Dir(ep)
	}

	// 保证 p.baseDirectory 必须是一个绝对路径
	if p.baseDirectory, err = filepath.Abs(p.baseDirectory); err != nil {
		err = errors.Wrap(err, "computing absolute path failed")
		return
	}
	return
}

// initDataDirectory 最终得到一个目录绝对路径
func (p *Paths) initDataDirectory(dataDirectoryPath, appName string) (err error) {
	// Path is specified in the options
	if len(dataDirectoryPath) > 0 {
		// We need the absolute path
		if p.dataDirectory, err = filepath.Abs(dataDirectoryPath); err != nil {
			err = errors.Wrapf(err, "computing absolute path of %s failed", dataDirectoryPath)
			return
		}
		return
	}

	// If the APPDATA env exists, we use it. APPDATA eg: "C:\Users\Caleb\AppData\Roaming"
	if v := os.Getenv("APPDATA"); len(v) > 0 {
		p.dataDirectory = filepath.Join(v, appName)
		return
	}

	// Default to base directory path
	p.dataDirectory = p.baseDirectory
	return
}

// AstilectronDownloadSrc returns the download URL of the (currently platform-independent) astilectron zip file
func AstilectronDownloadSrc() string {
	return fmt.Sprintf("https://github.com/Netpas/astilectron/archive/v%s.zip", VersionAstilectron)
}

// ElectronDownloadSrc returns the download URL of the platform-dependant electron zipfile
func ElectronDownloadSrc(os, arch string) string {
	// Get OS name
	var o string
	switch strings.ToLower(os) {
	case "darwin":
		o = "darwin"
	case "linux":
		o = "linux"
	case "windows":
		o = "win32"
	}

	// Get arch name
	var a = "ia32"
	if strings.ToLower(arch) == "amd64" || o == "darwin" {
		a = "x64"
	} else if strings.ToLower(arch) == "arm" && o == "linux" {
		a = "armv7l"
	}

	// Return url
	return fmt.Sprintf("https://github.com/electron/electron/releases/download/v%s/electron-v%s-%s-%s.zip", VersionElectron, VersionElectron, o, a)
}

// initAppExecutable initializes the app executable path
func (p *Paths) initAppExecutable(os, appName string) {
	switch os {
	case "darwin":
		if appName == "" {
			appName = "Electron"
		}
		p.appExecutable = filepath.Join(p.electronDirectory, appName+".app", "Contents", "MacOS", appName)
	case "linux":
		p.appExecutable = filepath.Join(p.electronDirectory, "electron")
	case "windows":
		p.appExecutable = filepath.Join(p.electronDirectory, "electron.exe")
	}
}

// AppExecutable returns the app executable path
func (p Paths) AppExecutable() string {
	return p.appExecutable
}

// AppIconDarwinSrc returns the darwin app icon path
func (p Paths) AppIconDarwinSrc() string {
	return p.appIconDarwinSrc
}

// AppIconDefaultSrc returns the default app icon path
func (p Paths) AppIconDefaultSrc() string {
	return p.appIconDefaultSrc
}

// BaseDirectory returns the base directory path
func (p Paths) BaseDirectory() string {
	return p.baseDirectory
}

// AstilectronApplication returns the astilectron application path
func (p Paths) AstilectronApplication() string {
	return p.astilectronApplication
}

// AstilectronDirectory returns the astilectron directory path
func (p Paths) AstilectronDirectory() string {
	return p.astilectronDirectory
}

// AstilectronDownloadDst returns the astilectron download destination path
func (p Paths) AstilectronDownloadDst() string {
	return p.astilectronDownloadDst
}

// AstilectronDownloadSrc returns the astilectron download source path
func (p Paths) AstilectronDownloadSrc() string {
	return p.astilectronDownloadSrc
}

// AstilectronUnzipSrc returns the astilectron unzip source path
func (p Paths) AstilectronUnzipSrc() string {
	return p.astilectronUnzipSrc
}

// DataDirectory returns the data directory path
func (p Paths) DataDirectory() string {
	return p.dataDirectory
}

// ElectronDirectory returns the electron directory path
func (p Paths) ElectronDirectory() string {
	return p.electronDirectory
}

// ElectronDownloadDst returns the electron download destination path
func (p Paths) ElectronDownloadDst() string {
	return p.electronDownloadDst
}

// ElectronDownloadSrc returns the electron download source path
func (p Paths) ElectronDownloadSrc() string {
	return p.electronDownloadSrc
}

// ElectronUnzipSrc returns the electron unzip source path
func (p Paths) ElectronUnzipSrc() string {
	return p.electronUnzipSrc
}

// ProvisionStatus returns the provision status path
func (p Paths) ProvisionStatus() string {
	return p.provisionStatus
}

// VendorDirectory returns the vendor directory path
func (p Paths) VendorDirectory() string {
	return p.vendorDirectory
}
