// Custom view/summary objects for consumers of RO-CRATE metadata.
package rocrate

import (
	"fmt"
)

// gocflSummary provides a summary compatible with gocfl user's
// expectations for the info.json object.
//
/*
	-- VIA ROCRATE
	signature       = id
	title           = name
	description     = description
	created         = datePublished
	sets            = @type
	keywords        = keywords
	licenses        = license

	-- AUDO
	last_changed    = now()

	-- VIA CONFIG
	organisation_id = user.config
	organisation    = user.config
	user            = user.config
	address         = user.config

*/
//
type gocflSummary struct {
	// provided by ro-crate.
	Signature   string
	Title       string
	Description []string
	Created     string
	Sets        []string
	Keywords    []string
	Licenses    []string
	// provided by caller.
	LastChanged    string
	OrganisationID string
	Organisation   string
	User           string
	Address        string
}

// newGocflSummary returns an initialized gocflSummary object for
// maximum safety.
func newGocflSummary() gocflSummary {
	return gocflSummary{
		"",
		"",
		[]string{},
		"",
		[]string{},
		[]string{},
		[]string{},
		// provided by caller.
		"",
		"",
		"",
		"",
		"",
	}
}

func (rcMeta rocrateMeta) GOCFLSummary() (gocflSummary, error) {
	if len(rcMeta.Graph) == 0 {
		return gocflSummary{}, fmt.Errorf("ro-crate-metadata.json is empty")
	}
	if len(rcMeta.Graph) == 1 {
		return gocflSummary{}, fmt.Errorf("ro-crate-metadata.json is non-conformant")
	}
	summary := newGocflSummary()
	summary.Signature = rcMeta.Graph[1].ID
	if rcMeta.Graph[1].Name != nil {
		name := rcMeta.Graph[1].Name.Value()
		if len(name) > 0 {
			summary.Title = rcMeta.Graph[1].Name.Value()[0]
		}
	}
	if rcMeta.Graph[1].Type != nil {
		summary.Sets = rcMeta.Graph[1].Type.Value()
	}
	if rcMeta.Graph[1].Description != nil {
		summary.Description = rcMeta.Graph[1].Description.Value()
	}
	summary.Created = rcMeta.Graph[1].DatePublished
	if rcMeta.Graph[1].License != nil {
		summary.Licenses = rcMeta.Graph[1].License.StringSlice()
	}
	if rcMeta.Graph[1].Keywords != nil {
		summary.Keywords = rcMeta.Graph[1].Keywords.Value()
	}
	return summary, nil
}
