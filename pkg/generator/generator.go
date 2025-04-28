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
	// Iterate over the slice and pass the Define struct directly
	for _, define := range runtimeAPI.Defines {
		g.generateDefine(&runtimeSB, define, "") // Pass the struct, start recursion with empty prefix
		runtimeSB.WriteString("\n")
	}

	// Generate Concepts (Runtime)
	runtimeSB.WriteString("-- Concepts (Runtime)\n\n")
	// Iterate over the slice and pass the Concept struct directly
	for _, concept := range runtimeAPI.Concepts {
		// Concepts can be aliases or complex types, need to handle based on Category and Type structure
		runtimeSB.WriteString(g.generateConcept(concept)) // Pass the struct
		runtimeSB.WriteString("\n")
	}

	// Generate Classes
	runtimeSB.WriteString("-- Classes\n\n")
	// Iterate over the slice and pass the Class struct directly
	for _, class := range runtimeAPI.Classes {
		runtimeSB.WriteString(g.generateClass(class)) // Pass the struct
		runtimeSB.WriteString("\n")
	}

	// Generate Global Objects
	runtimeSB.WriteString("-- Global Objects\n\n")
	// Iterate over the slice and pass the GlobalObject struct directly
	for _, global := range runtimeAPI.GlobalObjects {
		runtimeSB.WriteString(g.generateGlobalObject(global)) // Pass the struct
		runtimeSB.WriteString("\n")
	}

	// Generate Events
	// Events are typically handled by defining types for event data payloads
	// and potentially documenting the script.on_event function.
	runtimeSB.WriteString("-- Events\n\n")
	runtimeSB.WriteString("---@class EventData\n") // Base class for all event data
	runtimeSB.WriteString("EventData = {}\n\n")

	// Iterate over the slice and pass the Event struct directly
	for _, event := range runtimeAPI.Events {
		runtimeSB.WriteString(g.generateEventDataClass(event)) // Pass the struct
		runtimeSB.WriteString("\n")
	}

	// You might also want to document script.on_event with overloads
	// for better type checking when registering handlers. This is more complex
	// and depends on LuaLS capabilities for function overloads with specific
	// literal string arguments. For now, we focus on the data types.

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
		// Iterate over the slice and pass the Define struct directly
		for _, define := range prototypeAPI.Defines {
			g.generateDefine(&prototypeSB, define, "") // Pass the struct
			prototypeSB.WriteString("\n")
		}
	}

	// Generate Concepts (Prototype)
	prototypeSB.WriteString("-- Concepts (Prototype)\n\n")
	// Assuming prototypeAPI has a Concepts field
	if prototypeAPI.Concepts != nil {
		// Iterate over the slice and pass the Concept struct directly
		for _, concept := range prototypeAPI.Concepts {
			prototypeSB.WriteString(g.generateConcept(concept)) // Pass the struct
			prototypeSB.WriteString("\n")
		}
	}

	// Generate Prototypes
	// Prototypes themselves are definitions, not runtime objects.
	// You might define types representing each prototype type (e.g., "item", "recipe").
	prototypeSB.WriteString("-- Prototypes\n\n")
	// Assuming prototypeAPI has a Prototypes field
	if prototypeAPI.Prototypes != nil {
		// First, define a base class for all prototypes
		prototypeSB.WriteString("---@class Prototype\n")
		prototypeSB.WriteString("Prototype = {}\n\n")

		// Then, define a class for each specific prototype type (e.g., ItemPrototype, RecipePrototype)
		// and a class for each individual prototype instance (e.g., data.raw.item.iron_plate)
		// This requires iterating through prototypes and grouping them by typename.
		prototypesByTypeName := make(map[string]map[string]api.Prototype)
		for _, prototype := range prototypeAPI.Prototypes { // Iterate over the slice
			if prototypesByTypeName[prototype.TypeName] == nil {
				prototypesByTypeName[prototype.TypeName] = make(map[string]api.Prototype)
			}
			prototypesByTypeName[prototype.TypeName][prototype.Name] = prototype // Use prototype.Name as key
		}

		for typeName, prototypes := range prototypesByTypeName {
			// Define a class for the type name (e.g., ItemPrototype)
			typeClassName := strings.Title(typeName) + "Prototype" // Capitalize first letter
			// Pass the map of prototypes for this type, not an individual prototype
			prototypeSB.WriteString(g.generatePrototypeTypeClass(typeClassName, typeName, prototypes))
			prototypeSB.WriteString("\n")

			// Define a global table for data.raw.<typename>
			prototypeSB.WriteString(fmt.Sprintf("---@type table<string, %s> Table of %s prototypes by name.\n", typeClassName, typeName))
			prototypeSB.WriteString(fmt.Sprintf("data.raw.%s = {}\n\n", typeName))

			// Optionally, define individual fields on data.raw.<typename> for specific prototypes
			// This can make the definition file very large, but provides direct autocompletion
			// for known prototype names (e.g., data.raw.item.iron_plate).
			// for protoName, prototype := range prototypes { // Iterate over the map
			// 	prototypeSB.WriteString(fmt.Sprintf("---@field %s %s %s\n", protoName, typeClassName, prototype.Description))
			// }
			// prototypeSB.WriteString(fmt.Sprintf("data.raw.%s = {}\n\n", typeName)) // Redefine the table with fields
		}
	}

	definitions["prototype.lua"] = prototypeSB.String()

	return definitions, nil
}

