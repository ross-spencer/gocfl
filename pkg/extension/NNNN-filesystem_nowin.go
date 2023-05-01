//go:build !windows

package extension

import (
	"emperror.dev/errors"
	syscall "golang.org/x/sys/unix"
	"io/fs"
	"os"
	"runtime"
	"time"
)

func (fsm *filesystemMeta) init(fullpath string, fileInfo fs.FileInfo) error {
	fsm.OS = runtime.GOOS
	sys := fileInfo.Sys()
	if sys == nil {
		return errors.New("fileInfo.Sys() is nil")
	}
	stat_t, ok := sys.(*syscall.Stat_t)
	if !ok {
		return errors.New("fileInfo.Sys() is not *syscall.Stat_t")
	}
	fsm.ATime = time.Unix(stat_t.Atim.Sec, stat_t.Atim.NSec)
	fsm.CTime = time.Unix(stat_t.Ctim.Sec, stat_t.Ctim.NSec)
	fsm.MTime = time.Unix(stat_t.Mtim.Sec, stat_t.Mtim.NSec)
	fi, err := os.Lstat(fullpath)
	if err != nil {
		return errors.WithStack(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		fsm.Symlink, err = os.Readlink(fullpath)
		if err != nil {
			return errors.Wrapf(err, "cannot read Symlink %s", fullpath)
		}
	}
	unixPerms := fileMode & os.ModePerm
	fsm.Attr = unixPerms.String()
	fsm.OSStat = stat_t

	return nil
}
