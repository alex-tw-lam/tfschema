package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alex-tw-lam/tfschema/internal/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEnd converts every <category>/<scenario>/test.tf and compares the
// generated schema with the neighbouring test.schema.json.
func TestEndToEnd(t *testing.T) {
	terraformFiles, err := filepath.Glob("./*/*/test.tf")
	require.NoError(t, err)
	require.NotEmpty(t, terraformFiles)

	for _, tfFile := range terraformFiles {
		name := filepath.Base(filepath.Dir(tfFile))
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schema, err := converter.ConvertFile(tfFile)
			require.NoError(t, err, "Failed to convert Terraform file")

			expectedJSON, err := os.ReadFile(filepath.Join(filepath.Dir(tfFile), "test.schema.json"))
			require.NoError(t, err, "Failed to read expected schema file")

			generatedJSON, err := json.MarshalIndent(schema, "", "  ")
			require.NoError(t, err, "Failed to marshal generated schema")

			assert.JSONEq(t, string(expectedJSON), string(generatedJSON),
				"Generated JSON does not match expected JSON")
		})
	}
}
