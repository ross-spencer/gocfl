package rocrate

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

// TestContexts ensures that we can open an empty JSON-LD as expected
// and reason about it.
func TestContexts(t *testing.T) {

	expectedContext := "https://w3id.org/ro/crate/1.1/context"

	simpleContextRes := bytes.NewBuffer(simpleContext)
	res, err := ProcessMetadataStream(simpleContextRes)

	if err != nil {
		t.Errorf("error processing simpleContext: %s", err)
	}
	if res.Context() != expectedContext {
		t.Errorf("context wasn't read successfully, got: '%s'", res.Context())
	}
	if len(res.Graph) != 0 {
		t.Errorf("expecting empty graph, got graph len: '%d'", len(res.Graph))
	}

	complexContextRes := bytes.NewBuffer(complexContext)
	res, err = ProcessMetadataStream(complexContextRes)

	if err != nil {
		t.Errorf("error processing complexContext: %s", err)
	}
	if res.Context() != expectedContext {
		t.Errorf("context wasn't read successfully, got: '%s'", res.Context())
	}
	if len(res.Graph) != 0 {
		t.Errorf("expecting empty graph, got graph len: '%d'", len(res.Graph))
	}
}

// TestStringVariants tests our ability to decode single-value strings
// or slices. We convert the single-value to a slice so we expect a
// slice array at all times.
func TestStringVariants(t *testing.T) {

	// TODO: ADD TABLE-BASED TESTS FOR OTHER STRING VARIANTS...

	variantTest := bytes.NewBuffer(nameTest)
	res, _ := ProcessMetadataStream(variantTest)

	testID := res.Graph[0].ID
	if testID != "test1" {
		t.Errorf("test ID is incorrect: %s", testID)
	}
	value := res.Graph[0].Name.Value()
	compare := []string{"name one", "name two"}
	if !slices.Equal(value, compare) {
		t.Errorf(
			"string variant: '%v' result doesn't match expected: '%v'",
			value,
			compare,
		)
	}

	testID = res.Graph[1].ID
	if testID != "test2" {
		t.Errorf("test ID is incorrect: %s", testID)
	}
	value = res.Graph[1].Name.Value()
	compare = []string{"name one"}
	if !slices.Equal(value, compare) {
		t.Errorf(
			"string variant: '%v' result doesn't match expected: '%v'",
			value,
			compare,
		)
	}

	// TODO: variant table.
	// nameTest
	// authorTest
	// typeTest
	// keywordTest
}

// TestPrimitiveVariants tests the conversion of single-values to
// a slice of primtiives.
func TestPrimitiveVariants(t *testing.T) {
	variantTest := bytes.NewBuffer(authorTest)
	res, _ := ProcessMetadataStream(variantTest)
	testID := res.Graph[0].ID
	if testID != "test1" {
		t.Errorf("test ID is incorrect: %s", testID)
	}
	value := res.Graph[0].Author.Value()
	compare := []nodeIdentifier{
		nodeIdentifier{"https://orcid.org/1234-0003-1974-0000"},
		nodeIdentifier{"#Yann"},
		nodeIdentifier{"#Organization-SMUC"},
	}
	for idx, v := range value {
		if v.ID != compare[idx].ID {
			t.Errorf(
				"string variant: '%v' result doesn't match expected: '%v'",
				v,
				compare[idx],
			)
		}
	}
	testID = res.Graph[1].ID
	if testID != "test2" {
		t.Errorf("test ID is incorrect: %s", testID)
	}
	value = res.Graph[1].Author.Value()
	compare = []nodeIdentifier{
		nodeIdentifier{"https://orcid.org/0000-0003-1974-1234"},
	}
	for idx, v := range value {
		if v.ID != compare[idx].ID {
			t.Errorf(
				"string variant: '%v' result doesn't match expected: '%v'",
				v,
				compare[idx],
			)
		}
	}
}

func TestMetadata(t *testing.T) {
	// TODO: for test in testTable.
	// afternoonDrinks
	// carpentriesCrate
	// specCrate
	// galaxyCrate
}

func TestNothing(t *testing.T) {
	fmt.Println("")
}
