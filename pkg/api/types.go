package api

import (
	"bytes" // Import the bytes package
	"encoding/json"
	"fmt"
	"log" // Import the log package
)

// API represents the overall structure of the Factorio API JSON files.
// This struct is designed to handle both runtime-api.json and prototype-api.json,
// including common top-level keys.
// Note: Top-level collections are arrays in the JSON, hence the use of slices here.
type API struct {
	Classes       []Class        `json:"classes,omitempty"`
	Events        []Event        `json:"events,omitempty"`
	Defines       []Define       `json:"defines,omitempty"`
	GlobalObjects []GlobalObject `json:"global_objects,omitempty"`
	Concepts      []Concept      `json:"concepts,omitempty"`      // Found in both APIs, often custom types
	Prototypes    []Prototype    `json:"prototypes,omitempty"`    // Specific to prototype-api.json
	BuiltinTypes  []Type         `json:"builtin_types,omitempty"` // Documented built-in types
	// Add other top-level fields if needed after a full analysis
}

// BasicMember represents common fields found in many API objects.
type BasicMember struct {
	Name        string   `json:"name"`
	Order       int      `json:"order"` // Order as shown on the website
	Description string   `json:"description"`
	Lists       []string `json:"lists,omitempty"`    // Additional markdown lists
	Examples    []string `json:"examples,omitempty"` // Code examples
	// Images []Image `json:"images,omitempty"` // If you need to parse image info
	// Note: 'Notes' field also exists on some members
}

// Class represents a Factorio Lua API class.
// Methods and Properties are arrays in the JSON, not maps keyed by name.
type Class struct {
	BasicMember
	Methods    []Method   `json:"methods,omitempty"`    // Corrected to slice
	Properties []Property `json:"properties,omitempty"` // Corrected to slice
	Parent     string     `json:"parent,omitempty"`     // Inherited class name
	Abstract   bool       `json:"abstract,omitempty"`
	// Add other class-specific fields
}

// Event represents a Factorio Lua API event.
type Event struct {
	BasicMember
	Data []Parameter `json:"data,omitempty"` // Parameters passed to the event handler
	// Add other event-specific fields
}

// Define represents a Factorio Lua API define (enum-like structure).
// Values and Subkeys are arrays in the JSON, not maps keyed by name.
type Define struct {
	BasicMember
	Values  []DefineValue `json:"values,omitempty"`  // Corrected to slice
	Subkeys []Define      `json:"subkeys,omitempty"` // Corrected to slice
	// Add other define-specific fields
}

// DefineValue represents a value within a Define.
type DefineValue struct {
	BasicMember
	Value interface{} `json:"value,omitempty"` // The actual value (number, string, etc.)
}

// GlobalObject represents a global variable available in the Lua environment.
type GlobalObject struct {
	BasicMember
	Type Type `json:"type"` // The type of the global object
}

// Concept represents a custom type or concept used in the API.
type Concept struct {
	BasicMember
	Category string `json:"category"` // e.g., "type", "concept"
	Type     Type   `json:"type"`     // The underlying type definition
	// Add other concept-specific fields
}

// Prototype represents a Factorio Prototype definition.
type Prototype struct {
	BasicMember
	TypeName   string     `json:"typename,omitempty"` // The specific type name (e.g., "item", "recipe")
	Parent     string     `json:"parent,omitempty"`   // Parent prototype name
	Abstract   bool       `json:"abstract,omitempty"`
	Properties []Property `json:"properties,omitempty"` // Corrected to slice
	// Add other prototype-specific fields
}

// Method represents a method of a class.
type Method struct {
	BasicMember
	Parameters  []Parameter  `json:"parameters,omitempty"`
	ReturnTypes []ReturnType `json:"return_types,omitempty"` // Can return multiple values
	Variadic    bool         `json:"variadic,omitempty"`     // If it accepts variable arguments
	// Add other method-specific fields
}

// Property represents a property of a class or prototype.
type Property struct {
	BasicMember
	Type     Type        `json:"type"`
	Optional bool        `json:"optional,omitempty"`
	Nullable bool        `json:"nullable,omitempty"`
	Read     bool        `json:"read,omitempty"`     // Is readable
	Write    bool        `json:"write,omitempty"`    // Is writable
	Overload bool        `json:"overload,omitempty"` // If it overrides a parent property
	AltName  string      `json:"alt_name,omitempty"` // Alternative name
	Default  interface{} `json:"default,omitempty"`  // Default value
	// Add other property-specific fields
}

