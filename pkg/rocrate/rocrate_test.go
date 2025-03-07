package rocrate

import (
	"bytes"
	"testing"
)

// TODO: table-driven testing structures for the data in rocrate_data_test.go.

func TestVariants(t *testing.T) {
	// TODO: variant table.
	// nameTest
	// authorTest
	// typeTest
	// keywordTest
}

func TestMetadata(t *testing.T) {
	// TODO: for test in testTable.
	// afternoonDrinks
	// carpentriesCrate
	// specCrate
	// galaxyCrate
}

func TestContexts(t *testing.T) {

	simpleContextRes := bytes.NewBuffer(simpleContext)
	ProcessMetadataStream(simpleContextRes)

	// TODO: assert.

	complexContextRes := bytes.NewBuffer(complexContext)
	ProcessMetadataStream(complexContextRes)

	// TODO: assert.
}
