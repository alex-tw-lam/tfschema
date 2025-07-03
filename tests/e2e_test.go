package tests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/alex-tw-lam/tfschema/internal/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEnd(t *testing.T) {
	testCases, err := discoverTestCases("./")
	require.NoError(t, err, "Failed to discover test cases")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// Convert HCL to JSON Schema
			c := converter.New()
			generatedSchema, err := c.ConvertFile(tc.TerraformFile)
			require.NoError(t, err, "Failed to convert Terraform file")

			// Load expected JSON schema
			expectedJSON, err := ioutil.ReadFile(tc.SchemaFile)
			require.NoError(t, err, "Failed to read expected schema file")

			var expectedSchema map[string]interface{}
			err = json.Unmarshal(expectedJSON, &expectedSchema)
			require.NoError(t, err, "Failed to unmarshal expected schema")

			// Convert generated schema to map for comparison
			generatedJSON, err := json.Marshal(generatedSchema)
			require.NoError(t, err, "Failed to marshal generated schema")

			var actualSchema map[string]interface{}
			err = json.Unmarshal(generatedJSON, &actualSchema)
			require.NoError(t, err, "Failed to unmarshal generated schema")

			// Compare the maps
			assert.Equal(t, expectedSchema, actualSchema, "Generated JSON does not match expected JSON")

			if err := runTest(t, filepath.Dir(tc.TerraformFile)); err != nil {
				t.Fatalf("Test case %s failed: %v", tc.TerraformFile, err)
			}
		})
	}
}

type TestCase struct {
	Name          string
	TerraformFile string
	SchemaFile    string
}

func discoverTestCases(rootDir string) ([]TestCase, error) {
	var testCases []TestCase
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "test.tf" {
			dir := filepath.Dir(path)
			testName := filepath.Base(dir)
			testCases = append(testCases, TestCase{
				Name:          testName,
				TerraformFile: path,
				SchemaFile:    filepath.Join(dir, "test.schema.json"),
			})
		}
		return nil
	})
	return testCases, err
}

func runTest(_ *testing.T, _ string) error {
	// ... existing code ...
	return nil
}
