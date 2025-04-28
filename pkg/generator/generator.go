package generator

import (
	"fmt"
	"strings"

	"github.com/bry-guy/factorio-lsp-plugin/pkg/api"
)

// Generator holds the logic for converting API data to LuaLS definitions.
type Generator struct {
	// Add any necessary configuration or state here
}

// NewGenerator creates a new instance of the Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateDefinitions takes the parsed API data and returns a map of filenames
// to their generated Lua definition content.
func (g *Generator) GenerateDefinitions(runtimeAPI *api.API, prototypeAPI *api.API) (map[string]string, error) {
	definitions := make(map[string]string)

	// --- Runtime API ---
	var runtimeSB strings.Builder
	runtimeSB.WriteString("---@meta\n\n")
	runtimeSB.WriteString("-- Auto-generated Factorio Runtime API definitions\n")
	runtimeSB.WriteString("-- Generated from: https://lua-api.factorio.com/latest/runtime-api.json\n\n")

	// Generate Defines
	// Factorio defines are often nested, so we need a recursive approach.
	runtimeSB.WriteString("-- Defines\n\n")
	for name, define := range runtimeAPI.Defines {
		g.generateDefine(&runtimeSB, name, define, "") // Start recursion
		runtimeSB.WriteString("\n")
	}

	// Generate Concepts (Runtime)
	runtimeSB.WriteString("-- Concepts (Runtime)\n\n")
	for name, concept := range runtimeAPI.Concepts {
		// Concepts can be aliases or complex types, need to handle based on Category and Type structure
		runtimeSB.WriteString(g.generateConcept(name, concept))
		runtimeSB.WriteString("\n")
	}

	// Generate Classes
	runtimeSB.WriteString("-- Classes\n\n")
	for name, class := range runtimeAPI.Classes {
		runtimeSB.WriteString(g.generateClass(name, class))
		runtimeSB.WriteString("\n")
	}

	// Generate Global Objects
	runtimeSB.WriteString("-- Global Objects\n\n")
	for name, global := range runtimeAPI.GlobalObjects {
		runtimeSB.WriteString(g.generateGlobalObject(name, global))
		runtimeSB.WriteString("\n")
	}

	// Generate Events
	// Events are typically handled by documenting the event handler signature
	runtimeSB.WriteString("-- Events\n\n")
	// This part requires careful consideration of how LuaLS handles events.
	// A common approach is to define a global 'script.on_event' overload
	// or define types for event data payloads.
	// For simplicity here, we'll just add comments for now.
	runtimeSB.WriteString("-- Event definitions require specific LuaLS handling, refer to LuaLS documentation.\n")
	runtimeSB.WriteString("-- You might need to define types for event data payloads (e.g., EventData.on_built_entity).\n")
	runtimeSB.WriteString("-- See https://luals.github.io/wiki/definition-files/ for more details.\n")
	// Example:
	// runtimeSB.WriteString("---@class EventData\n") // Base class for event data
	// for name, event := range runtimeAPI.Events {
	// 	// Generate a class or type alias for the event data
	// 	runtimeSB.WriteString(g.generateEventDataClass(name, event))
	// }

	definitions["runtime.lua"] = runtimeSB.String()

	// --- Prototype API ---
	// The Prototype API structure might be slightly different, requiring
	// separate parsing and generation logic. Assuming a similar top-level
	// structure for now, but you might need a separate api.PrototypeAPI struct.
	var prototypeSB strings.Builder
	prototypeSB.WriteString("---@meta\n\n")
	prototypeSB.WriteString("-- Auto-generated Factorio Prototype API definitions\n")
	prototypeSB.WriteString("-- Generated from: https://lua-api.factorio.com/latest/prototype-api.json\n\n")

	// Prototypes API also has Concepts and Defines, potentially with different content
	// Generate Defines (Prototype)
	prototypeSB.WriteString("-- Defines (Prototype)\n\n")
	// Assuming prototypeAPI has a Defines field like runtimeAPI
	if prototypeAPI.Defines != nil {
		for name, define := range prototypeAPI.Defines {
			g.generateDefine(&prototypeSB, name, define, "")
			prototypeSB.WriteString("\n")
		}
	}

	// Generate Concepts (Prototype)
	prototypeSB.WriteString("-- Concepts (Prototype)\n\n")
	// Assuming prototypeAPI has a Concepts field
	if prototypeAPI.Concepts != nil {
		for name, concept := range prototypeAPI.Concepts {
			prototypeSB.WriteString(g.generateConcept(name, concept))
			prototypeSB.WriteString("\n")
		}
	}

	// Generate Prototypes
	// Prototypes themselves are definitions, not runtime objects.
	// You might define types representing each prototype type (e.g., "item", "recipe").
	prototypeSB.WriteString("-- Prototypes\n\n")
	// Assuming prototypeAPI has a Prototypes field (need to add this to api.API or a new struct)
	// for name, prototype := range prototypeAPI.Prototypes {
	// 	prototypeSB.WriteString(g.generatePrototypeClass(name, prototype))
	// 	prototypeSB.WriteString("\n")
	// }
	prototypeSB.WriteString("-- Prototype definitions require specific handling, often defining types for each prototype category.\n")

	definitions["prototype.lua"] = prototypeSB.String()

	return definitions, nil
}

