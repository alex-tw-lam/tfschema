// Package extensions provides a unified registry system for all tfschema extensions.
// This allows new functionality to be added by simply creating new files that
// register themselves during package initialization.
package extensions

import (
	"fmt"
	"log"

	"github.com/atwlam/tfschema/internal/converter/types"
	"github.com/atwlam/tfschema/internal/validation"
)

// ExtensionRegistry manages all extension points in tfschema
type ExtensionRegistry struct {
	typeConverters    *types.TypeConverterRegistry
	validationRules   []validation.ParserFunc
	attributeAppliers map[string]AttributeApplier
	postProcessors    []PostProcessor
	preProcessors     []PreProcessor
}

// AttributeApplier defines the interface for applying variable attributes to schemas
type AttributeApplier interface {
	Name() string
	Apply(schema interface{}, attribute interface{}) error
}

// PostProcessor defines operations that run after schema generation
type PostProcessor interface {
	Name() string
	Process(schema interface{}) error
	Priority() int // Lower numbers run first
}

// PreProcessor defines operations that run before schema generation
type PreProcessor interface {
	Name() string
	Process(input interface{}) (interface{}, error)
	Priority() int // Lower numbers run first
}

// ExtensionInfo provides metadata about an extension
type ExtensionInfo struct {
	Name        string
	Version     string
	Author      string
	Description string
	Type        string // "type_converter", "validation_rule", "attribute_applier", etc.
}

// Extension represents a complete extension package
type Extension interface {
	Info() ExtensionInfo
	Register(registry *ExtensionRegistry) error
}

var globalRegistry *ExtensionRegistry

// GetGlobalRegistry returns the singleton extension registry
func GetGlobalRegistry() *ExtensionRegistry {
	if globalRegistry == nil {
		globalRegistry = &ExtensionRegistry{
			typeConverters:    types.NewTypeConverterRegistry(),
			validationRules:   make([]validation.ParserFunc, 0),
			attributeAppliers: make(map[string]AttributeApplier),
			postProcessors:    make([]PostProcessor, 0),
			preProcessors:     make([]PreProcessor, 0),
		}
	}
	return globalRegistry
}

// RegisterTypeConverter adds a type converter to the registry
func (r *ExtensionRegistry) RegisterTypeConverter(name string, converter types.TypeConverter) {
	r.typeConverters.Register(name, converter)
	log.Printf("Registered type converter: %s", name)
}

// RegisterValidationRule adds a validation rule parser
func (r *ExtensionRegistry) RegisterValidationRule(parser validation.ParserFunc) {
	r.validationRules = append(r.validationRules, parser)
	log.Printf("Registered validation rule parser")
}

// RegisterAttributeApplier adds an attribute applier
func (r *ExtensionRegistry) RegisterAttributeApplier(applier AttributeApplier) {
	r.attributeAppliers[applier.Name()] = applier
	log.Printf("Registered attribute applier: %s", applier.Name())
}

// RegisterPostProcessor adds a post-processor
func (r *ExtensionRegistry) RegisterPostProcessor(processor PostProcessor) {
	r.postProcessors = append(r.postProcessors, processor)
	log.Printf("Registered post-processor: %s", processor.Name())
}

// RegisterPreProcessor adds a pre-processor
func (r *ExtensionRegistry) RegisterPreProcessor(processor PreProcessor) {
	r.preProcessors = append(r.preProcessors, processor)
	log.Printf("Registered pre-processor: %s", processor.Name())
}

// RegisterExtension registers a complete extension package
func (r *ExtensionRegistry) RegisterExtension(ext Extension) error {
	info := ext.Info()
	log.Printf("Registering extension: %s v%s by %s", info.Name, info.Version, info.Author)

	if err := ext.Register(r); err != nil {
		return fmt.Errorf("failed to register extension %s: %w", info.Name, err)
	}

	log.Printf("Successfully registered extension: %s", info.Name)
	return nil
}

// GetTypeConverterRegistry returns the type converter registry
func (r *ExtensionRegistry) GetTypeConverterRegistry() *types.TypeConverterRegistry {
	return r.typeConverters
}

// GetValidationRuleParsers returns all registered validation rule parsers
func (r *ExtensionRegistry) GetValidationRuleParsers() []validation.ParserFunc {
	return r.validationRules
}

// GetAttributeAppliers returns all registered attribute appliers
func (r *ExtensionRegistry) GetAttributeAppliers() map[string]AttributeApplier {
	return r.attributeAppliers
}

// GetPostProcessors returns all post-processors sorted by priority
func (r *ExtensionRegistry) GetPostProcessors() []PostProcessor {
	// Sort by priority (lower numbers first)
	sorted := make([]PostProcessor, len(r.postProcessors))
	copy(sorted, r.postProcessors)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority() > sorted[j].Priority() {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// GetPreProcessors returns all pre-processors sorted by priority
func (r *ExtensionRegistry) GetPreProcessors() []PreProcessor {
	// Sort by priority (lower numbers first)
	sorted := make([]PreProcessor, len(r.preProcessors))
	copy(sorted, r.preProcessors)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority() > sorted[j].Priority() {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// ListExtensions returns information about all registered extensions
func (r *ExtensionRegistry) ListExtensions() []ExtensionInfo {
	// This would need to be implemented to track registered extensions
	// For now, return an empty slice
	return []ExtensionInfo{}
}
