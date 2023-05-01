package extension

import (
	"bufio"
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
	"os"
	"time"
)

const FilesystemName = "NNNN-filesystem"
const FilesystemDescription = "preserves filesytem metadata"

func NewFilesystemFS(fsys fs.FS) (*Filesystem, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &FilesystemConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	ext, err := NewFilesystem(config)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new indexer")
	}
	return ext, nil
}
func NewFilesystem(config *FilesystemConfig) (*Filesystem, error) {
	sl := &Filesystem{
		FilesystemConfig: config,
		buffer:           map[string]*bytes.Buffer{},
	}
	//	sl.writer = brotli.NewWriter(sl.buffer)
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type filesystemMeta struct {
	ATime   time.Time `json:"atime"`
	MTime   time.Time `json:"mtime"`
	CTime   time.Time `json:"ctime"`
	Attr    string    `json:"attr,omitempty"`
	Symlink string    `json:"symlink,omitempty"`
	OS      string    `json:"os"`
	OSStat  any       `json:"osstat,omitempty"`
}

type fileSystemLine struct {
	Path string          `json:"path"`
	Meta *filesystemMeta `json:"meta"`
}

type FilesystemConfig struct {
	*ocfl.ExtensionConfig
	StorageType string
	StorageName string
	Compress    string
}

type Filesystem struct {
	*FilesystemConfig
	fsys        fs.FS
	lastHead    string
	currentHead string
	buffer      map[string]*bytes.Buffer
	writer      *brotli.Writer
}

func (extFS *Filesystem) AddFileBefore(object ocfl.Object, sourceFS fs.FS, source, dest, area string) error {
	return nil
}

func (extFS *Filesystem) UpdateFileBefore(object ocfl.Object, sourceFS fs.FS, source, dest, area string) error {
	return nil
}

func (extFS *Filesystem) DeleteFileBefore(object ocfl.Object, dest string, area string) error {
	return nil
}

func (extFS *Filesystem) AddFileAfter(object ocfl.Object, sourceFS fs.FS, source []string, internalPath, digest, area string) error {
	inventory := object.GetInventory()
	head := inventory.GetHead()
	if _, ok := extFS.buffer[head]; !ok {
		extFS.buffer[head] = &bytes.Buffer{}
	}
	if extFS.currentHead != head {
		extFS.writer = brotli.NewWriter(extFS.buffer[head])
		extFS.currentHead = head
	}

	for _, src := range source {
		fullpath, err := writefs.Fullpath(sourceFS, src)
		if err != nil {
			if errors.Cause(err) == writefs.ErrNotImplemented {
				continue
			}
			return errors.Wrapf(err, "cannot get fullpath for '%v/%s'", sourceFS, src)
		}
		fsMeta := &filesystemMeta{}
		// we work only on local filesystems with this extension
		stat, err := os.Stat(fullpath)
		if err != nil {
			return errors.Wrapf(err, "cannot stat file '%s'", src)
		}
		if err := fsMeta.init(fullpath, stat); err != nil {
			return errors.Wrapf(err, "cannot init fsMeta for '%s'", fullpath)
		}
		newSrc, err := object.GetExtensionManager().BuildObjectStatePath(object, src, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build object extract path for '%s'", src)
		}
		fsLine := &fileSystemLine{
			Path: newSrc,
			Meta: fsMeta,
		}
		data, err := json.Marshal(fsLine)
		if err != nil {
			return errors.Errorf("cannot marshal result %v", fsLine)
		}
		if _, err := extFS.writer.Write(append(data, []byte("\n")...)); err != nil {
			return errors.Errorf("cannot brotli %s", string(data))
		}
	}
	return nil
}

func (extFS *Filesystem) UpdateFileAfter(object ocfl.Object, sourceFS fs.FS, source, dest, area string) error {
	return errors.WithStack(
		extFS.AddFileAfter(object, sourceFS, []string{source}, "", "", ""),
	)

}

func (extFS *Filesystem) DeleteFileAfter(object ocfl.Object, dest string, area string) error {
	return nil
}

func (extFS *Filesystem) NeedNewVersion(object ocfl.Object) (bool, error) {
	return false, nil
}

func (extFS *Filesystem) DoNewVersion(object ocfl.Object) error {
	return nil
}

func (extFS *Filesystem) GetMetadata(object ocfl.Object) (map[string]any, error) {
	var err error
	var result = map[string]any{}

	inventory := object.GetInventory()
	manifest := inventory.GetManifest()
	path2digest := map[string]string{}
	for checksum, names := range manifest {
		for _, name := range names {
			path2digest[name] = checksum
		}
	}
	for v := range inventory.GetVersions() {
		var data []byte
		if buf, ok := extFS.buffer[v]; ok && buf.Len() > 0 {
			//		if v == inventory.GetHead() && sl.buffer.Len() > 0 {
			// need a new reader on the buffer
			reader := brotli.NewReader(bytes.NewBuffer(buf.Bytes()))
			data, err = io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read buffer for '%s' '%s'", object.GetID(), v)
			}
		} else {
			data, err = ocfl.ReadJsonL(object, "filesystem", v, extFS.FilesystemConfig.Compress, extFS.StorageType, extFS.StorageName, extFS.fsys)
			if err != nil {
				continue
				// return nil, errors.Wrapf(err, "cannot read jsonl for '%s' version '%s'", object.GetID(), v)
			}
		}

		reader := bytes.NewReader(data)
		r := bufio.NewScanner(reader)
		r.Buffer(make([]byte, 128*1024), 16*1024*1024)
		r.Split(bufio.ScanLines)
		lines := []*fileSystemLine{}
		for r.Scan() {
			lineStr := r.Text()
			var meta = &fileSystemLine{}
			if err := json.Unmarshal([]byte(lineStr), &meta); err != nil {
				return nil, errors.Wrapf(err, "cannot unmarshal line from for '%s' %s - [%s]", object.GetID(), v, lineStr)
			}
			lines = append(lines, meta)
		}
		if err := inventory.IterateStateFiles(v, func(internal, external, digest string) error {
			for _, line := range lines {
				if line.Path == external {
					if _, ok := result[digest]; !ok {
						result[digest] = []*fileSystemLine{}
					}
					result[digest] = append(result[digest].([]*fileSystemLine), line)
				}
			}
			return nil
		}); err != nil {
			return nil, errors.Wrapf(err, "cannot iterate state files for '%s' version '%s'", object.GetID(), v)
		}
	}
	return result, nil
}