// Parameter represents a parameter of a method or event.
type Parameter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        Type   `json:"type"`
	Optional    bool   `json:"optional,omitempty"`
	Nullable    bool   `json:"nullable,omitempty"`
	Order       int    `json:"order"` // Order of the parameter
	// Add other parameter-specific fields
}

// ReturnType represents a return value of a method.
type ReturnType struct {
	Type        Type   `json:"type"`
	Description string `json:"description"`
	Optional    bool   `json:"optional,omitempty"`
	Nullable    bool   `json:"nullable,omitempty"`
	Order       int    `json:"order"` // Order of the return value
}

// Type represents a data type in the Factorio API. This struct and its
// UnmarshalJSON method are designed to handle the various ways types
// are defined in the JSON (simple name, complex structure, unions, etc.).
// It includes an anonymous BasicMember field to capture common fields
// like Description when present in complex types (e.g., literal, union, type, tuple).
type Type struct {
	Name string `json:"name,omitempty"` // For simple types (string, int, bool, etc.) or named complex types

	// Fields for complex types. Only one of these should typically be present
	// depending on the 'complex_type' or implicit structure.
	ComplexType string `json:"complex_type,omitempty"` // e.g., "array", "dictionary", "union", "literal", "type", "struct", "tuple"

	// Details for specific complex types:
	Value *Type `json:"value,omitempty"` // For "array" (element type) or "type" (actual type)
	Key   *Type `json:"key,omitempty"`   // For "dictionary" (key type)
	// Value field is also used for "dictionary" (value type)

	Values []Type `json:"values,omitempty"` // For "tuple" (element elements) or "union" (possible types)

	LiteralValue interface{} `json:"value,omitempty"` // For "literal" (the literal value)

	FullFormat bool `json:"full_format,omitempty"` // For "union" (if options have descriptions)

	// Include BasicMember anonymously to get Description and other common fields
	// when they are present in complex type definitions (e.g., for literals, unions).
	BasicMember
}

