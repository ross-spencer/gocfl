package extension

import (
	"io/fs"
	"regexp"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
)

// ROCrateFileName ...
const ROCrateFileName = "NNNN-ro-crate"

// RoCrateFileDescription ...
const RoCrateFileDescription = "Description for RO-Crate extension"

// registered ...
const registered = false

// ROCrateFileConfig ...
type ROCrateFileConfig struct {
	*ocfl.ExtensionConfig
	// StorageType ...
	StorageType string `json:"storageType"`
	// StorageName ...
	StorageName string `json:"storageName"`
	// MetadataFile ...
	MetadataFile []string `json:"metadataFile"`
	// SupportedSchema ...
	SupportedSchema []string `json:"supportedSchema"`
	// Documentation ...
	Documentation []string `json:"documentation"`
}

// ROCrateFile provides a combination of configuration and other metadata.
type ROCrateFile struct {
	// combination of the config and other metadata.
	*ROCrateFileConfig
	fsys   fs.FS
	stored bool
	info   map[string][]byte
}

// GetROCrateFileParams ...
func GetROCrateFileParams() []*ocfl.ExtensionExternalParam {
	// not implemented.
	return []*ocfl.ExtensionExternalParam{}
}

// NewROCrateFileFS ...
func NewROCrateFileFS(fsys fs.FS) (*ROCrateFile, error) {
	// not implemented.
	return &ROCrateFile{}, nil
}

// NewROCrateFile provides a helper to create a new object that helps us
// to understand the internals of the extension
func NewROCrateFile(config *ROCrateFileConfig) (*ROCrateFile, error) {
	// not implemented.
	return &ROCrateFile{}, nil
}

// Terminate ...
func (sl *ROCrateFile) Terminate() error {
	// not implemented.
	return nil
}

// GetFS ...
func (sl *ROCrateFile) GetFS() fs.FS {
	return sl.fsys
}

func (sl *ROCrateFile) GetConfig() any {
	return sl.ROCrateFileConfig
}

// IsRegistered describes whether this is an official GOCL extension.
func (sl *ROCrateFile) IsRegistered() bool {
	return registered
}

// SetParams allows us to set parameters provided to the extension via
// the config, e.g. CLI (or TOML?)
func (sl *ROCrateFile) SetParams(params map[string]string) error {
	// not implemented.
	return nil
}

// SetFS ...
func (sl *ROCrateFile) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

// GetName returns the name of this extension to the caller.
func (sl *ROCrateFile) GetName() string {
	return ROCrateFileName
}

func (sl *ROCrateFile) WriteConfig() error {
	// not implemented.
	return nil
}

// UpdateObjectBefore (before a new version of an OCFL object is
// created...) TODO...
func (sl *ROCrateFile) UpdateObjectBefore(object ocfl.Object) error {
	// not implemented.
	return nil
}

var xxwindowsPathWithDrive = regexp.MustCompile("^/[a-zA-Z]:")

// UpdateObjectAfter (after all content to the new version is written)
// TODO...
func (sl *ROCrateFile) UpdateObjectAfter(object ocfl.Object) error {
	// not implemented.
	return nil
}

// GetMetadata (is called by any tool, which wants to report about
// content) TODO...
func (sl *ROCrateFile) GetMetadata(object ocfl.Object) (map[string]any, error) {
	// not implemented.
	return nil, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension             = &ROCrateFile{}
	_ ocfl.ExtensionObjectChange = &ROCrateFile{}
	_ ocfl.ExtensionMetadata     = &ROCrateFile{}
)