func (extFS *Filesystem) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func (extFS *Filesystem) UpdateObjectAfter(object ocfl.Object) error {
	if extFS.writer == nil {
		return nil
	}
	if err := extFS.writer.Flush(); err != nil {
		return errors.Wrap(err, "cannot flush brotli writer")
	}
	if err := extFS.writer.Close(); err != nil {
		return errors.Wrap(err, "cannot close brotli writer")
	}
	head := object.GetInventory().GetHead()
	if head == "" {
		return errors.Errorf("no head for object '%s'", object.GetID())
	}
	buffer, ok := extFS.buffer[head]
	if !ok {
		return nil
	}
	if err := ocfl.WriteJsonL(
		object,
		"filesystem",
		buffer.Bytes(),
		extFS.FilesystemConfig.Compress,
		extFS.StorageType,
		extFS.StorageName,
		extFS.fsys,
	); err != nil {
		return errors.Wrap(err, "cannot write jsonl")
	}
	return nil
}

func (extFS *Filesystem) SetFS(fsys fs.FS) {
	extFS.fsys = fsys
}

func (extFS *Filesystem) SetParams(params map[string]string) error {
	return nil
}

func (extFS *Filesystem) WriteConfig() error {
	if extFS.fsys == nil {
		return errors.New("no filesystem set")
	}
	err := writefs.WriteFile(extFS.fsys, "config.json", []byte(extFS.GetConfigString()))
	if err != nil {
		return errors.Wrap(err, "cannot write config.json")
	}
	return nil

}

func (extFS *Filesystem) GetConfigString() string {
	str, _ := json.MarshalIndent(extFS.FilesystemConfig, "", "  ")
	return string(str)
}

func (extFS *Filesystem) IsRegistered() bool {
	return false
}

func (extFS *Filesystem) GetName() string {
	return FilesystemName
}

var (
	_ ocfl.Extension              = &Filesystem{}
	_ ocfl.ExtensionObjectChange  = &Filesystem{}
	_ ocfl.ExtensionContentChange = &Filesystem{}
	_ ocfl.ExtensionMetadata      = &Filesystem{}
	_ ocfl.ExtensionNewVersion    = &Filesystem{}
)