// generateDefine recursively generates LuaLS annotations for Defines.
func (g *Generator) generateDefine(sb *strings.Builder, name string, define api.Define, prefix string) {
	fullName := prefix + name
	sb.WriteString(fmt.Sprintf("---@class %s %s\n", fullName, define.Description))
	sb.WriteString(fmt.Sprintf("%s = {}\n", fullName))

	// Generate values (enum fields)
	for valName, value := range define.Values {
		// LuaLS often represents enum values as fields on the enum table
		// The type might be inferred or explicitly set if known (e.g., number, string)
		// For simplicity, we'll just add a field annotation with a generic type.
		valType := "any" // Or try to infer from value.Value
		if value.Value != nil {
			// Attempt to infer type from value.Value
			switch value.Value.(type) {
			case int, float64:
				valType = "number"
			case string:
				valType = "string"
			case bool:
				valType = "boolean"
				// Add other types as needed
			}
		}
		sb.WriteString(fmt.Sprintf("---@field %s %s %s\n", valName, valType, value.Description))
	}

	// Recurse into subkeys (nested defines)
	for subName, subDefine := range define.Subkeys {
		g.generateDefine(sb, subName, subDefine, fullName+".")
	}
}

// generateConcept generates LuaLS annotations for Concepts.
func (g *Generator) generateConcept(name string, concept api.Concept) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---@alias %s %s %s\n", name, g.translateFactorioTypeToLuaLS(concept.Type), concept.Description))
	return sb.String()
}

// generateClass generates LuaLS annotations for a Class.
func (g *Generator) generateClass(name string, class api.Class) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---@class %s %s\n", name, class.Description))
	if class.Parent != "" {
		sb.WriteString(fmt.Sprintf("---@field __parent %s\n", class.Parent)) // Indicate parent class
	}
	sb.WriteString(fmt.Sprintf("%s = {}\n", name)) // Classes are typically represented as tables in Lua

	// Generate Properties
	for propName, prop := range class.Properties {
		sb.WriteString(g.generatePropertyAnnotation(propName, prop))
		sb.WriteString("\n")
	}

	// Generate Methods
	for methodName, method := range class.Methods {
		sb.WriteString(g.generateMethodAnnotation(methodName, method))
		sb.WriteString("\n")
	}

	return sb.String()
}

