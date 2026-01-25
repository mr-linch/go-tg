package ir

// PrimitiveType represents a Telegram Bot API primitive type name.
type PrimitiveType string

const (
	TypeInteger   PrimitiveType = "Integer"
	TypeInteger64 PrimitiveType = "Integer64" // Integer with 52-bit note in description
	TypeFloat     PrimitiveType = "Float"     // "Float number" in docs, normalized
	TypeString    PrimitiveType = "String"
	TypeBoolean   PrimitiveType = "Boolean"
	TypeTrue      PrimitiveType = "True" // always-true boolean literal
)

// IsPrimitive reports whether the given type name is a known primitive type.
func IsPrimitive(typ string) bool {
	switch PrimitiveType(typ) {
	case TypeInteger, TypeInteger64, TypeFloat, TypeString, TypeBoolean, TypeTrue:
		return true
	}
	return false
}

// TypeRef represents a single type reference (name + optional anchor).
type TypeRef struct {
	Type string `yaml:"type"`
	Ref  string `yaml:"ref,omitempty"`
}

// TypeExpr represents a type expression: possibly an array of a possibly-union type.
type TypeExpr struct {
	Types []TypeRef `yaml:"types"`
	Array int       `yaml:"array,omitempty"`
}

// Enum represents a named set of string constants.
type Enum struct {
	Name   string   `yaml:"name"`
	Values []string `yaml:"values"`
	Fields []string `yaml:"fields,omitempty"` // "TypeName.field_name" references
}

// API is the top-level intermediate representation of the Telegram Bot API.
type API struct {
	Types   []Type   `yaml:"types,omitempty"`
	Methods []Method `yaml:"methods,omitempty"`
	Enums   []Enum   `yaml:"enums,omitempty"`
}

// Type represents a Telegram Bot API type (struct or union).
type Type struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Fields      []Field  `yaml:"fields,omitempty"`
	Subtypes    []string `yaml:"subtypes,omitempty"`
}

// Field represents a field in a type definition.
type Field struct {
	Name        string   `yaml:"name"`
	TypeExpr    TypeExpr `yaml:"type"`
	Optional    bool     `yaml:"optional,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Const       string   `yaml:"const,omitempty"` // discriminator: always this value
	Enum        []string `yaml:"enum,omitempty"`  // allowed values set
}

// Method represents a Telegram Bot API method.
type Method struct {
	Name        string   `yaml:"name"`
	Description []string `yaml:"description,omitempty"`
	Params      []Param  `yaml:"params,omitempty"`
	Returns     TypeExpr `yaml:"returns"`
}

// Param represents a method parameter.
type Param struct {
	Name        string   `yaml:"name"`
	TypeExpr    TypeExpr `yaml:"type"`
	Required    bool     `yaml:"required,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Enum        []string `yaml:"enum,omitempty"` // allowed values set (from <em> tags)
}
