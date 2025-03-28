// Custom view/summary objects for consumers of RO-CRATE metadata.
package rocrate

// gocflSummary provides a summary compatible with gocfl user's
// expectations for the info.json object.
type gocflSummary struct {
}

func (rcMeta rocrateMeta) GOCFLSummary() gocflSummary {
	return gocflSummary{}
}
