package astilectron

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/os"
	"github.com/asticode/go-astitools/regexp"
	"github.com/pkg/errors"
)

// Var
var (
	defaultHTTPClient     = &http.Client{}
	regexpDarwinInfoPList = regexp.MustCompile("<string>Electron")
)

// Provisioner 是一个可以提供 Astilectron 的接口
type Provisioner interface {
	Provision(ctx context.Context, appName, os, arch string, p Paths) error
}

// mover is a function that moves a package
type mover func(ctx context.Context, p Paths) error

// DefaultProvisioner represents the default provisioner
var DefaultProvisioner = &defaultProvisioner{
	// 下载 https://github.com/Netpas/astilectron/archive/v0.27.1.zip 到指定目录
	moverAstilectron: func(ctx context.Context, p Paths) (err error) {
		if err = Download(ctx, defaultHTTPClient, p.AstilectronDownloadSrc(), p.AstilectronDownloadDst()); err != nil {
			return errors.Wrapf(err, "downloading %s into %s failed", p.AstilectronDownloadSrc(), p.AstilectronDownloadDst())
		}
		return
	},
	// 下载 https://github.com/electron/electron/releases/download/v1.8.1/electron-windows-386-v1.8.1.zip 到指定目录
	moverElectron: func(ctx context.Context, p Paths) (err error) {
		if err = Download(ctx, defaultHTTPClient, p.ElectronDownloadSrc(), p.ElectronDownloadDst()); err != nil {
			return errors.Wrapf(err, "downloading %s into %s failed", p.ElectronDownloadSrc(), p.ElectronDownloadDst())
		}
		return
	},
}

// defaultProvisioner represents the default provisioner
type defaultProvisioner struct {
	moverAstilectron mover
	moverElectron    mover
}

// provisionStatusElectronKey returns the electron's provision status key
func provisionStatusElectronKey(os, arch string) string {
	// eg: "windows-386"，用于 electron 的命名
	return fmt.Sprintf("%s-%s", os, arch)
}

// Provision 准备好 astilectron 和 electron，包括下载解压
func (p *defaultProvisioner) Provision(ctx context.Context, appName, os, arch string, paths Paths) (err error) {
	// 从 status.json 文件中读取 json 串，并解析到 PrivisionStatus 即s的结构中
	// 得到 s的内容如 {"astilectron":{"version":"0.27.1"},"electron":{"windows-386":{"version":"1.8.1"}}}
	var s ProvisionStatus
	if s, err = p.ProvisionStatus(paths); err != nil {
		err = errors.Wrap(err, "retrieving provisioning status failed")
		return
	}
	defer p.updateProvisionStatus(paths, &s)  // 函数退出时，用 s 更新 status.json 文件

	// 准备好astilectron
	if err = p.provisionAstilectron(ctx, paths, s); err != nil {
		err = errors.Wrap(err, "provisioning astilectron failed")
		return
	}
	s.Astilectron = &ProvisionStatusPackage{Version: VersionAstilectron}

	// 准备好electron
	if err = p.provisionElectron(ctx, paths, s, appName, os, arch); err != nil {
		err = errors.Wrap(err, "provisioning electron failed")
		return
	}
	s.Electron[provisionStatusElectronKey(os, arch)] = &ProvisionStatusPackage{Version: VersionElectron}
	return
}

// ProvisionStatus represents the provision status
type ProvisionStatus struct {
	Astilectron *ProvisionStatusPackage            `json:"astilectron,omitempty"`
	Electron    map[string]*ProvisionStatusPackage `json:"electron,omitempty"`
}

// ProvisionStatusPackage represents the provision status of a package
type ProvisionStatusPackage struct {
	Version string `json:"version"`
}