// UnmarshalJSON is a custom unmarshaler for the Type struct to handle
// the varied structure of type definitions in the Factorio API JSON.
// It first attempts to unmarshal into a temporary struct to capture
// the complex_type and name, then uses json.RawMessage to handle
// nested structures based on the complex_type.
func (t *Type) UnmarshalJSON(data []byte) error {
	log.Printf("UnmarshalJSON for Type: Raw data size %d, Data: %s", len(data), string(data))

	// First, check if the data is a simple string.
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		// If it's a string, set the Name field and return.
		t.Name = stringValue
		t.ComplexType = "" // Ensure complex type is empty for simple types
		log.Printf("UnmarshalJSON: Unmarshaled as simple string: '%s'", t.Name)
		return nil
	}

	// If it's not a string, assume it's a JSON object and proceed with complex unmarshalling.
	// Define a temporary struct to capture the complex_type and name first,
	// and use json.RawMessage for fields that depend on complex_type.
	// This avoids infinite recursion and allows conditional unmarshalling.
	temp := struct {
		Name        string `json:"name,omitempty"`
		ComplexType string `json:"complex_type,omitempty"`
		FullFormat  bool   `json:"full_format,omitempty"` // Included here as it's a direct field

		// Use raw messages for fields whose structure depends on ComplexType
		ValueRaw  json.RawMessage `json:"value,omitempty"`
		KeyRaw    json.RawMessage `json:"key,omitempty"`
		ValuesRaw json.RawMessage `json:"values,omitempty"`

		// BasicMember fields might be present for some complex types (union, literal, type, tuple)
		// Unmarshal these into a separate struct first.
		BasicMemberRaw json.RawMessage `json:",inline"` // Use inline to capture top-level BasicMember fields
	}{}

	log.Println("UnmarshalJSON: Data is not a string, attempting complex unmarshalling.")
	if err := json.Unmarshal(data, &temp); err != nil {
		log.Printf("Error initial complex unmarshal of Type struct: %v", err)
		return fmt.Errorf("failed initial complex unmarshal of Type struct: %w", err)
	}

	// Assign the basic fields from the temporary struct
	t.Name = temp.Name
	t.ComplexType = temp.ComplexType
	t.FullFormat = temp.FullFormat

	log.Printf("UnmarshalJSON (Complex): Name='%s', ComplexType='%s'", t.Name, t.ComplexType)

	// Unmarshal BasicMember fields if they were present
	if len(temp.BasicMemberRaw) > 0 {
		// Need to unmarshal into a BasicMember struct to populate it
		var bm BasicMember
		// Re-unmarshal the raw data into the BasicMember struct.
		// This is safe because BasicMemberRaw contains the original JSON data
		// and BasicMember only has simple fields or fields handled by default unmarshalling.
		// Check if the raw data is not null or an empty object before attempting to unmarshal BasicMember
		if !bytes.Equal(temp.BasicMemberRaw, []byte("null")) && !bytes.Equal(temp.BasicMemberRaw, []byte("{}")) {
			if err := json.Unmarshal(temp.BasicMemberRaw, &bm); err != nil {
				log.Printf("Warning: Failed to unmarshal BasicMember within Type: %v", err)
				// Continue without BasicMember data if it fails
			} else {
				t.BasicMember = bm
				log.Printf("UnmarshalJSON (Complex): Unmarshaled BasicMember - Name='%s', Description='%s'", bm.Name, bm.Description)
			}
		}
	}

	// Now, based on ComplexType, unmarshal the raw fields into the correct Type fields
	switch t.ComplexType {
	case "array":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'array'")
		if len(temp.ValueRaw) > 0 {
			t.Value = &Type{} // Initialize nested Type
			if err := json.Unmarshal(temp.ValueRaw, t.Value); err != nil {
				log.Printf("Error unmarshalling array value type: %v", err)
				return fmt.Errorf("failed to unmarshal array value type: %w", err)
			}
			log.Printf("UnmarshalJSON (Complex): Unmarshaled array value type")
		}
	case "dictionary":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'dictionary'")
		if len(temp.KeyRaw) > 0 {
			t.Key = &Type{} // Initialize nested Type
			if err := json.Unmarshal(temp.KeyRaw, t.Key); err != nil {
				log.Printf("Error unmarshalling dictionary key type: %v", err)
				return fmt.Errorf("failed to unmarshal dictionary key type: %w", err)
			}
			log.Printf("UnmarshalJSON (Complex): Unmarshaled dictionary key type")
		}
		if len(temp.ValueRaw) > 0 { // Note: Dictionary value also uses the "value" key
			t.Value = &Type{} // Initialize nested Type
			if err := json.Unmarshal(temp.ValueRaw, t.Value); err != nil {
				log.Printf("Error unmarshalling dictionary value type: %v", err)
				return fmt.Errorf("failed to unmarshal dictionary value type: %w", err)
			}
			log.Printf("UnmarshalJSON (Complex): Unmarshaled dictionary value type")
		}
	case "union":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'union'")
		if len(temp.ValuesRaw) > 0 {
			if err := json.Unmarshal(temp.ValuesRaw, &t.Values); err != nil {
				log.Printf("Error unmarshalling union values: %v", err)
				return fmt.Errorf("failed to unmarshal union values: %w", err)
			}
			log.Printf("UnmarshalJSON (Complex): Unmarshaled %d union values", len(t.Values))
		}
		// BasicMember fields (like Description) are handled by the BasicMemberRaw unmarshalling
		// FullFormat is handled by the initial unmarshalling
	case "literal":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'literal'")
		// Literal value can be string, number, or boolean. Unmarshal RawMessage directly.
		// The key for the literal value is also "value".
		if len(temp.ValueRaw) > 0 {
			// Try unmarshalling into an interface{} to keep the original type
			var val interface{}
			if err := json.Unmarshal(temp.ValueRaw, &val); err != nil {
				log.Printf("Error unmarshalling literal value: %v", err)
				return fmt.Errorf("failed to unmarshal literal value: %w", err)
			}
			t.LiteralValue = val
			log.Printf("UnmarshalJSON (Complex): Unmarshaled literal value: %v (Type: %T)", val, val)
		}
		// BasicMember fields (like Description) are handled by the BasicMemberRaw unmarshalling
	case "type":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'type'")
		// This complex type wraps another type, using the "value" key
		if len(temp.ValueRaw) > 0 {
			t.Value = &Type{} // Initialize nested Type
			if err := json.Unmarshal(temp.ValueRaw, t.Value); err != nil {
				log.Printf("Error unmarshalling wrapped type value: %v", err)
				return fmt.Errorf("failed to unmarshal wrapped type value: %w", err)
			}
			log.Printf("UnmarshalJSON (Complex): Unmarshaled wrapped type value")
		}
		// BasicMember fields (like Description) are handled by the BasicMemberRaw unmarshalling
	case "struct":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'struct'")
		// 'struct' often just has a name and description, or might imply fields
		// defined elsewhere. The BasicMember fields handle name/description.
		// If there were inline field definitions, they would need to be handled here.
		// Based on the Factorio JSON docs, 'struct' often appears as a complex_type
		// for named concepts or types that are essentially tables/structs.
		// No additional unmarshalling is needed for the basic 'struct' case as defined.
		// BasicMember fields (like Description) are handled by the BasicMemberRaw unmarshalling
	case "tuple":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'tuple'")
		if len(temp.ValuesRaw) > 0 {
			if err := json.Unmarshal(temp.ValuesRaw, &t.Values); err != nil {
				log.Printf("Error unmarshalling tuple values: %v", err)
				return fmt.Errorf("failed to unmarshal tuple values: %w", err)
			}
			log.Printf("UnmarshalJSON (Complex): Unmarshaled %d tuple values", len(t.Values))
		}
		// BasicMember fields (like Description) are handled by the BasicMemberRaw unmarshalling

	case "builtin":
		log.Println("UnmarshalJSON (Complex): Handling complex_type 'builtin'")
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
		if t.Name == "" {
			// This case might indicate an issue with the JSON or an unhandled type structure.
			// Log a warning or return an error if strict parsing is needed.
			log.Printf("Warning: UnmarshalJSON (Complex): Encountered type with no Name and no ComplexType: %s", string(data))
		} else {
			log.Printf("UnmarshalJSON (Complex): Handling simple type or unknown complex type with Name='%s'", t.Name)
		}
	}

	log.Println("UnmarshalJSON: Finished Type unmarshalling")
	return nil
}

