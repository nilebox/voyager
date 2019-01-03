package crd

import (
	"encoding/json"
	"testing"

	"github.com/atlassian/voyager/pkg/util/testutil"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadSchemaForCRD(t *testing.T) *spec.Schema {
	crd := ServiceDescriptorCrd()

	// ugly way of loading a schema is to marshal to JSON to get the schema
	// and then back into a schema object
	b, err := json.Marshal(crd.Spec.Validation.OpenAPIV3Schema)
	require.NoError(t, err)

	var schemaProps spec.SchemaProps
	err = json.Unmarshal(b, &schemaProps)
	require.NoError(t, err)

	return &spec.Schema{
		SchemaProps: schemaProps,
	}
}

func runCRDTestCase(t *testing.T, schema *spec.Schema, filename string, errorMsg string) {
	srcFile, err := testutil.LoadFileFromTestData(filename)
	require.NoError(t, err)

	var testData interface{}
	err = yaml.Unmarshal(srcFile, &testData)
	require.NoError(t, err)

	err = validate.AgainstSchema(schema, testData, strfmt.Default)

	if errorMsg == "" {
		assert.NoError(t, err)
		return
	}

	require.Error(t, err)

	// validation errors are returned as a composite error
	ce, ok := err.(*errors.CompositeError)
	require.True(t, ok)

	assert.EqualError(t, ce.Errors[0], errorMsg)
}

func TestServiceDescriptorSchema(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		errorMsg string
	}{
		{
			"test_crd_requires_resource_group_locations",
			"spec.resourceGroups.locations in body is required",
		},
		{
			"test_crd_allows_no_resource_group_resources",
			"",
		},
		{
			"test_crd_allows_scope_with_account_no_label",
			"",
		},
	}

	schema := loadSchemaForCRD(t)

	for _, tc := range cases {
		filename := tc.name + ".yml" // files are all yaml files
		t.Run(filename, func(t *testing.T) {
			runCRDTestCase(t, schema, filename, tc.errorMsg)
		})
	}

}