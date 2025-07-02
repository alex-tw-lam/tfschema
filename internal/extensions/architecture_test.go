package extensions

import (
	"testing"

	"github.com/atwlam/tfschema/internal/converter/types"
	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/atwlam/tfschema/internal/validation"
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtensionRegistryBasicFunctionality tests the core registry functionality
func TestExtensionRegistryBasicFunctionality(t *testing.T) {
	// Create a fresh registry for testing
	registry := &ExtensionRegistry{
		typeConverters:    types.NewTypeConverterRegistry(),
		validationRules:   make([]validation.ParserFunc, 0),
		attributeAppliers: make(map[string]AttributeApplier),
		postProcessors:    make([]PostProcessor, 0),
		preProcessors:     make([]PreProcessor, 0),
	}

	// Test type converter registration
	mockConverter := &mockTypeConverter{}
	registry.RegisterTypeConverter("mock", mockConverter)

	retrievedConverter, err := registry.GetTypeConverterRegistry().Get("mock")
	require.NoError(t, err)
	assert.Equal(t, mockConverter, retrievedConverter)

	// Test validation rule registration
	mockRuleParser := func(expr hcl.Expression, varName string) (validation.Rule, []string, error) {
		return &mockValidationRule{}, []string{}, nil
	}
	registry.RegisterValidationRule(mockRuleParser)

	parsers := registry.GetValidationRuleParsers()
	assert.Len(t, parsers, 1)

	// Test attribute applier registration
	mockApplier := &mockAttributeApplier{name: "test"}
	registry.RegisterAttributeApplier(mockApplier)

	appliers := registry.GetAttributeAppliers()
	assert.Contains(t, appliers, "test")
	assert.Equal(t, mockApplier, appliers["test"])
}

// TestProcessorPriority tests that processors are sorted by priority
func TestProcessorPriority(t *testing.T) {
	registry := &ExtensionRegistry{
		typeConverters:    types.NewTypeConverterRegistry(),
		validationRules:   make([]validation.ParserFunc, 0),
		attributeAppliers: make(map[string]AttributeApplier),
		postProcessors:    make([]PostProcessor, 0),
		preProcessors:     make([]PreProcessor, 0),
	}

	// Register processors with different priorities
	proc1 := &mockPostProcessor{name: "high", priority: 100}
	proc2 := &mockPostProcessor{name: "low", priority: 10}
	proc3 := &mockPostProcessor{name: "medium", priority: 50}

	registry.RegisterPostProcessor(proc1)
	registry.RegisterPostProcessor(proc2)
	registry.RegisterPostProcessor(proc3)

	// Should be sorted by priority (low to high)
	processors := registry.GetPostProcessors()
	require.Len(t, processors, 3)
	assert.Equal(t, "low", processors[0].Name())
	assert.Equal(t, "medium", processors[1].Name())
	assert.Equal(t, "high", processors[2].Name())
}

// TestLegacyBridge tests backward compatibility
func TestLegacyBridge(t *testing.T) {
	// Test legacy type converter registration
	mockConverter := &mockTypeConverter{}
	RegisterLegacyTypeConverter("legacy", mockConverter)

	registry := GetLegacyTypeConverterRegistry()
	retrievedConverter, err := registry.Get("legacy")
	require.NoError(t, err)
	assert.Equal(t, mockConverter, retrievedConverter)

	// Test legacy validation rule registration
	mockRuleParser := func(expr hcl.Expression, varName string) (validation.Rule, []string, error) {
		return &mockValidationRule{}, []string{}, nil
	}
	RegisterLegacyValidationRule(mockRuleParser)

	parsers := GetLegacyValidationRuleParsers()
	assert.NotEmpty(t, parsers)

	// Test legacy attribute applier registration
	mockApplyFunc := func(schema *jsonschema.Schema, attribute *hcl.Attribute) error {
		return nil
	}
	RegisterLegacyAttributeApplier("legacy_attr", mockApplyFunc)

	appliers := GetLegacyAttributeAppliers()
	assert.Contains(t, appliers, "legacy_attr")
}

// TestExtensionInterface tests the complete extension interface
func TestExtensionInterface(t *testing.T) {
	registry := &ExtensionRegistry{
		typeConverters:    types.NewTypeConverterRegistry(),
		validationRules:   make([]validation.ParserFunc, 0),
		attributeAppliers: make(map[string]AttributeApplier),
		postProcessors:    make([]PostProcessor, 0),
		preProcessors:     make([]PreProcessor, 0),
	}

	ext := &mockExtension{}
	err := registry.RegisterExtension(ext)
	require.NoError(t, err)

	// Verify the extension registered its components
	assert.NotEmpty(t, registry.GetValidationRuleParsers())
	assert.NotEmpty(t, registry.GetAttributeAppliers())
}

// Mock implementations for testing

type mockTypeConverter struct{}

func (m *mockTypeConverter) Convert(expr hcl.Expression) (*jsonschema.Schema, error) {
	return &jsonschema.Schema{Type: "mock"}, nil
}

type mockValidationRule struct{}

func (m *mockValidationRule) Apply(schema *jsonschema.Schema) error {
	schema.Description = "mock validation applied"
	return nil
}

type mockAttributeApplier struct {
	name string
}

func (m *mockAttributeApplier) Name() string {
	return m.name
}

func (m *mockAttributeApplier) Apply(schema interface{}, attribute interface{}) error {
	return nil
}

type mockPostProcessor struct {
	name     string
	priority int
}

func (m *mockPostProcessor) Name() string {
	return m.name
}

func (m *mockPostProcessor) Process(schema interface{}) error {
	return nil
}

func (m *mockPostProcessor) Priority() int {
	return m.priority
}

type mockExtension struct{}

func (m *mockExtension) Info() ExtensionInfo {
	return ExtensionInfo{
		Name:        "MockExtension",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "A mock extension for testing",
		Type:        "test",
	}
}

func (m *mockExtension) Register(registry *ExtensionRegistry) error {
	// Register some components
	registry.RegisterValidationRule(func(expr hcl.Expression, varName string) (validation.Rule, []string, error) {
		return &mockValidationRule{}, []string{}, nil
	})

	registry.RegisterAttributeApplier(&mockAttributeApplier{name: "mock_ext"})

	return nil
}