// ProvisionStatus 从 status.json 文件中读取 json 串，并解析到 PrivisionStatus 结构中
func (p *defaultProvisioner) ProvisionStatus(paths Paths) (s ProvisionStatus, err error) {
	// 打开文件: C:\Users\Caleb\AppData\Roaming\Lets\vendor\status.json
	var f *os.File
	s.Electron = make(map[string]*ProvisionStatusPackage)
	if f, err = os.Open(paths.ProvisionStatus()); err != nil {
		if !os.IsNotExist(err) {
			err = errors.Wrapf(err, "opening file %s failed", paths.ProvisionStatus())
		} else {
			err = nil
		}
		return
	}
	defer f.Close()

	// 文件内容如：{"astilectron":{"version":"0.27.1"},"electron":{"windows-386":{"version":"1.8.1"}}}，可以解析到 s 中
	if errLocal := json.NewDecoder(f).Decode(&s); errLocal != nil {
		// For backward compatibility purposes, if there's an unmarshal error we delete the status file and make the
		// assumption that provisioning has to be done all over again
		astilog.Error(errors.Wrapf(errLocal, "json decoding from %s failed", paths.ProvisionStatus()))
		astilog.Debugf("Removing %s", f.Name())
		if errLocal = os.RemoveAll(f.Name()); errLocal != nil {
			astilog.Error(errors.Wrapf(errLocal, "removing %s failed", f.Name()))
		}
		return
	}
	return
}

// ProvisionStatus 使用 s 来更新 status.json 文件
func (p *defaultProvisioner) updateProvisionStatus(paths Paths, s *ProvisionStatus) (err error) {
	// Create the file: C:\Users\Caleb\AppData\Roaming\Lets\vendor\status.json
	var f *os.File
	if f, err = os.Create(paths.ProvisionStatus()); err != nil {
		err = errors.Wrapf(err, "creating file %s failed", paths.ProvisionStatus())
		return
	}
	defer f.Close()

	// Marshal
	if err = json.NewEncoder(f).Encode(s); err != nil {
		err = errors.Wrapf(err, "json encoding into %s failed", paths.ProvisionStatus())
		return
	}
	return
}

// provisionAstilectron 准备好 Astilectron
func (p *defaultProvisioner) provisionAstilectron(ctx context.Context, paths Paths, s ProvisionStatus) error {
	return p.provisionPackage(ctx, paths, s.Astilectron, p.moverAstilectron, "Astilectron", VersionAstilectron,
		paths.AstilectronUnzipSrc(), paths.AstilectronDirectory(), nil)
}

// provisionElectron 准备好 Electron
func (p *defaultProvisioner) provisionElectron(ctx context.Context, paths Paths, s ProvisionStatus, appName, os, arch string) error {
	return p.provisionPackage(ctx, paths, s.Electron[provisionStatusElectronKey(os, arch)], p.moverElectron, "Electron", VersionElectron, paths.ElectronUnzipSrc(), paths.ElectronDirectory(), func() (err error) {
		switch os {
		case "darwin":
			if err = p.provisionElectronFinishDarwin(appName, paths); err != nil {
				return errors.Wrap(err, "finishing provisioning electron for darwin systems failed")
			}
		default:
			astilog.Debug("System doesn't require finshing provisioning electron, moving on...")
		}
		return
	})
}

