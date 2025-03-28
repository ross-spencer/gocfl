// Package rocrate enables simple processing of ro-crate data so that
// it can be remapped in custome metadata.
//
// For additional reference for some of the structural information in
// this module, including cardinality, please look at crate-o:
//
//   - https://language-research-technology.github.io/crate-o
package rocrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

var versions []string = []string{
	"https://w3id.org/ro/crate/1.1/context",
}

type rocrateMeta struct {
	LDContext any     `json:"@context"`
	Graph     []graph `json:"@graph"`
}

type graph struct {
	ID               string                 `json:"@id"`
	Type             *StringOrSlice         `json:"@type,omitempty"`
	About            *NodeIdentifierOrSlice `json:"about,omitempty"`
	Affiliation      *NodeIdentifierOrSlice `json:"affiliation,omitempty"`
	Author           *NodeIdentifierOrSlice `json:"author,omitempty"`
	Conforms         *NodeIdentifierOrSlice `json:"conformsTo,omitempty"`
	ContentLocation  *NodeIdentifierOrSlice `json:"contentLocation,omitempty"`
	ContentURL       string                 `json:"contentUrl,omitempty"`
	DatePublished    string                 `json:"datePublished,omitempty"`
	Description      string                 `json:"description,omitempty"`
	EncodingFormat   string                 `json:"enncodingFormat,omitempty"`
	FamilyName       string                 `json:"familyName,omitempty"`
	Funder           *NodeIdentifierOrSlice `json:"funder,omitempty"`
	GivenName        string                 `json:"givenName,omitempty"`
	HasPart          *NodeIdentifierOrSlice `json:"hasPart,omitempty"`
	Identifier       string                 `json:"identifier,omitempty"`
	Keywords         *StringOrSlice         `json:"keywords,omitempty"`
	License          *NodeIdentifierOrSlice `json:"license,omitempty"`
	Latitude         string                 `json:"latitude,omitempty"`
	Longitude        string                 `json:"longitude,omitempty"`
	Name             *StringOrSlice         `json:"name,omitempty"`
	Publisher        *NodeIdentifierOrSlice `json:"publisher,omitempty"`
	TemporalCoverage string                 `json:"temporalCoverage,omitempty"`
	URL              string                 `json:"url,omitempty"`
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
		return fmt.Sprintf("cannot determine @context from json-ld input")
	}
	context := (rcContext[0].(string))
	// vocab
	_ = rcContext[1].(map[string]interface{})
	return context
}

// rocrateSummary provides a summary structure we can access safefly
// to reason about the data in a ro-crate.
type rocrateSummary struct {

	// graph[1]
	ID string
	// graph[1]
	Name []string
	// graph[1]
	Type []string
	// graph[1]
	Description string
	// graph[1]
	DatePublished string
	// graph[1]
	Author []string
	// graph[1]
	License string
	// graph[1]
	HasPart []string
	// graph[1]
	ContentURL string
	// graph[1]
	Keywords []string
	// graph[1]
	Publisher []string
	// graph[0]
	// Referenced by refers to sections of the RO-CRATE that reference
	// this summary.
	About string
}

// newSummary creates a new ro-crate summary to provide safe access
// from the caller.
func newSummary() rocrateSummary {
	return rocrateSummary{
		"",
		[]string{},
		[]string{},
		"",
		"",
		[]string{},
		"",
		[]string{},
		"",
		[]string{},
		[]string{},
		"",
	}
}

func (rcMeta rocrateMeta) Summary() (rocrateSummary, error) {

	if len(rcMeta.Graph) == 0 {
		return rocrateSummary{}, fmt.Errorf("ro-crate-metadata.json is empty")
	}
	if len(rcMeta.Graph) == 1 {
		return rocrateSummary{}, fmt.Errorf("ro-crate-metadata.json is non-conformant")
	}
	summary := newSummary()
	summary.ID = rcMeta.Graph[1].ID
	if rcMeta.Graph[1].Name != nil {
		summary.Name = rcMeta.Graph[1].Name.Value()
	}
	if rcMeta.Graph[1].Type != nil {
		summary.Type = rcMeta.Graph[1].Type.Value()
	}
	summary.Description = rcMeta.Graph[1].Description
	summary.DatePublished = rcMeta.Graph[1].DatePublished
	if rcMeta.Graph[1].Author != nil {
		summary.Author = rcMeta.Graph[1].Author.StringSlice()
	}
	if rcMeta.Graph[1].License != nil {
		license := rcMeta.Graph[1].License.StringSlice()
		if len(license) != 0 {
			summary.License = license[0]
		}
	}
	if rcMeta.Graph[1].Author != nil {
		summary.Author = rcMeta.Graph[1].Author.StringSlice()
	}
	if rcMeta.Graph[1].HasPart != nil {
		summary.HasPart = rcMeta.Graph[1].HasPart.StringSlice()
	}
	summary.ContentURL = rcMeta.Graph[1].URL
	if rcMeta.Graph[1].Keywords != nil {
		summary.Keywords = rcMeta.Graph[1].Keywords.Value()
	}
	if rcMeta.Graph[1].Publisher != nil {
		summary.Publisher = rcMeta.Graph[1].Publisher.StringSlice()
	}
	if rcMeta.Graph[0].About != nil {
		about := rcMeta.Graph[0].About.StringSlice()
		if len(about) != 0 {
			summary.About = about[0]
		}
	}
	return summary, nil
}

