package api

// API represents the overall structure of the Factorio API JSON files.
// Note: The structure can vary slightly between runtime and prototype APIs,
// this struct is a simplified representation. You might need to adjust
// based on a full analysis of both JSON files.
type API struct {
	Classes      map[string]Class      `json:"classes,omitempty"`
	Events       map[string]Event      `json:"events,omitempty"`
	Defines      map[string]Define     `json:"defines,omitempty"`
	GlobalObjects map[string]GlobalObject `json:"global_objects,omitempty"`
	Concepts     map[string]Concept    `json:"concepts,omitempty"` // Found in both APIs, often custom types
	// Add other top-level fields as needed after reviewing the JSON
}

// BasicMember represents common fields found in many API objects.
type BasicMember struct {
	Name        string   `json:"name"`
	Order       int      `json:"order"` // Order as shown on the website
	Description string   `json:"description"`
	Lists       []string `json:"lists,omitempty"`   // Additional markdown lists
	Examples    []string `json:"examples,omitempty"` // Code examples
	// Images []Image `json:"images,omitempty"` // If you need to parse image info
}

// Class represents a Factorio Lua API class.
type Class struct {
	BasicMember
	Methods    map[string]Method   `json:"methods,omitempty"`
	Properties map[string]Property `json:"properties,omitempty"`
	Parent     string              `json:"parent,omitempty"` // Inherited class
	Abstract   bool                `json:"abstract,omitempty"`
	// Add other class-specific fields
}

// Event represents a Factorio Lua API event.
type Event struct {
	BasicMember
	Data []Parameter `json:"data,omitempty"` // Parameters passed to the event handler
	// Add other event-specific fields
}

// Define represents a Factorio Lua API define (enum-like structure).
type Define struct {
	BasicMember
	Values  map[string]DefineValue `json:"values,omitempty"`  // Individual enum values
	Subkeys map[string]Define      `json:"subkeys,omitempty"` // Nested defines
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
	Type      Type `json:"type"`
	Optional  bool `json:"optional,omitempty"`
	Nullable  bool `json:"nullable,omitempty"`
	Read      bool `json:"read,omitempty"`  // Is readable
	Write     bool `json:"write,omitempty"` // Is writable
	Overload  bool `json:"overload,omitempty"` // If it overrides a parent property
	AltName   string `json:"alt_name,omitempty"` // Alternative name
	Default   interface{} `json:"default,omitempty"` // Default value
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

// Type represents a data type in the Factorio API. This is a complex structure
// that can represent simple types, arrays, dictionaries, unions, literals, etc.
// This struct needs to be carefully designed to handle all variations.
type Type struct {
	Name string `json:"name,omitempty"` // For simple types (string, int, bool, etc.) or named complex types

	// Fields for complex types. Only one of these should typically be present
	// depending on the 'complex_type' or implicit structure.
	ComplexType string `json:"complex_type,omitempty"` // e.g., "array", "dictionary", "union", "literal", "type", "struct"

	// Details for specific complex types:
	Value *Type `json:"value,omitempty"` // For "array" (element type) or "type" (actual type)
	Key   *Type `json:"key,omitempty"`   // For "dictionary" (key type)
	// Value field is also used for "dictionary" (value type)

	Values []Type `json:"values,omitempty"` // For "tuple" (element types) or "union" (possible types)

	LiteralValue interface{} `json:"value,omitempty"` // For "literal" (the literal value)
	// Description field from BasicMember is used for "literal" and "type"

	FullFormat bool `json:"full_format,omitempty"` // For "union" (if options have descriptions)

	// Add other type-specific fields as needed
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
func (t Type) IsNamedComplex() bool {
	return t.Name != "" && t.ComplexType != "" // e.g., Name="Color", ComplexType="struct"
}

// Add structs for Prototype API if its top-level structure differs significantly
// type PrototypeAPI struct { ... }
// type Prototype struct { ... } // Prototypes also inherit from BasicMember and have properties




