package extension

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"

	"emperror.dev/errors"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/rocrate"
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
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: ROCrateFileName,
			Functions:     []string{"add", "update", "create"},
			Param:         "enabled",
			Description:   "replace metafile extension functionality if enabled and map RO-CRATE metadata",
			Default:       "false",
		},
	}
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
	// not implemented.
	return nil

}

// GetMetadata (is called by any tool, which wants to report about
// content) TODO...
func (rcFile *ROCrateFile) GetMetadata(object ocfl.Object) (map[string]any, error) {
	// not implemented.
	return nil, nil
}

// findROCrateMeta looks for the RO-CRATE metadata file within the
// objects spplied to the function.
func (rcFile *ROCrateFile) findROCrateMeta(stateFiles []string) bool {
	f := stateFiles[0]
	if f != "data/ro-crate-metadata.json" {
		return false
	}
	return true
}

// copyStream allows StreamObject to make a copy of a reader so that it
// can be given back safely to the caller and other stream functions
// can be performed on the object.
func copyStream(reader io.Reader) (io.Reader, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// writeMetafile ...
func (rcFile *ROCrateFile) writeMetafile(object ocfl.Object, rcMeta string) error {
	log.Println("ROCRATE: extension...")
	mappingFile := "mapping.txt2" // TODO: make const, no magic.
	log.Println(mappingFile)
	data := []byte(rcMeta)
	if _, err := object.AddReader(
		io.NopCloser(
			bytes.NewBuffer(data),
		),
		[]string{mappingFile},
		rcFile.StorageName,
		true,
		false,
	); err != nil {
		log.Println("there was an error")
		return err
	}
	return nil
}

func (rcFile *ROCrateFile) StreamObject(
	object ocfl.Object,
	reader io.Reader,
	stateFiles []string,
	dest string,
) error {
	// TODO: check idiom, this might need to use the object data.
	if !rcFile.findROCrateMeta(stateFiles) {
		return nil
	}
	// copy file so that it can then be sent to another interface to
	// be read. In this case a ro-crate-metadata json reader.
	metaCopy, err := copyStream(reader)
	if err != nil {
		return err
	}
	rocrate.ProcessMetadataStream(metaCopy)
	rcMeta := "this is some data..."
	rcFile.writeMetafile(object, rcMeta)
	return nil
}

// check interface satisfaction
var (
	_ ocfl.Extension             = &ROCrateFile{}
	_ ocfl.ExtensionObjectChange = &ROCrateFile{}
	_ ocfl.ExtensionMetadata     = &ROCrateFile{}
)