func (rcMeta rocrateMeta) String() string {
	if len(rcMeta.Graph) == 0 {
		return fmt.Sprintf("ro-crate-metadata.json is empty")
	}
	if len(rcMeta.Graph) == 1 {
		return fmt.Sprintf("ro-crate-metadata.json is non-conformant")
	}
	out := fmt.Sprintf(`
Type: %s
ID: %s
Identifier: %s
Published: %s
Name: %s`,
		rcMeta.Graph[0].Type,
		rcMeta.Graph[0].ID,
		rcMeta.Graph[1].Identifier,
		rcMeta.Graph[1].DatePublished,
		rcMeta.Graph[1].Name,
	)
	return strings.TrimSpace(out)
}

/* String-slice type and handler.

For more info on the type handling below:

StringOrSlice:
   https://gitlab.com/flimzy/talks/-/blob/master/2020/go-json/string-or-array.go

*/

// StringOrSlice represents a type that can interpret both single-value
// strings or slices of strings.
type StringOrSlice []string

// Implement Unmarshal for the StringOrSlice type.
func (s *StringOrSlice) UnmarshalJSON(d []byte) error {
	if d[0] == '"' {
		var v string
		err := json.Unmarshal(d, &v)
		*s = StringOrSlice{v}
		return err
	}
	var v []string
	err := json.Unmarshal(d, &v)
	*s = StringOrSlice(v)
	return err
}

// Return the StringOrSlice value as something sensible.
func (s StringOrSlice) Value() []string {
	return s
}

// String provides a stringer method for this type. It might not
// be needed eventually.
func (s StringOrSlice) String() string {
	var out string = "["
	for _, v := range s {
		out = fmt.Sprintf("%s%s; ", out, v)
	}
	out = fmt.Sprintf("%s]", strings.TrimSpace(out))
	return out
}

// Node-identifier handlers. These seem to only contain relative
// links in the RO-CRATE specification.

// nodePrimitive look like they only contain links. relative, or
// absolute, in RO-CRATE metadata. They can be single-value objects
// or slices of objects.
type nodeIdentifier struct {
	ID string `json:"@id"`
}

type NodeIdentifierOrSlice []nodeIdentifier

func (s *NodeIdentifierOrSlice) UnmarshalJSON(d []byte) error {
	//fmt.Println(d[0], '{', '[', string(d))
	if d[0] == '{' {
		var v nodeIdentifier
		err := json.Unmarshal(d, &v)
		*s = NodeIdentifierOrSlice{v}
		return err
	}
	//fmt.Println(d[0], '"', '{', '[')
	var v []nodeIdentifier
	err := json.Unmarshal(d, &v)
	*s = NodeIdentifierOrSlice(v)
	return err
}

func (s NodeIdentifierOrSlice) Value() []nodeIdentifier {
	return s
}

func (s NodeIdentifierOrSlice) StringSlice() []string {
	var res []string
	for _, v := range s {
		res = append(res, v.ID)
	}
	return res
}

// ProcessMetadataStream enables processing of ro-crate-metadata.json
// and return in the simple structs made available in this package.
func ProcessMetadataStream(meta io.Reader) (rocrateMeta, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, meta)
	if err != nil {
		return rocrateMeta{}, err
	}
	res := rocrateMeta{}
	json.Unmarshal(buf.Bytes(), &res)
	//j, _ := json.MarshalIndent(res, "", "   ")
	if !slices.Contains(versions, res.Context()) {
		return rocrateMeta{}, fmt.Errorf("cannot provess this version")
	}
	//fmt.Println(j)
	return res, nil
}