// generatePropertyAnnotation generates the LuaLS annotation for a property.
func (g *Generator) generatePropertyAnnotation(name string, property api.Property) string {
	luaLSType := g.translateFactorioTypeToLuaLS(property.Type)
	optional := ""
	// LuaLS handles optionality often within the type string (e.g., Type | nil)
	// The [opt] tag is more for parameters.
	// If property.Optional { optional = " [opt]" }

	// Combine type and nullability
	if property.Nullable && !strings.Contains(luaLSType, "| nil") {
		luaLSType = luaLSType + " | nil"
	}

	// Indicate read/write status in description or a custom tag if LuaLS supports it
	access := ""
	if property.Read && property.Write {
		access = "(Read/Write)"
	} else if property.Read {
		access = "(Read-only)"
	} else if property.Write {
		access = "(Write-only)"
	}

	desc := property.Description
	if access != "" {
		if desc != "" {
			desc = desc + " " + access
		} else {
			desc = access
		}
	}

	return fmt.Sprintf("---@field %s %s %s", name, luaLSType, desc)
}

// generateMethodAnnotation generates the LuaLS annotation for a method.
func (g *Generator) generateMethodAnnotation(name string, method api.Method) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---@method %s\n", name))

	// Add parameter annotations
	for _, param := range method.Parameters {
		luaLSType := g.translateFactorioTypeToLuaLS(param.Type)
		optional := ""
		if param.Optional {
			optional = " [opt]" // [opt] is common for parameters
		}
		// Handle nullability within the type string for parameters too if needed
		if param.Nullable && !strings.Contains(luaLSType, "| nil") {
			luaLSType = luaLSType + " | nil"
		}

		sb.WriteString(fmt.Sprintf("---@param %s%s %s %s\n", param.Name, optional, luaLSType, param.Description))
	}

	// Add return type annotations
	// Handle multiple return values - LuaLS supports this with multiple @return tags
	for _, ret := range method.ReturnTypes {
		luaLSType := g.translateFactorioTypeToLuaLS(ret.Type)
		if ret.Nullable && !strings.Contains(luaLSType, "| nil") {
			luaLSType = luaLSType + " | nil"
		}
		sb.WriteString(fmt.Sprintf("---@return %s %s\n", luaLSType, ret.Description))
	}

	// Add method description
	sb.WriteString(fmt.Sprintf("%s: %s\n", name, method.Description))

	return sb.String()
}

// translateFactorioTypeToLuaLS translates a Factorio API Type struct to a LuaLS annotation type string.
// This function is crucial and requires careful implementation to handle all Factorio type variations.
func (g *Generator) translateFactorioTypeToLuaLS(t api.Type) string {
	// Handle simple types
	if t.IsSimple() {
		// Map common Factorio types to LuaLS equivalents
		switch t.Name {
		case "string":
			return "string"
		case "int", "uint", "long", "ulong", "float", "double":
			return "number" // Lua has a single number type
		case "boolean":
			return "boolean"
		case "table":
			return "table" // Generic table, might need more specific type if possible
		case "nil":
			return "nil" // Represents the nil value/type
		case "object":
			return "any" // Generic object, use 'any' or a more specific base class if defined
		// Add other simple types as needed
		default:
			// Assume it's a reference to a defined class, concept, or simple type
			return t.Name
		}
	}

	// Handle complex types based on ComplexType field
	switch t.ComplexType {
	case "array":
		if t.Value != nil {
			// Array of a specific type: Type[] or table<integer, Type>
			// LuaLS supports both, Type[] is often cleaner.
			return g.translateFactorioTypeToLuaLS(*t.Value) + "[]"
		}
		return "table" // Generic array if element type is unknown

	case "dictionary":
		if t.Key != nil && t.Value != nil {
			// Dictionary with specific key and value types: table<KeyType, ValueType>
			keyType := g.translateFactorioTypeToLuaLS(*t.Key)
			valueType := g.translateFactorioTypeToLuaLS(*t.Value)
			return fmt.Sprintf("table<%s, %s>", keyType, valueType)
		}
		return "table" // Generic dictionary if types are unknown

	case "union":
		if len(t.Values) > 0 {
			// Union of types: Type1 | Type2 | ...
			var options []string
			for _, optionType := range t.Values {
				options = append(options, g.translateFactorioTypeToLuaLS(optionType))
			}
			return strings.Join(options, " | ")
		}
		return "any" // Union with no options? Shouldn't happen based on docs.

	case "literal":
		// Literal value type. LuaLS might represent this as a union of literal values
		// or a specific type depending on context. For simplicity, return the inferred type.
		if t.LiteralValue != nil {
			switch t.LiteralValue.(type) {
			case int, float64:
				return "number"
			case string:
				return "string"
			case bool:
				return "boolean"
			default:
				return "any" // Unknown literal type
			}
		}
		return "any" // Literal with no value?

	case "type":
		// This seems to be a wrapper around another type, possibly with a description.
		// Just return the translation of the wrapped type.
		if t.Value != nil {
			return g.translateFactorioTypeToLuaLS(*t.Value)
		}
		return "any" // Type wrapper with no inner type?

	case "struct":
		// Represents a structure, often used for concepts that are tables with specific fields.
		// If this 'Type' struct itself has field definitions (not currently modeled in api.Type),
		// you would generate an inline table type definition: { field1: Type1, field2: Type2, ... }
		// As currently modeled, it might represent a named concept that is a struct.
		// If t.Name is present, it's likely a reference to a defined concept/type.
		if t.Name != "" {
			return t.Name
		}
		return "table" // Generic struct/table if no name or fields are defined here

	// Add other complex types as needed based on the JSON analysis
	default:
		// Unknown complex type or a named complex type (like a Concept reference)
		if t.Name != "" {
			return t.Name // Assume it's a reference to a defined type/concept
		}
		return "any" // Fallback for unknown types
	}
}

