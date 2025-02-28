package extension

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"strings"

	"emperror.dev/errors"

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
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &ROCrateFileConfig{
		ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: ROCrateFileName},
		StorageType:     "extension",
		StorageName:     "metadata",
		MetadataFile:    []string{"ro-crate-metadata.json", "ro-crate-metadata.jsonld"},
		SupportedSchema: []string{"https://w3id.org/ro/crate/1.1/context"},
		Documentation:   []string{"ro-crate-preview.html"},
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	rcFile, err := NewROCrateFile(config)
	return rcFile, err
}

// NewROCrateFile provides a helper to create a new object that helps us
// to understand the internals of the extension
func NewROCrateFile(config *ROCrateFileConfig) (*ROCrateFile, error) {
	rcFile := &ROCrateFile{
		ROCrateFileConfig: config,
		info:              map[string][]byte{},
	}
	// check internal extension name is correct..
	if config.ExtensionName != rcFile.GetName() {
		return nil, errors.New(
			fmt.Sprintf(
				"invalid extension name'%s'for extension %s",
				config.ExtensionName,
				rcFile.GetName(),
			),
		)
	}
	return rcFile, nil
}

// Terminate ...
func (rcFile *ROCrateFile) Terminate() error {
	// not implemented.
	return nil
}

// GetFS ...
func (rcFile *ROCrateFile) GetFS() fs.FS {
	return rcFile.fsys
}

func (rcFile *ROCrateFile) GetConfig() any {
	return rcFile.ROCrateFileConfig
}

// IsRegistered describes whether this is an official GOCL extension.
func (rcFile *ROCrateFile) IsRegistered() bool {
	return registered
}

// SetParams allows us to set parameters provided to the extension via
// the config, e.g. CLI (or TOML?)
func (rcFile *ROCrateFile) SetParams(params map[string]string) error {
	// not implemented.
	return nil
}

// SetFS ...
func (rcFile *ROCrateFile) SetFS(fsys fs.FS, create bool) {
	rcFile.fsys = fsys
}

// GetName returns the name of this extension to the caller.
func (rcFile *ROCrateFile) GetName() string {
	return ROCrateFileName
}

func (rcFile *ROCrateFile) WriteConfig() error {
	// not implemented.
	return nil
}

// UpdateObjectBefore (before a new version of an OCFL object is
// created...) TODO...
func (rcFile *ROCrateFile) UpdateObjectBefore(object ocfl.Object) error {
	// not implemented.
	return nil
}

// UpdateObjectAfter (after all content to the new version is written)
// TODO...
func (rcFile *ROCrateFile) UpdateObjectAfter(object ocfl.Object) error {

	// TODO: this needs to be moved to where we read the ro-crate meta.
	//
	switch strings.ToLower(rcFile.StorageType) {
	case "area":
		// case area.
	case "path":
		// case path.
	case "extension":
		log.Println("ROCRATE: extension...")
		f := "mapping.txt" // TODO: make const, no magic.
		log.Println(f)
		data := []byte("some data about rocrate...")
		if _, err := object.AddReader(
			io.NopCloser(
				bytes.NewBuffer(data),
			),
			[]string{f},
			rcFile.StorageName,
			true,
			false,
		); err != nil {
			log.Println("there was an error")
			return err
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", rcFile.StorageType)
	}
	return nil

}

// GetMetadata (is called by any tool, which wants to report about
// content) TODO...
func (rcFile *ROCrateFile) GetMetadata(object ocfl.Object) (map[string]any, error) {
	// not implemented.
	return nil, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension             = &ROCrateFile{}
	_ ocfl.ExtensionObjectChange = &ROCrateFile{}
	_ ocfl.ExtensionMetadata     = &ROCrateFile{}
)
