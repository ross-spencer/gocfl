package rocrate

import (
	"bytes"
	"fmt"
	"reflect"
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

type stringTest struct {
	label    string
	testData []byte
	compare1 []string
	compare2 []string
}

var stringTests []stringTest = []stringTest{
	stringTest{
		"name",
		nameTest,
		[]string{"name one", "name two"},
		[]string{"name one"},
	},
	stringTest{
		"keyword",
		keywordTest,
		[]string{"kw1", "kw2"},
		[]string{"kw1"},
	},
	stringTest{
		"type",
		typeTest,
		[]string{"Person", "Artist"},
		[]string{"Person"},
	},
}

// getStringSliceValue gives us some way of dynamically accessing
// attributes to avoid a decent amount of code replication.
func getStringSliceValue(data graph, label string) []string {
	switch label {
	case "name":
		return data.Name.Value()
	case "keyword":
		return data.Keywords.Value()
	case "type":
		return data.Type.Value()
	}
	return []string{""}
}

// TestStringVariants tests our ability to decode single-value strings
// or slices. We convert the single-value to a slice so we expect a
// slice array at all times.
func TestStringVariants(t *testing.T) {
	for _, test := range stringTests {
		variantTest := bytes.NewBuffer(test.testData)
		res, _ := ProcessMetadataStream(variantTest)
		testID := res.Graph[0].ID
		if testID != "test1" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value := getStringSliceValue(res.Graph[0], test.label)
		if !slices.Equal(value, test.compare1) {
			t.Errorf(
				"%s: string variant: '%v' result doesn't match expected: '%v'",
				fmt.Sprintf("%s test 1", test.label),
				value,
				test.compare1,
			)
		}
		testID = res.Graph[1].ID
		if testID != "test2" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value = getStringSliceValue(res.Graph[1], test.label)
		if !slices.Equal(value, test.compare2) {
			fmt.Println(test.label, "test 2")
			t.Errorf(
				"%s: string variant: '%v' result doesn't match expected: '%v'",
				fmt.Sprintf("%s test 2", test.label),
				value,
				test.compare2,
			)
		}
	}
}

type nodeTest struct {
	label    string
	testData []byte
	compare1 []nodeIdentifier
	compare2 []nodeIdentifier
}

var nodeTests []nodeTest = []nodeTest{
	nodeTest{
		"name",
		authorTest,
		[]nodeIdentifier{
			nodeIdentifier{"https://orcid.org/1234-0003-1974-0000"},
			nodeIdentifier{"#Yann"},
			nodeIdentifier{"#Organization-SMUC"},
		},
		[]nodeIdentifier{
			nodeIdentifier{"https://orcid.org/0000-0003-1974-1234"},
		},
	},
}

// getNodeIdentifierSliceValue allows us to get values more dynamically
// from the nodeIdentifier tests.
func getNodeIdentifierSliceValue(data graph, label string) []nodeIdentifier {
	switch label {
	case "author":
		return data.Author.Value()
	}
	return []nodeIdentifier{}
}

// TestNodeIdentifierVariants tests the conversion of single-values to
// a slice of nodeIdentifiers.
func TestNodeIdentifierVariants(t *testing.T) {
	for _, test := range nodeTests {
		variantTest := bytes.NewBuffer(test.testData)
		res, _ := ProcessMetadataStream(variantTest)
		testID := res.Graph[0].ID
		if testID != "test1" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value := getNodeIdentifierSliceValue(res.Graph[0], "author")
		for idx, v := range value {
			if v.ID != test.compare1[idx].ID {
				t.Errorf(
					"%s: string variant: '%v' result doesn't match expected: '%v'",
					fmt.Sprintf("%s test 1", test.label),
					v,
					test.compare1[idx],
				)
			}
		}
		testID = res.Graph[1].ID
		if testID != "test2" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value = getNodeIdentifierSliceValue(res.Graph[1], "author")
		for idx, v := range value {
			if v.ID != test.compare2[idx].ID {
				t.Errorf(
					"%s: string variant: '%v' result doesn't match expected: '%v'",
					fmt.Sprintf("%s test 2", test.label),
					v,
					test.compare2[idx],
				)
			}
		}
	}
}

type metadataTest struct {
	label    string
	testData []byte
}

var metadataTests []metadataTest = []metadataTest{
	metadataTest{
		"empty",
		emptyCrate,
	},
	metadataTest{
		"afternoon drinks",
		afternoonDrinks,
	},
	metadataTest{
		"carpentries",
		carpentriesCrate,
	},
	metadataTest{
		"spec",
		specCrate,
	},
	metadataTest{
		"galaxy",
		galaxyCrate,
	},
}

// TestNewSummary ensures new summary is as safe as possible.
func TestNewSummary(t *testing.T) {
	summary := newSummary()
	structType := reflect.TypeOf(summary)
	structVal := reflect.ValueOf(summary)
	fieldNum := structVal.NumField()
	for i := 0; i < fieldNum; i++ {
		field := structVal.Field(i)
		if fmt.Sprintf("%s", field.Type()) == "string" {
			continue
		}
		fieldName := structType.Field(i).Name
		isSet := field.IsValid() && !field.IsZero()
		if !isSet {
			t.Errorf("summary constructor isn't setting: %s", fieldName)
		}
	}
}

// TestNewGocflSummary ensures new gocfl summary is as safe as possible.
func TestNewGocflSummary(t *testing.T) {
	summary := newGocflSummary()
	structType := reflect.TypeOf(summary)
	structVal := reflect.ValueOf(summary)
	fieldNum := structVal.NumField()
	for i := 0; i < fieldNum; i++ {
		field := structVal.Field(i)
		if fmt.Sprintf("%s", field.Type()) == "string" {
			continue
		}
		fieldName := structType.Field(i).Name
		isSet := field.IsValid() && !field.IsZero()
		if !isSet {
			t.Errorf("summary constructor isn't setting: %s", fieldName)
		}
	}
}

// TestMetadata provides more generic testing of the test data within
// this package.
func TestMetadata(t *testing.T) {
	for _, test := range metadataTests {
		variantTest := bytes.NewBuffer(test.testData)
		res, _ := ProcessMetadataStream(variantTest)
		fmt.Println(test.label)
		fmt.Println(res.Summary())
	}
}

// TestGOCFLMetadata provides more generic testing the summary output
// of the GOCFL struct.
func TestGOCFLMetadata(t *testing.T) {
	for _, test := range metadataTests {
		variantTest := bytes.NewBuffer(test.testData)
		res, _ := ProcessMetadataStream(variantTest)
		fmt.Println(test.label
		fmt.Println(res.GOCFLSummary())
	}
}