// generateGlobalObject generates the LuaLS annotation for a global object.
func (g *Generator) generateGlobalObject(name string, global api.GlobalObject) string {
	luaLSType := g.translateFactorioTypeToLuaLS(global.Type)
	return fmt.Sprintf("---@type %s %s\n%s = {}", luaLSType, global.Description, name)
}

// generateEventDataClass generates a class or type alias for event data.
// This is a placeholder and needs proper implementation based on how you want
// to represent event data for LuaLS.
// func (g *Generator) generateEventDataClass(eventName string, event api.Event) string {
// 	var sb strings.Builder
// 	dataTypeName := "EventData." + eventName // e.g., EventData.on_built_entity
// 	sb.WriteString(fmt.Sprintf("---@class %s : EventData %s\n", dataTypeName, event.Description)) // Inherit from base EventData
// 	sb.WriteString(fmt.Sprintf("%s = {}\n", dataTypeName))
//
// 	// Add fields for event data parameters
// 	for _, param := range event.Data {
// 		luaLSType := g.translateFactorioTypeToLuaLS(param.Type)
//         if param.Nullable && !strings.Contains(luaLSType, "| nil") {
//              luaLSType = luaLSType + " | nil"
//         }
// 		sb.WriteString(fmt.Sprintf("---@field %s %s %s\n", param.Name, luaLSType, param.Description))
// 	}
// 	return sb.String()
// }

// generatePrototypeClass generates a class or type alias for a prototype type.
// This is a placeholder and needs proper implementation based on how you want
// to represent prototype definitions for LuaLS.
// func (g *Generator) generatePrototypeClass(name string, prototype api.Prototype) string {
// 	var sb strings.Builder
// 	// Prototypes might be represented as classes or just documented structures
// 	sb.WriteString(fmt.Sprintf("---@class data.%s %s\n", name, prototype.Description)) // Prefix with data.
// 	if prototype.Parent != "" {
// 		sb.WriteString(fmt.Sprintf("---@field __parent data.%s\n", prototype.Parent))
// 	}
// 	sb.WriteString(fmt.Sprintf("data.%s = {}\n", name))
//
// 	// Add properties for the prototype
// 	for propName, prop := range prototype.Properties {
// 		// Prototype properties might use different types or structures than runtime properties
// 		sb.WriteString(g.generatePropertyAnnotation(propName, prop)) // Reuse property annotation logic
// 		sb.WriteString("\n")
// 	}
// 	return sb.String()
// }