// generateDefine recursively generates LuaLS annotations for Defines.
// Now accepts the Define struct directly.
func (g *Generator) generateDefine(sb *strings.Builder, define api.Define, prefix string) {
	fullName := prefix + define.Name // Use the Name field from the struct
	sb.WriteString(fmt.Sprintf("---@class %s %s\n", fullName, define.Description))
	sb.WriteString(fmt.Sprintf("%s = {}\n", fullName))

	// Generate values (enum fields)
	// Iterate over the slice
	for _, value := range define.Values {
		// LuaLS often represents enum values as fields on the enum table
		// The type might be inferred or explicitly set if known (e.g., number, string)
		valType := "any" // Default type
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
		sb.WriteString(fmt.Sprintf("---@field %s %s %s\n", value.Name, valType, value.Description)) // Use value.Name
	}

	// Recurse into subkeys (nested defines)
	// Iterate over the slice
	for _, subDefine := range define.Subkeys {
		g.generateDefine(sb, subDefine, fullName+".") // Pass the subDefine struct
	}
}

// generateConcept generates LuaLS annotations for Concepts.
// Now accepts the Concept struct directly.
func (g *Generator) generateConcept(concept api.Concept) string {
	var sb strings.Builder
	// Concepts are often aliases or specific table structures.
	// If the concept has a complex type defined directly, generate an alias.
	// If it's just a named concept with a category like "type", it might be
	// a reference handled by translateFactorioTypeToLuaLS.
	if concept.Type.IsComplex() || concept.Type.IsSimple() { // Check if the nested Type has definition details
		sb.WriteString(fmt.Sprintf("---@alias %s %s %s\n", concept.Name, g.translateFactorioTypeToLuaLS(concept.Type), concept.Description)) // Use concept.Name
	} else {
		// If the nested type is just a name without complex details here,
		// it's likely already handled as a direct type reference.
		// We could add a comment or skip, depending on desired output verbosity.
		// For now, we'll generate an alias if the type has a name, assuming it
		// refers to a defined type elsewhere.
		if concept.Type.Name != "" {
			sb.WriteString(fmt.Sprintf("---@alias %s %s %s\n", concept.Name, concept.Type.Name, concept.Description)) // Use concept.Name
		} else {
			// If the concept has no type name or complex type, it's hard to define.
			// Add a comment indicating this.
			sb.WriteString(fmt.Sprintf("-- Undefined concept: %s %s\n", concept.Name, concept.Description)) // Use concept.Name
		}
	}

	return sb.String()
}

// generateClass generates LuaLS annotations for a Class.
// Now accepts the Class struct directly.
func (g *Generator) generateClass(class api.Class) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---@class %s %s\n", class.Name, class.Description)) // Use class.Name
	if class.Parent != "" {
		sb.WriteString(fmt.Sprintf("---@field __parent %s\n", class.Parent)) // Indicate parent class
	}
	sb.WriteString(fmt.Sprintf("%s = {}\n", class.Name)) // Use class.Name // Classes are typically represented as tables in Lua

	// Generate Properties
	// Iterate over the slice
	for _, prop := range class.Properties {
		sb.WriteString(g.generatePropertyAnnotation(prop.Name, prop)) // Use prop.Name
		sb.WriteString("\n")
	}

	// Generate Methods
	// Iterate over the slice
	for _, method := range class.Methods {
		sb.WriteString(g.generateMethodAnnotation(method.Name, method)) // Use method.Name
		sb.WriteString("\n")
	}

	return sb.String()
}