// Helper to check if a type is a complex type
func (t Type) IsComplex() bool {
	return t.ComplexType != ""
}

// Helper to check if a type is a simple named type
func (t Type) IsSimple() bool {
	return t.Name != "" && t.ComplexType == ""
}

// Helper to check if a type is an array
func (t Type) IsArray() bool {
	return t.ComplexType == "array" && t.Value != nil
}

// Helper to check if a type is a dictionary
func (t Type) IsDictionary() bool {
	return t.ComplexType == "dictionary" && t.Key != nil && t.Value != nil
}

// Helper to check if a type is a union
func (t Type) IsUnion() bool {
	return t.ComplexType == "union" && len(t.Values) > 0
}

// Helper to check if a type is a literal
func (t Type) IsLiteral() bool {
	return t.ComplexType == "literal"
}

// Helper to check if a type is a reference to another named type (like a Concept)
// This is a heuristic: if it has a Name and a ComplexType that isn't just a basic
// structural type like "array" or "dictionary", it might be a named complex type.
// Or, if it has a Name and ComplexType is "struct", it's likely a named struct concept.
func (t Type) IsNamedComplex() bool {
	return t.Name != "" && t.ComplexType != "" && t.ComplexType != "array" && t.ComplexType != "dictionary" && t.ComplexType != "union" && t.ComplexType != "literal" && t.ComplexType != "type" && t.ComplexType != "tuple" && t.ComplexType != "struct" && t.ComplexType != "builtin" // Added struct and builtin here
}

// Helper to check if a type is a tuple
func (t Type) IsTuple() bool {
	return t.ComplexType == "tuple" && len(t.Values) > 0
}

// Helper to check if a type is a builtin type marker
func (t Type) IsBuiltinMarker() bool {
	return t.ComplexType == "builtin" && t.Name == "" && t.Value == nil && t.Key == nil && len(t.Values) == 0 && t.LiteralValue == nil
}