// provisionPackage 下载并解压Astilectron 或 Electron，如果已经存在则直接返回
func (p *defaultProvisioner) provisionPackage(ctx context.Context, paths Paths, s *ProvisionStatusPackage,
	m mover, name, version, pathUnzipSrc, pathDirectory string, finish func() error) (err error) {
	// Astilectron 或 Electron 已经下载并安装好了
	if s != nil && s.Version == version {
		astilog.Debugf("%s has already been provisioned to version %s, moving on...", name, version)
		return
	}
	astilog.Debugf("Provisioning %s...", name)

	// 移除之前安装目录
	astilog.Debugf("Removing directory %s", pathDirectory)
	if err = os.RemoveAll(pathDirectory); err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "removing %s failed", pathDirectory)
	}

	// 下载 https://github.com/Netpas/astilectron/archive/v0.27.1.zip 到指定目录 (同 electron)
	if err = m(ctx, paths); err != nil {
		return errors.Wrapf(err, "moving %s failed", name)
	}

	// 解压压缩包 astilectron-v0.27.1.zip 到 C:\Users\Caleb\AppData\Roaming\Lets\vendor\astilectron (同 electron)
	var unzip func(string) error
	if name == "Electron" {
		unzip = unzipForElectron
	} else if name == "Astilectron" {
		unzip = unzipForAstilectron
	}
	if err = unzip(paths.vendorDirectory); err != nil {
		astilog.Debug(err)
		if name == "Astilectron" {
			os.RemoveAll(paths.vendorDirectory + `\astilectron`)
		} else if name == "Electron" {
			os.RemoveAll(fmt.Sprintf("%s\\electron-%s-%s", paths.vendorDirectory, runtime.GOOS, runtime.GOARCH))
		}
		// 创建目录
		astilog.Debugf("Creating directory %s", pathDirectory)
		if err = os.MkdirAll(pathDirectory, 0755); err != nil {
			return errors.Wrapf(err, "mkdirall %s failed", pathDirectory)
		}
		// 解压
		if err = Unzip(ctx, pathUnzipSrc, pathDirectory); err != nil {
			return errors.Wrapf(err, "unzipping %s into %s failed", pathUnzipSrc, pathDirectory)
		}
	}

	// Finish：nil
	if finish != nil {
		if err = finish(); err != nil {
			return errors.Wrap(err, "finishing failed")
		}
	}
	return
}

// provisionElectronFinishDarwin finishes provisioning electron for Darwin systems
// https://github.com/electron/electron/blob/v1.8.1/docs/tutorial/application-distribution.md#macos
func (p *defaultProvisioner) provisionElectronFinishDarwin(appName string, paths Paths) (err error) {
	// Log
	astilog.Debug("Finishing provisioning electron for darwin system")

	// Custom app icon
	if paths.AppIconDarwinSrc() != "" {
		if err = p.provisionElectronFinishDarwinCopy(paths); err != nil {
			return errors.Wrap(err, "copying for darwin system finish failed")
		}
	}

	// Custom app name
	if appName != "" {
		// Replace
		if err = p.provisionElectronFinishDarwinReplace(appName, paths); err != nil {
			return errors.Wrap(err, "replacing for darwin system finish failed")
		}

		// Rename
		if err = p.provisionElectronFinishDarwinRename(appName, paths); err != nil {
			return errors.Wrap(err, "renaming for darwin system finish failed")
		}
	}
	return
}

// provisionElectronFinishDarwinCopy copies the proper darwin files
func (p *defaultProvisioner) provisionElectronFinishDarwinCopy(paths Paths) (err error) {
	// Icon
	var src, dst = paths.AppIconDarwinSrc(), filepath.Join(paths.ElectronDirectory(), "Electron.app", "Contents", "Resources", "electron.icns")
	if src != "" {
		astilog.Debugf("Copying %s to %s", src, dst)
		if err = astios.Copy(context.Background(), src, dst); err != nil {
			return errors.Wrapf(err, "copying %s to %s failed", src, dst)
		}
	}
	return
}