// generatePropertyAnnotation generates the LuaLS annotation for a property.
func (g *Generator) generatePropertyAnnotation(name string, property api.Property) string {
	luaLSType := g.translateFactorioTypeToLuaLS(property.Type)
	// LuaLS handles optionality often within the type string (e.g., Type | nil)
	// The [opt] tag is more for parameters.

	// Combine type and nullability
	if property.Nullable && !strings.Contains(luaLSType, "| nil") {
		luaLSType = luaLSType + " | nil"
	}
	// Also consider optional properties as potentially nil if not explicitly nullable
	// This might be a matter of preference or how LuaLS interprets optional fields.
	// If property.Optional && !property.Nullable && !strings.Contains(luaLSType, "| nil") {
	//      luaLSType = luaLSType + " | nil"
	// }

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
		case "int", "uint", "long", "ulong", "float", "double", "number": // Added "number" explicitly
			return "number" // Lua has a single number type
		case "boolean":
			return "boolean"
		case "table":
			return "table" // Generic table, might need more specific type if possible
		case "nil":
			return "nil" // Represents the nil value/type
		case "object":
			return "any" // Generic object, use 'any' or a more specific base class if defined
		case "LuaObject":
			return "LuaObject" // Base class for Lua objects
		case "void":
			return "nil" // Methods returning nothing might be 'void' in JSON, map to nil
		// Add other common simple types or built-in types if needed
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
			switch val := t.LiteralValue.(type) {
			case int, float64:
				return fmt.Sprintf("%v", val) // Represent literal numbers directly
			case string:
				// Escape string literals for Lua
				escapedString := strings.ReplaceAll(val, `"`, `\"`)
				escapedString = strings.ReplaceAll(escapedString, `\`, `\\`)   // Escape backslashes
				escapedString = strings.ReplaceAll(escapedString, "\n", "\\n") // Escape newlines
				escapedString = strings.ReplaceAll(escapedString, "\r", "\\r") // Escape carriage returns
				escapedString = strings.ReplaceAll(escapedString, "\t", "\\t") // Escape tabs
				return fmt.Sprintf(`"%s"`, escapedString)                      // Represent literal strings directly
			case bool:
				return fmt.Sprintf("%v", val) // Represent literal booleans directly (true or false)
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
		// 'struct' often just has a name and description, or might imply fields
		// defined elsewhere. The BasicMember fields handle name/description.
		// If there were inline field definitions, they would need to be handled here.
		// Based on the Factorio JSON docs, 'struct' often appears as a complex_type
		// for named concepts or types that are essentially tables/structs.
		// If t.Name is present, it's likely a reference to a defined concept/type.
		if t.Name != "" {
			return t.Name
		}
		return "table" // Generic struct/table if no name or fields are defined here

	case "tuple":
		if len(t.Values) > 0 {
			// Tuple of types: {Type1, Type2, ...} or LuaLS specific tuple syntax if available/preferred
			// LuaLS often represents tuples as tables with specific field types or ordered elements.
			// A common representation is a table type with numeric keys: table<integer, Type1|Type2|...>
			// Or if the order/specific types are crucial, an inline table type: {1: Type1, 2: Type2, ...}
			// Let's use the inline table type for stricter tuple representation.
			var fields []string
			for i, elementType := range t.Values {
				fields = append(fields, fmt.Sprintf("%d: %s", i+1, g.translateFactorioTypeToLuaLS(elementType)))
			}
			return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
		}
		return "table" // Generic table if tuple elements are unknown

	case "builtin":
		// The log shows {"complex_type":"builtin"} which implies no name or value here.
		// The name for builtin types might be the key in the BuiltinTypes map at the top level.
		// No further unmarshalling is needed for this structure based on the log.
		// The Name field would be populated if this Type struct was part of a map/slice
		// where the name is the key/part of the surrounding structure.
		// If this "builtin" type appears in a context where it needs a name (e.g., a parameter type),
		// the name should come from the surrounding structure, not the Type object itself.
		// We'll return "any" or a specific Lua primitive if the context implies it.
		// Based on the log, it seems these "builtin" markers appear where a type is expected.
		// Returning "any" is a safe fallback, or we could try to infer from context if possible.
		// For now, returning "any" for the marker itself. The actual builtin types (like "boolean")
		// are handled by the IsSimple() case.
		return "any"

	default:
		// If ComplexType is empty or unknown, it might be a simple type with just a Name.
		// The initial unmarshalling into temp.Name handles the Name field.
		// If Name is also empty and ComplexType is empty, it might be an error in the JSON
		// or a type we haven't accounted for.
		if t.Name != "" {
			return t.Name // Assume it's a reference to a defined type/concept
		}
		return "any" // Fallback for unknown types or parsing issues
	}
}

// generateGlobalObject generates the LuaLS annotation for a global object.
// Now accepts the GlobalObject struct directly.
func (g *Generator) generateGlobalObject(global api.GlobalObject) string {
	luaLSType := g.translateFactorioTypeToLuaLS(global.Type)
	// Global objects are typically defined as global variables with type annotations.
	return fmt.Sprintf("---@type %s %s\n%s = {}", luaLSType, global.Description, global.Name) // Use global.Name
}

// generateEventDataClass generates a class for event data payload.
// Now accepts the Event struct directly.
func (g *Generator) generateEventDataClass(event api.Event) string {
	var sb strings.Builder
	// Event data classes are typically named EventData.<event_name> and inherit from a base EventData class.
	dataTypeName := "EventData." + event.Name                                                     // Use event.Name
	sb.WriteString(fmt.Sprintf("---@class %s : EventData %s\n", dataTypeName, event.Description)) // Inherit from base EventData
	sb.WriteString(fmt.Sprintf("%s = {}\n\n", dataTypeName))                                      // Define the class table

	// Add fields for event data parameters
	for _, param := range event.Data {
		luaLSType := g.translateFactorioTypeToLuaLS(param.Type)
		// Handle nullability within the type string for parameters
		if param.Nullable && !strings.Contains(luaLSType, "| nil") {
			luaLSType = luaLSType + " | nil"
		}
		// Optional parameters in event data are still fields, but their value might be nil
		// The [opt] tag is more for function parameters.
		// If param.Optional && !param.Nullable && !strings.Contains(luaLSType, "| nil") {
		//      luaLSType = luaLSType + " | nil"
		// }

		sb.WriteString(fmt.Sprintf("---@field %s %s %s\n", param.Name, luaLSType, param.Description))
	}
	return sb.String()
}

// generatePrototypeTypeClass generates a class for a specific prototype type (e.g., ItemPrototype).
// Now accepts the map of prototypes for this type.
func (g *Generator) generatePrototypeTypeClass(className string, typeName string, prototypes map[string]api.Prototype) string {
	var sb strings.Builder
	// Define a class for the prototype type, inheriting from the base Prototype class.
	sb.WriteString(fmt.Sprintf("---@class %s : Prototype Represents a %s prototype definition.\n", className, typeName))
	sb.WriteString(fmt.Sprintf("%s = {}\n\n", className)) // Define the class table

	// Collect all unique properties across all prototypes of this type.
	// This is a simplification; ideally, properties might vary per specific prototype.
	// A more complex approach would be to define unions or intersections of types.
	// For now, we'll define fields for properties found in at least one prototype of this type.
	allProperties := make(map[string]api.Property)
	for _, prototype := range prototypes { // Iterate over the map values
		for _, prop := range prototype.Properties {
			// Simple merge: if property exists, use the one encountered last.
			// A more robust approach would merge types for properties with the same name.
			allProperties[prop.Name] = prop
		}
	}

	// Generate fields for the collected properties.
	for propName, prop := range allProperties {
		luaLSType := g.translateFactorioTypeToLuaLS(prop.Type)
		// Prototype properties are part of the definition data, not runtime objects.
		// Optional/nullable might be handled differently than runtime properties.
		// For now, apply nullable annotation if specified.
		if prop.Nullable && !strings.Contains(luaLSType, "| nil") {
			luaLSType = luaLSType + " | nil"
		}
		// Optional properties might also be nil in the data.raw table.
		if prop.Optional && !prop.Nullable && !strings.Contains(luaLSType, "| nil") {
			luaLSType = luaLSType + " | nil"
		}

		// Indicate read/write status (less relevant for static prototype data, but include description)
		access := ""
		if prop.Read && prop.Write {
			access = "(Read/Write)"
		} else if prop.Read {
			access = "(Read-only)"
		} else if prop.Write {
			access = "(Write-only)"
		}

		desc := prop.Description
		if access != "" {
			if desc != "" {
				desc = desc + " " + access
			} else {
				desc = access
			}
		}

		sb.WriteString(fmt.Sprintf("---@field %s %s %s\n", propName, luaLSType, desc))
	}

	return sb.String()
}
