// Package rocrate enables simple processing of ro-crate data so that
// it can be remapped in custome metadata.
package rocrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
)

type rocrateMeta struct {
	LDContext any     `json:"@context"`
	Graph     []graph `json:"@graph"`
}

type graph struct {
	ID               string            `json:"@id"`
	Type             *SliceOrString    `json:"@type,omitempty"`
	About            *SliceOrPrimitive `json:"about,omitempty"`
	Affiliation      *primitive        `json:"affiliation,omitempty"`
	Author           json.RawMessage   `json:"author,omitempty"`
	Conforms         *primitive        `json:"conformsTo,omitempty"`
	ContentLocation  *primitive        `json:"contentLocation,omitempty"`
	ContentURL       string            `json:"contentUrl,omitempty"`
	Date             string            `json:"datePublished,omitempty"`
	Description      string            `json:"description,omitempty"`
	EncodingFormat   string            `json:"enncodingFormat,omitempty"`
	FamilyName       string            `json:"familyName,omitempty"`
	Funderr          *primitive        `json:"funder,omitempty"`
	GivenName        string            `json:"givenName,omitempty"`
	HasPart          *SliceOrPrimitive `json:"hasPart,omitempty"`
	Identifier       string            `json:"identifier,omitempty"`
	Keywords         *SliceOrString    `json:"keywords,omitempty"`
	License          *primitive        `json:"license,omitempty"`
	Latitude         string            `json:"latitude,omitempty"`
	Longitude        string            `json:"longitude,omitempty"`
	Name             *SliceOrPrimitive `json:"name,omitempty"`
	Publisher        *primitive        `json:"publisher,omitempty"`
	TemporalCoverage string            `json:"temporalCoverage,omitempty"`
	URL              string            `json:"url,omitempty"`
}

type rocrateContext struct {
	string
	vocab map[string]string
}

// Context provides a helpter to return the RO-CRATE context from the
// RO-CRATE data structure.
//
// TODO: must be able to cast to rocrateContext some way, but it is
// currently eluding me.
func (rcMeta *rocrateMeta) Context() string {

	// "context":
	// [
	//		https://w3id.org/ro/crate/1.1/context,
	//		map[@vocab:http://schema.org/],
	// ]
	//
	// or
	//
	// "@context": "https://example.com/vocab/context"
	//

	switch rcMeta.LDContext.(type) {
	case string:
		return rcMeta.LDContext.(string)
	default:
		// expect string/map variant
	}
	rcContext, ok := rcMeta.LDContext.([]interface{})
	if !ok {
		panic("TODO: cannot cast")
	}
	context := (rcContext[0].(string))
	// vocab
	_ = rcContext[1].(map[string]interface{})
	return context
}

// SliceOrString: https://gitlab.com/flimzy/talks/-/blob/master/2020/go-json/string-or-array.go

// SliceOrPrimitive ... TODO: type naming.
type SliceOrPrimitive []primitive

func (s *SliceOrPrimitive) UnmarshalJSON(d []byte) error {
	if d[0] == '"' {
		var v primitive
		err := json.Unmarshal(d, &v)
		*s = SliceOrPrimitive{v}
		return err
	}
	var v []primitive
	err := json.Unmarshal(d, &v)
	*s = SliceOrPrimitive(v)
	return err
}

// SliceOrString ... TODO: type naming.
type SliceOrString []string

func (s *SliceOrString) UnmarshalJSON(d []byte) error {
	if d[0] == '"' {
		var v string
		err := json.Unmarshal(d, &v)
		*s = SliceOrString{v}
		return err
	}
	var v []string
	err := json.Unmarshal(d, &v)
	*s = SliceOrString(v)
	return err
}

type primitive struct {
	ID string `json:"@id"`
}

var versions []string = []string{
	"https://w3id.org/ro/crate/1.1/context",
}

// ProcessMetadataStream enables processing of ro-crate-metadata.json
// and return in the simple structs made available in this package.
func ProcessMetadataStream(meta io.Reader) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, meta)
	if err != nil {
		return err
	}
	res := rocrateMeta{}
	json.Unmarshal(buf.Bytes(), &res)
	j, _ := json.MarshalIndent(res, "", "   ")
	if !slices.Contains(versions, res.Context()) {
		return fmt.Errorf("cannot provess this version")
	}
	fmt.Println("we can parse:", string(j))

	// TODO: return data...
	return nil
}
