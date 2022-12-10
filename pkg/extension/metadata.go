package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const MetadataName = "NNNN-metadata"
const MetadataDescription = "technical metadata for all files"

func GetMetadataParams() []ocfl.ExtensionExternalParam {
	return []ocfl.ExtensionExternalParam{
		{
			Functions:   []string{"add", "update", "create"},
			Param:       "source",
			File:        "Source",
			Description: "url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json",
		},
		{
			Functions:   []string{"extract", "objectextension"},
			Param:       "target",
			File:        "Target",
			Description: "url with metadata target folder",
		},
	}
}

type MetadataConfig struct {
	*ocfl.ExtensionConfig
	Versioned bool `json:"versioned"`
}
type Metadata struct {
	*MetadataConfig
	metadataSource *url.URL
	fs             ocfl.OCFLFS
}

func NewMetadataFS(fs ocfl.OCFLFS, params map[string]string) (*Metadata, error) {
	fp, err := fs.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MetadataConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewMetadata(config, params)
}
func NewMetadata(config *MetadataConfig, params map[string]string) (*Metadata, error) {
	sl := &Metadata{MetadataConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	if params != nil {
		if urlString, ok := params["source"]; ok {
			u, err := url.Parse(urlString)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid url '%s'", urlString)
			}
			sl.metadataSource = u
		}
	}
	if sl.metadataSource == nil {
		return nil, errors.Errorf("no metadata-source for extension '%s'", MetadataName)
	}
	return sl, nil
}

func (sl *Metadata) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *Metadata) GetName() string { return MetadataName }
func (sl *Metadata) WriteConfig() error {
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := sl.fs.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.MetadataConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func (sl *Metadata) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func downloadFile(u string) ([]byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get '%s'", u)
	}
	defer resp.Body.Close()
	metadata, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body from '%s'", u)
	}
	return metadata, nil
}

var windowsPathWithDrive = regexp.MustCompile("^/[a-zA-Z]:")

func (sl *Metadata) UpdateObjectAfter(object ocfl.Object) error {
	var err error
	inventory := object.GetInventory()
	if inventory == nil {
		return errors.New("no inventory available")
	}
	if sl.metadataSource == nil {
		// only a problem, if first version
		if len(inventory.GetVersionStrings()) < 2 {
			return errors.New("no metadata source configured")
		}
		return nil
	}
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	var rc io.ReadCloser
	switch sl.metadataSource.Scheme {
	case "http":
		fname := strings.Replace(sl.metadataSource.String(), "$ID", object.GetID(), -1)
		resp, err := http.Get(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot get '%s'", fname)
		}
		rc = resp.Body
	case "https":
		fname := strings.Replace(sl.metadataSource.String(), "$ID", object.GetID(), -1)
		resp, err := http.Get(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot get '%s'", fname)
		}
		rc = resp.Body
	case "file":
		fname := strings.Replace(sl.metadataSource.Path, "$ID", object.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		if windowsPathWithDrive.Match([]byte(fname)) {
			fname = strings.TrimLeft(fname, "/")
		}
		rc, err = os.Open(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot open '%s'", fname)
		}
	case "":
		fname := strings.Replace(sl.metadataSource.Path, "$ID", object.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		rc, err = os.Open(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot open '%s'", fname)
		}
	default:
		return errors.Errorf("url scheme '%s' not supported", sl.metadataSource.Scheme)
	}

	//todo: clear old files

	// complex writes to prevent simultaneous writes on filesystems, which do not support that
	targetBase := filepath.Base(sl.metadataSource.Path)
	w2, err := sl.fs.Create(targetBase)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", targetBase)
	}

	allTargets := []io.Writer{w2}
	var buf = bytes.NewBuffer(nil)
	if sl.Versioned {
		allTargets = append(allTargets, buf)
	}
	csWriter := checksum.NewChecksumWriter([]checksum.DigestAlgorithm{inventory.GetDigestAlgorithm()})

	mw := io.MultiWriter(allTargets...)
	digests, err := csWriter.Copy(mw, rc)
	if err != nil {
		w2.Close()
		return errors.Wrap(err, "cannot write data")
	}
	w2.Close()

	digest, ok := digests[inventory.GetDigestAlgorithm()]
	if !ok {
		return errors.Wrapf(err, "digest '%s' not created", inventory.GetDigestAlgorithm())
	}

	targetBaseSidecar := fmt.Sprintf("%s.%s", targetBase, inventory.GetDigestAlgorithm())
	w2Sidecar, err := sl.fs.Create(targetBaseSidecar)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", targetBaseSidecar)
	}
	if _, err := io.WriteString(w2Sidecar, fmt.Sprintf("%s %s", digest, targetBase)); err != nil {
		w2Sidecar.Close()
		return errors.Wrapf(err, "cannot write to sidecar '%s'", targetBaseSidecar)
	}
	w2Sidecar.Close()

	if sl.Versioned {
		targetVersioned := fmt.Sprintf("%s/%s", inventory.GetHead(), targetBase)
		w, err := sl.fs.Create(targetVersioned)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%s'", targetVersioned)
		}
		if _, err := io.Copy(w, buf); err != nil {
			w.Close()
			return errors.Wrapf(err, "cannot write data to '%s'", targetVersioned)
		}
		w.Close()
		targetVersionedSidecar := fmt.Sprintf("%s.%s", targetVersioned, inventory.GetDigestAlgorithm())
		wSidecar, err := sl.fs.Create(targetVersionedSidecar)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%s'", targetVersionedSidecar)
		}
		if _, err := io.WriteString(wSidecar, fmt.Sprintf("%s %s", digest, targetVersioned)); err != nil {
			wSidecar.Close()
			return errors.Wrapf(err, "cannot write to sidecar '%s'", targetVersionedSidecar)
		}
		wSidecar.Close()
	}

	return nil
}

// check interface satisfaction
var (
	_ ocfl.Extension             = &Metadata{}
	_ ocfl.ExtensionObjectChange = &Metadata{}
)