// provisionElectronFinishDarwinReplace makes the proper replacements in the proper darwin files
func (p *defaultProvisioner) provisionElectronFinishDarwinReplace(appName string, paths Paths) (err error) {
	for _, p := range []string{
		filepath.Join(paths.electronDirectory, "Electron.app", "Contents", "Info.plist"),
		filepath.Join(paths.electronDirectory, "Electron.app", "Contents", "Frameworks", "Electron Helper EH.app", "Contents", "Info.plist"),
		filepath.Join(paths.electronDirectory, "Electron.app", "Contents", "Frameworks", "Electron Helper NP.app", "Contents", "Info.plist"),
		filepath.Join(paths.electronDirectory, "Electron.app", "Contents", "Frameworks", "Electron Helper.app", "Contents", "Info.plist"),
	} {
		// Log
		astilog.Debugf("Replacing in %s", p)

		// Read file
		var b []byte
		if b, err = ioutil.ReadFile(p); err != nil {
			return errors.Wrapf(err, "reading %s failed", p)
		}

		// Open and truncate file
		var f *os.File
		if f, err = os.Create(p); err != nil {
			return errors.Wrapf(err, "creating %s failed", p)
		}
		defer f.Close()

		// Replace
		astiregexp.ReplaceAll(regexpDarwinInfoPList, &b, []byte("<string>"+appName))

		// Write
		if _, err = f.Write(b); err != nil {
			return errors.Wrapf(err, "writing to %s failed", p)
		}
	}
	return
}

// rename represents a rename
type rename struct {
	src, dst string
}

// provisionElectronFinishDarwinRename renames the proper darwin folders
func (p *defaultProvisioner) provisionElectronFinishDarwinRename(appName string, paths Paths) (err error) {
	var appDirectory = filepath.Join(paths.electronDirectory, appName+".app")
	var frameworksDirectory = filepath.Join(appDirectory, "Contents", "Frameworks")
	var helperEH = filepath.Join(frameworksDirectory, appName+" Helper EH.app")
	var helperNP = filepath.Join(frameworksDirectory, appName+" Helper NP.app")
	var helper = filepath.Join(frameworksDirectory, appName+" Helper.app")
	for _, r := range []rename{
		{src: filepath.Join(paths.electronDirectory, "Electron.app"), dst: appDirectory},
		{src: filepath.Join(appDirectory, "Contents", "MacOS", "Electron"), dst: paths.AppExecutable()},
		{src: filepath.Join(frameworksDirectory, "Electron Helper EH.app"), dst: helperEH},
		{src: filepath.Join(helperEH, "Contents", "MacOS", "Electron Helper EH"), dst: filepath.Join(helperEH, "Contents", "MacOS", appName+" Helper EH")},
		{src: filepath.Join(frameworksDirectory, "Electron Helper NP.app"), dst: filepath.Join(helperNP)},
		{src: filepath.Join(helperNP, "Contents", "MacOS", "Electron Helper NP"), dst: filepath.Join(helperNP, "Contents", "MacOS", appName+" Helper NP")},
		{src: filepath.Join(frameworksDirectory, "Electron Helper.app"), dst: filepath.Join(helper)},
		{src: filepath.Join(helper, "Contents", "MacOS", "Electron Helper"), dst: filepath.Join(helper, "Contents", "MacOS", appName+" Helper")},
	} {
		astilog.Debugf("Renaming %s into %s", r.src, r.dst)
		if err = os.Rename(r.src, r.dst); err != nil {
			return errors.Wrapf(err, "renaming %s into %s failed", r.src, r.dst)
		}
	}
	return
}

// Disembedder is a functions that allows to disembed data from a path
// 这种函数用在 Asset、AssetDir 这些成员上
type Disembedder func(src string) ([]byte, error)

// NewDisembedderProvisioner creates a provisioner that can provision based on embedded data
func NewDisembedderProvisioner(d Disembedder, pathAstilectron, pathElectron string) Provisioner {
	return &defaultProvisioner{
		moverAstilectron: func(ctx context.Context, p Paths) (err error) {
			if err = Disembed(ctx, d, pathAstilectron, p.AstilectronDownloadDst()); err != nil {
				return errors.Wrapf(err, "disembedding %s into %s failed", pathAstilectron, p.AstilectronDownloadDst())
			}
			return
		},
		moverElectron: func(ctx context.Context, p Paths) (err error) {
			if err = Disembed(ctx, d, pathElectron, p.ElectronDownloadDst()); err != nil {
				return errors.Wrapf(err, "disembedding %s into %s failed", pathElectron, p.ElectronDownloadDst())
			}
			return
		},
	}
}
