package typegen

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"text/template"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/docutil"
	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/naming"
	"mvdan.cc/gofumpt/format"
)

//go:embed types.go.tmpl
var typesTmpl string

// GoField represents a resolved struct field for template rendering.
type GoField struct {
	Comment string
	Name    string
	Type    string
	JSONTag string
}

// GoType represents a resolved struct type for template rendering.
type GoType struct {
	Comment string
	Name    string
	Fields  []GoField
}

// GoUnionType represents a union type (has Subtypes, no Fields).
type GoUnionType struct {
	Comment                 string
	Name                    string
	Variants                []GoUnionVariant
	Discriminator           string // JSON field name, e.g. "type"
	DiscriminatorGoField    string // Go field name on subtypes, e.g. "Type"
	DiscriminatorTypeName   string // Enum type name, e.g. "MessageOriginType"
	DiscriminatorMethodName string // Method name on union, e.g. "Type"
	CanUnmarshal            bool   // true if discriminator values are unique (can generate UnmarshalJSON)
	HasConstructors         bool   // true if this union is used in method params (generate New* functions)
}

// GoUnionVariant represents one arm of a union type.
type GoUnionVariant struct {
	FieldName      string
	TypeName       string
	ConstVal       string
	EnumConst      string               // Enum constant name, e.g. "MessageOriginTypeUser"
	RequiredFields []GoConstructorParam // non-discriminator required fields (for constructors)
}

// GoConstructorParam represents a required parameter for a union variant constructor.
type GoConstructorParam struct {
	GoField   string // Go field name, e.g. "ChatID"
	GoType    string // Go type string, e.g. "ChatID"
	ParamName string // camelCase parameter name, e.g. "chatID"
}

// GoInterfaceUnion represents a union type without a discriminator (marker interface pattern).
type GoInterfaceUnion struct {
	Comment      string
	Name         string
	MarkerMethod string // unexported marker method, e.g. "isReplyMarkup"
	Variants     []GoInterfaceUnionVariant
}

// GoInterfaceUnionVariant represents one arm of an interface union.
type GoInterfaceUnionVariant struct {
	TypeName    string                  // Go type name, e.g. "InlineKeyboardMarkup"
	Constructor *GoInterfaceConstructor // constructor for this variant (nil if no constructor)
}

// GoInterfaceConstructor holds data for generating a constructor function for an interface union variant.
type GoInterfaceConstructor struct {
	FuncName  string                           // e.g. "NewReplyKeyboardMarkup"
	TypeName  string                           // e.g. "ReplyKeyboardMarkup"
	ReturnPtr bool                             // true if type has optional fields (return *T for chaining)
	Params    []GoInterfaceConstructorParam    // required fields as params
	Sentinels []GoInterfaceConstructorSentinel // required bool fields auto-set to true
}

// GoInterfaceConstructorParam represents a required parameter for an interface union constructor.
type GoInterfaceConstructorParam struct {
	GoField string // Go field name, e.g. "Keyboard"
	Name    string // param name, e.g. "keyboard"
	Type    string // Go type, e.g. "...[]KeyboardButton" (variadic for ArrayDepth >= 2)
}

// GoInterfaceConstructorSentinel represents a required bool field that is auto-set to true.
type GoInterfaceConstructorSentinel struct {
	GoField string // Go field name, e.g. "RemoveKeyboard"
	Value   string // e.g. "true"
}

// GoBuilderType holds With* methods for a single type.
type GoBuilderType struct {
	TypeName string
	Methods  []GoBuilderMethod
}

// GoBuilderMethod represents a single With* method on a type.
type GoBuilderMethod struct {
	FieldName string // "Caption"
	ParamName string // "caption" (empty for bool)
	GoType    string // "string" (empty for bool)
	IsBool    bool   // no param, sets to true
	IsPtr     bool   // true if field is a pointer type (accept value, assign &param)
}

// GoEnum represents a standalone enum type for template rendering.
type GoEnum struct {
	Name       string
	Underlying string // e.g. "int8"
	Values     []GoEnumValue
	HasUnknown bool   // include Unknown sentinel at 0
	Marshal    string // "json", "text", "stringer", ""
}

// GoEnumValue represents a single enum constant.
type GoEnumValue struct {
	ConstName string // e.g. "ChatTypePrivate"
	StringVal string // e.g. "private"
}

// GoTypeMethod represents a method on a type that returns an enum based on optional fields.
type GoTypeMethod struct {
	TypeName   string             // e.g. "Message"
	MethodName string             // e.g. "Type"
	ReturnType string             // e.g. "MessageType"
	Cases      []GoTypeMethodCase // switch cases
}

// GoTypeMethodCase represents a single case in a Type() method switch.
type GoTypeMethodCase struct {
	FieldName string // Go field name, e.g. "Text"
	EnumConst string // Enum constant, e.g. "MessageTypeText"
	Condition string // condition expression, e.g. "v.Text != \"\""
}

// GoVariantConstructorType holds all variants for a single type.
type GoVariantConstructorType struct {
	TypeName    string                 // e.g. "InlineKeyboardButton"
	BaseGoField string                 // e.g. "Text"
	BaseGoType  string                 // e.g. "string"
	BaseParam   string                 // e.g. "text string"
	BaseArg     string                 // e.g. "text"
	Variants    []GoVariantConstructor // list of variants
}

// GoVariantConstructor represents a single variant constructor.
type GoVariantConstructor struct {
	Name         string // e.g. "URL" -> NewInlineKeyboardButtonURL
	GoField      string // e.g. "URL"
	GoType       string // e.g. "string"
	Param        string // e.g. "url string" (empty if HasDefault)
	Arg          string // e.g. "url" (empty if HasDefault)
	HasDefault   bool   // true if uses default value instead of param
	DefaultValue string // e.g. "true" or "&CallbackGame{}"
	Comment      string // e.g. "with URL"
}

// TemplateData is the data passed to the template.
type TemplateData struct {
	Package string
	ir.Metadata
	Types               []GoType
	UnionTypes          []GoUnionType
	InterfaceUnions     []GoInterfaceUnion
	Enums               []GoEnum
	TypeMethods         []GoTypeMethod
	VariantConstructors []GoVariantConstructorType
	BuilderTypes        []GoBuilderType
	NeedJSON            bool
	NeedFmt             bool
}

// Options controls generation behavior.
type Options struct {
	// Package is the Go package name for the generated file (default: "tg").
	Package string
	// SkipFormat disables gofumpt formatting of the output (useful for debugging).
	SkipFormat bool
}

// Generate writes the generated types to w.
func Generate(api *ir.API, w io.Writer, cfg *config.TypeGen, log *slog.Logger, opts Options) error {
	if opts.Package == "" {
		opts.Package = "tg"
	}

	rules, err := CompileFieldTypeRules(cfg.FieldTypeRules)
	if err != nil {
		return fmt.Errorf("compile field type rules: %w", err)
	}

	data := buildTemplateData(api, cfg, rules, log)
	data.Package = opts.Package
	data.Metadata = api.Metadata

	log.Info("generating types", "structs", len(data.Types), "unions", len(data.UnionTypes))

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
	}
	tmpl, err := template.New("types").Funcs(funcMap).Parse(ir.HeaderTemplate)
	if err != nil {
		return fmt.Errorf("parse header template: %w", err)
	}
	tmpl, err = tmpl.Parse(typesTmpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if opts.SkipFormat {
		_, err = w.Write(buf.Bytes())
		return err
	}

	formatted, err := format.Source(buf.Bytes(), format.Options{})
	if err != nil {
		return fmt.Errorf("format source: %w", err)
	}

	_, err = w.Write(formatted)
	return err
}

func buildTemplateData(api *ir.API, cfg *config.TypeGen, rules *CompiledFieldTypeRules, log *slog.Logger) *TemplateData {
	data := &TemplateData{}

	// Track which config entries are actually used.
	usedExcludes := make(map[string]bool)
	usedNameOverrides := make(map[string]bool)
	usedTypeOverrides := make(map[string]bool)

	// Collect types used in method parameters (need constructors).
	inputTypes := collectInputTypes(api)

	// Build lookup maps for Go doc link resolution.
	knownTypes := make(map[string]bool, len(api.Types))
	for _, t := range api.Types {
		knownTypes[naming.NormalizeTypeName(t.Name)] = true
	}
	knownMethods := make(map[string]string, len(api.Methods))
	for _, m := range api.Methods {
		knownMethods[m.Name] = "Client." + naming.MethodName(m.Name)
	}

	typeMap := make(map[string]*ir.Type)
	for i := range api.Types {
		typeMap[api.Types[i].Name] = &api.Types[i]
	}

	// Pre-compute interface union type names (interfaces are nilable, no pointer needed).
	interfaceTypes := collectInterfaceTypeNames(api, cfg, typeMap)

	for _, t := range api.Types {
		if cfg.IsExcluded(t.Name) {
			log.Debug("excluding type", "name", t.Name)
			usedExcludes[t.Name] = true
			continue
		}

		if len(t.Subtypes) > 0 && len(t.Fields) == 0 {
			resolveUnionOrInterface(t, data, api.Types, inputTypes, cfg, rules, usedNameOverrides, usedTypeOverrides, typeMap, knownTypes, knownMethods, interfaceTypes, log)
			continue
		}

		log.Debug("generating type", "name", t.Name)
		goType := resolveType(t, cfg, rules, usedNameOverrides, usedTypeOverrides, knownTypes, knownMethods, interfaceTypes)
		data.Types = append(data.Types, goType)
	}

	// Resolve standalone enums.
	for _, enumCfg := range cfg.Enums {
		irEnum := findEnum(api.Enums, enumCfg.Name)
		if irEnum == nil {
			log.Warn("enum not found in IR", "name", enumCfg.Name)
			continue
		}
		goEnum := resolveEnum(irEnum, &enumCfg)
		data.Enums = append(data.Enums, goEnum)
		data.NeedFmt = true
		if enumCfg.Marshal == "json" {
			data.NeedJSON = true
		}
		log.Debug("generating enum", "name", enumCfg.Name, "values", len(goEnum.Values))
	}

	// Resolve type methods (e.g., Message.Type() -> MessageType).
	for _, methodCfg := range cfg.TypeMethods {
		irType := findType(api.Types, methodCfg.Type)
		if irType == nil {
			log.Warn("type not found for type_method", "type", methodCfg.Type)
			continue
		}
		irEnum := findEnum(api.Enums, methodCfg.Return)
		if irEnum == nil {
			log.Warn("enum not found for type_method", "enum", methodCfg.Return)
			continue
		}
		goMethod := resolveTypeMethod(irType, irEnum, &methodCfg)
		data.TypeMethods = append(data.TypeMethods, goMethod)
		log.Debug("generating type method", "type", methodCfg.Type, "method", methodCfg.Method, "cases", len(goMethod.Cases))
	}

	// Resolve variant constructors (e.g., NewInlineKeyboardButtonURL).
	for _, vcCfg := range cfg.VariantConstructors {
		irType := findType(api.Types, vcCfg.Type)
		if irType == nil {
			log.Warn("type not found for variant_constructor", "type", vcCfg.Type)
			continue
		}
		vc := buildVariantConstructors(irType, &vcCfg, cfg, rules, usedNameOverrides, usedTypeOverrides, interfaceTypes, log)
		if vc != nil {
			data.VariantConstructors = append(data.VariantConstructors, *vc)
			log.Debug("generating variant constructors", "type", vcCfg.Type, "variants", len(vc.Variants))
		}
	}

	// Resolve config-defined interface unions (synthetic unions not in spec, e.g. ReplyMarkup).
	for _, iuCfg := range cfg.InterfaceUnions {
		resolveConfigInterfaceUnion(&iuCfg, data, cfg, rules, usedNameOverrides, usedTypeOverrides, typeMap, interfaceTypes, log)
	}

	warnUnusedConfig(cfg, rules, usedExcludes, usedNameOverrides, usedTypeOverrides, log)

	return data
}

func warnUnusedConfig(cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedExcludes, usedNameOverrides, usedTypeOverrides map[string]bool, log *slog.Logger) {
	for _, name := range cfg.Exclude {
		if !usedExcludes[name] {
			log.Warn("unused exclude entry", "name", name)
		}
	}
	for key := range cfg.NameOverrides {
		if !usedNameOverrides[key] {
			log.Warn("unused name_override", "key", key)
		}
	}
	for key := range cfg.TypeOverrides {
		if !usedTypeOverrides[key] {
			log.Warn("unused type_override", "key", key)
		}
	}
	for _, expr := range rules.Unmatched() {
		log.Warn("unmatched field_type_rule", "expr", expr)
	}
}

func resolveType(t ir.Type, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes, knownTypes map[string]bool, knownMethods map[string]string, interfaceTypes map[string]bool) GoType {
	name := naming.NormalizeTypeName(t.Name)
	gt := GoType{
		Comment: formatTypeComment(name, t.Description, knownTypes, knownMethods),
		Name:    name,
	}

	for _, f := range t.Fields {
		// Skip discriminator fields — they are handled by the union's MarshalJSON/UnmarshalJSON.
		if f.Const != "" {
			continue
		}
		gf := resolveField(t.Name, f, cfg, rules, usedNames, usedTypes, knownTypes, knownMethods, interfaceTypes)
		gt.Fields = append(gt.Fields, gf)
	}

	return gt
}

func resolveField(typeName string, f ir.Field, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes, knownTypes map[string]bool, knownMethods map[string]string, interfaceTypes map[string]bool) GoField {
	goName := resolveFieldName(typeName, f.Name, cfg, usedNames)
	goType := resolveGoType(typeName, f, cfg, rules, usedTypes, interfaceTypes)
	jsonTag := buildJSONTag(f.Name, f.Optional)
	comment := formatFieldComment(f.Description, knownTypes, knownMethods)

	return GoField{
		Comment: comment,
		Name:    goName,
		Type:    goType,
		JSONTag: jsonTag,
	}
}

func resolveFieldName(typeName, fieldName string, cfg *config.TypeGen, usedNameOverrides map[string]bool) string {
	key := typeName + "." + fieldName
	if override, ok := cfg.NameOverrides[key]; ok {
		usedNameOverrides[key] = true
		return override
	}
	return naming.SnakeToPascal(fieldName)
}

func buildJSONTag(fieldName string, optional bool) string {
	if optional {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", fieldName)
	}
	return fmt.Sprintf("`json:%q`", fieldName)
}

func formatTypeComment(name, desc string, knownTypes map[string]bool, knownMethods map[string]string) string {
	if desc == "" {
		return name
	}
	runes := []rune(desc)
	first := strings.ToLower(string(runes[0]))
	text := name + " " + first + string(runes[1:])
	text = docutil.ConvertLinks(text, knownTypes, knownMethods)
	return wrapComment(text)
}

// wrapComment ensures multi-line text is properly formatted as Go comments.
// Each line after the first gets a "// " prefix.
func wrapComment(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= 1 {
		return s
	}
	var sb strings.Builder
	for i, line := range lines {
		if i > 0 {
			sb.WriteString("\n// ")
		}
		sb.WriteString(line)
	}
	return sb.String()
}

// formatFieldComment formats a field description for use as a Go comment.
func formatFieldComment(s string, knownTypes map[string]bool, knownMethods map[string]string) string {
	return wrapComment(docutil.ConvertLinks(s, knownTypes, knownMethods))
}

func resolveUnionType(t ir.Type, allTypes []ir.Type, inputTypes map[string]bool, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes, knownTypes map[string]bool, knownMethods map[string]string) *GoUnionType {
	firstSubtype := findType(allTypes, t.Subtypes[0])
	if firstSubtype == nil {
		return nil
	}

	var discField string
	for _, f := range firstSubtype.Fields {
		if f.Const != "" {
			discField = f.Name
			break
		}
	}
	if discField == "" {
		return nil
	}

	name := naming.NormalizeTypeName(t.Name)
	discGoField := naming.SnakeToPascal(discField)
	discTypeName := name + discGoField
	union := &GoUnionType{
		Name:                    name,
		Comment:                 formatTypeComment(name, t.Description, knownTypes, knownMethods),
		Discriminator:           discField,
		DiscriminatorGoField:    discGoField,
		DiscriminatorTypeName:   discTypeName,
		DiscriminatorMethodName: discGoField,
		HasConstructors:         inputTypes[t.Name],
	}

	seenConstVals := make(map[string]bool)
	hasDuplicates := false

	for _, stName := range t.Subtypes {
		st := findType(allTypes, stName)
		if st == nil {
			continue
		}
		constVal := getConstValue(st, discField)
		normalizedSt := naming.NormalizeTypeName(stName)
		fieldName := strings.TrimPrefix(normalizedSt, name)
		// Use fieldName for enum constant to ensure uniqueness (some unions have
		// variants with duplicate discriminator values, e.g. InlineQueryResult)
		enumConst := discTypeName + fieldName

		variant := GoUnionVariant{
			FieldName: fieldName,
			TypeName:  normalizedSt,
			ConstVal:  constVal,
			EnumConst: enumConst,
		}

		// Collect required non-discriminator fields for constructor params.
		if union.HasConstructors {
			for _, f := range st.Fields {
				if f.Const != "" || f.Optional {
					continue
				}
				goFieldName := resolveFieldName(st.Name, f.Name, cfg, usedNames)
				goType := resolveGoType(st.Name, f, cfg, rules, usedTypes, nil)
				paramName := naming.EscapeReserved(naming.SnakeToCamel(f.Name))
				variant.RequiredFields = append(variant.RequiredFields, GoConstructorParam{
					GoField:   goFieldName,
					GoType:    goType,
					ParamName: paramName,
				})
			}
		}

		union.Variants = append(union.Variants, variant)

		if seenConstVals[constVal] {
			hasDuplicates = true
		}
		seenConstVals[constVal] = true
	}

	// Can only generate UnmarshalJSON if discriminator values are unique
	union.CanUnmarshal = !hasDuplicates

	return union
}

// resolveUnionOrInterface handles a type with subtypes but no fields, resolving it as either a
// discriminator union (struct-based) or an interface union (marker interface).
func resolveUnionOrInterface(t ir.Type, data *TemplateData, allTypes []ir.Type, inputTypes map[string]bool, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool, typeMap map[string]*ir.Type, knownTypes map[string]bool, knownMethods map[string]string, interfaceTypes map[string]bool, log *slog.Logger) {
	if u := resolveUnionType(t, allTypes, inputTypes, cfg, rules, usedNames, usedTypes, knownTypes, knownMethods); u != nil {
		log.Debug("generating union", "name", t.Name, "variants", len(u.Variants), "hasConstructors", u.HasConstructors)
		data.UnionTypes = append(data.UnionTypes, *u)
		data.NeedJSON = true
		data.NeedFmt = true

		// Generate With* builders for discriminator union variants that have constructors.
		if u.HasConstructors {
			appendVariantBuilders(t.Subtypes, data, cfg, rules, usedNames, usedTypes, typeMap, interfaceTypes, log)
		}
		return
	}

	iu := resolveInterfaceUnion(t, cfg, rules, usedNames, usedTypes, typeMap, knownTypes, knownMethods)
	if iu == nil {
		return
	}
	log.Debug("generating interface union", "name", t.Name, "variants", len(iu.Variants))
	data.InterfaceUnions = append(data.InterfaceUnions, *iu)

	// Generate builders for interface union variants.
	for _, stName := range t.Subtypes {
		st := typeMap[stName]
		if st == nil {
			continue
		}
		if bt := resolveBuilderType(st, cfg, rules, usedNames, usedTypes, interfaceTypes); bt != nil {
			data.BuilderTypes = append(data.BuilderTypes, *bt)
			log.Debug("generating builder", "type", bt.TypeName, "methods", len(bt.Methods))
		}
	}
}

// appendVariantBuilders generates With* builder methods for each subtype that has optional fields.
func appendVariantBuilders(subtypes []string, data *TemplateData, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool, typeMap map[string]*ir.Type, interfaceTypes map[string]bool, log *slog.Logger) {
	for _, stName := range subtypes {
		st := typeMap[stName]
		if st == nil {
			continue
		}
		bt := resolveBuilderType(st, cfg, rules, usedNames, usedTypes, interfaceTypes)
		if bt == nil {
			continue
		}
		data.BuilderTypes = append(data.BuilderTypes, *bt)
		log.Debug("generating builder", "type", bt.TypeName, "methods", len(bt.Methods))
	}
}

// resolveConfigInterfaceUnion resolves a config-defined interface union and its variant constructors/builders.
func resolveConfigInterfaceUnion(iuCfg *config.InterfaceUnionDef, data *TemplateData, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool, typeMap map[string]*ir.Type, interfaceTypes map[string]bool, log *slog.Logger) {
	name := naming.NormalizeTypeName(iuCfg.Name)
	iu := GoInterfaceUnion{
		Comment:      name + " is a marker interface for " + iuCfg.Name + " variants.",
		Name:         name,
		MarkerMethod: "is" + name,
	}
	for _, v := range iuCfg.Variants {
		variant := GoInterfaceUnionVariant{
			TypeName: naming.NormalizeTypeName(v),
		}
		// Add constructor and builder if the variant type exists in the API.
		if vt := typeMap[v]; vt != nil {
			variant.Constructor = resolveInterfaceConstructor(vt, cfg, rules, usedNames, usedTypes)
			if bt := resolveBuilderType(vt, cfg, rules, usedNames, usedTypes, interfaceTypes); bt != nil {
				data.BuilderTypes = append(data.BuilderTypes, *bt)
				log.Debug("generating builder", "type", bt.TypeName, "methods", len(bt.Methods))
			}
		}
		iu.Variants = append(iu.Variants, variant)
	}
	data.InterfaceUnions = append(data.InterfaceUnions, iu)
	log.Debug("generating config interface union", "name", name, "variants", len(iu.Variants))
}

// resolveInterfaceUnion creates a GoInterfaceUnion for union types without a discriminator field.
// These are spec types with subtypes but no Const fields (e.g., InputMessageContent).
func resolveInterfaceUnion(t ir.Type, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool, typeMap map[string]*ir.Type, knownTypes map[string]bool, knownMethods map[string]string) *GoInterfaceUnion {
	name := naming.NormalizeTypeName(t.Name)
	iu := &GoInterfaceUnion{
		Comment:      formatTypeComment(name, t.Description, knownTypes, knownMethods),
		Name:         name,
		MarkerMethod: "is" + name,
	}

	for _, stName := range t.Subtypes {
		variant := GoInterfaceUnionVariant{
			TypeName: naming.NormalizeTypeName(stName),
		}
		// Add constructor if the variant type exists in the API.
		if vt := typeMap[stName]; vt != nil {
			variant.Constructor = resolveInterfaceConstructor(vt, cfg, rules, usedNames, usedTypes)
		}
		iu.Variants = append(iu.Variants, variant)
	}

	return iu
}

func findType(types []ir.Type, name string) *ir.Type {
	for i := range types {
		if types[i].Name == name {
			return &types[i]
		}
	}
	return nil
}

func getConstValue(t *ir.Type, fieldName string) string {
	for _, f := range t.Fields {
		if f.Name == fieldName {
			return f.Const
		}
	}
	return ""
}

func findEnum(enums []ir.Enum, name string) *ir.Enum {
	for i := range enums {
		if enums[i].Name == name {
			return &enums[i]
		}
	}
	return nil
}

func resolveEnum(irEnum *ir.Enum, cfg *config.EnumGenDef) GoEnum {
	underlying := cfg.Underlying
	if underlying == "" {
		underlying = "int"
	}
	marshal := cfg.Marshal
	if marshal == "" {
		marshal = "text"
	}

	e := GoEnum{
		Name:       irEnum.Name,
		Underlying: underlying,
		HasUnknown: cfg.Unknown,
		Marshal:    marshal,
	}

	for _, val := range irEnum.Values {
		name := val
		if slug, ok := emojiName(val); ok {
			name = slug
		}
		constName := irEnum.Name + naming.SnakeToPascal(name)
		e.Values = append(e.Values, GoEnumValue{
			ConstName: constName,
			StringVal: val,
		})
	}

	return e
}

// collectInputTypes returns a set of type names that are used in method parameters.
// This recursively includes types referenced in fields of parameter types.
// Union types in this set should have constructor functions generated.
func collectInputTypes(api *ir.API) map[string]bool {
	inputTypes := make(map[string]bool)
	typeMap := make(map[string]*ir.Type)
	for i := range api.Types {
		typeMap[api.Types[i].Name] = &api.Types[i]
	}

	// Collect types from method params.
	var queue []string
	for _, m := range api.Methods {
		for _, p := range m.Params {
			for _, tr := range p.TypeExpr.Types {
				if !ir.IsPrimitive(tr.Type) && !inputTypes[tr.Type] {
					inputTypes[tr.Type] = true
					queue = append(queue, tr.Type)
				}
			}
		}
	}

	// Recursively expand through fields and subtypes.
	for len(queue) > 0 {
		typeName := queue[0]
		queue = queue[1:]

		t := typeMap[typeName]
		if t == nil {
			continue
		}

		// Add subtypes (union variants).
		for _, st := range t.Subtypes {
			if !inputTypes[st] {
				inputTypes[st] = true
				queue = append(queue, st)
			}
		}

		// Add field types.
		for _, f := range t.Fields {
			for _, tr := range f.TypeExpr.Types {
				if !ir.IsPrimitive(tr.Type) && !inputTypes[tr.Type] {
					inputTypes[tr.Type] = true
					queue = append(queue, tr.Type)
				}
			}
		}
	}

	return inputTypes
}

// collectInterfaceTypeNames returns Go type names for all interface union types.
// These types are interfaces (already nilable), so optional fields of these types must not get a pointer.
func collectInterfaceTypeNames(api *ir.API, cfg *config.TypeGen, typeMap map[string]*ir.Type) map[string]bool {
	result := make(map[string]bool)

	// Spec-derived interface unions: types with subtypes, no fields, not excluded, no discriminator.
	for _, t := range api.Types {
		if cfg.IsExcluded(t.Name) || len(t.Subtypes) == 0 || len(t.Fields) > 0 {
			continue
		}
		firstSubtype := typeMap[t.Subtypes[0]]
		if firstSubtype == nil {
			continue
		}
		hasDiscriminator := false
		for _, f := range firstSubtype.Fields {
			if f.Const != "" {
				hasDiscriminator = true
				break
			}
		}
		if !hasDiscriminator {
			result[naming.NormalizeTypeName(t.Name)] = true
		}
	}

	// Config-defined interface unions.
	for _, iuCfg := range cfg.InterfaceUnions {
		result[naming.NormalizeTypeName(iuCfg.Name)] = true
	}

	return result
}

// resolveTypeMethod builds a GoTypeMethod from a type and its corresponding enum.
func resolveTypeMethod(t *ir.Type, enum *ir.Enum, cfg *config.TypeMethodDef) GoTypeMethod {
	method := GoTypeMethod{
		TypeName:   naming.NormalizeTypeName(t.Name),
		MethodName: cfg.Method,
		ReturnType: cfg.Return,
	}

	// Build a map of field name -> field for quick lookup.
	fieldMap := make(map[string]*ir.Field)
	for i := range t.Fields {
		fieldMap[t.Fields[i].Name] = &t.Fields[i]
	}

	// For each enum value, find the corresponding field and generate a case.
	for _, val := range enum.Values {
		f, ok := fieldMap[val]
		if !ok {
			continue // enum value doesn't correspond to a field
		}
		if !f.Optional {
			continue // only optional fields are checked
		}

		goFieldName := naming.SnakeToPascal(val)
		enumConst := cfg.Return + goFieldName
		condition := buildFieldCondition(goFieldName, f)

		method.Cases = append(method.Cases, GoTypeMethodCase{
			FieldName: goFieldName,
			EnumConst: enumConst,
			Condition: condition,
		})
	}

	return method
}

// buildFieldCondition generates the condition expression for a field check.
func buildFieldCondition(goFieldName string, f *ir.Field) string {
	if len(f.TypeExpr.Types) == 0 {
		return "v." + goFieldName + " != nil"
	}

	// Handle array types
	if f.TypeExpr.Array > 0 {
		return "len(v." + goFieldName + ") > 0"
	}

	// Get the primary type
	primaryType := f.TypeExpr.Types[0].Type

	switch primaryType {
	case "String":
		return "v." + goFieldName + ` != ""`
	case "Integer", "Integer64", "Float":
		return "v." + goFieldName + " != 0"
	case "Boolean":
		return "v." + goFieldName
	case "True":
		return "v." + goFieldName
	default:
		// Complex type (pointer)
		return "v." + goFieldName + " != nil"
	}
}

// buildVariantConstructors creates variant constructor data for a type.
func buildVariantConstructors(t *ir.Type, vcCfg *config.VariantConstructorDef, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes, interfaceTypes map[string]bool, log *slog.Logger) *GoVariantConstructorType {
	// Find the base field.
	var baseField *ir.Field
	for i := range t.Fields {
		if t.Fields[i].Name == vcCfg.BaseField {
			baseField = &t.Fields[i]
			break
		}
	}
	if baseField == nil {
		log.Warn("base_field not found for variant_constructor", "type", vcCfg.Type, "base_field", vcCfg.BaseField)
		return nil
	}

	typeName := naming.NormalizeTypeName(t.Name)
	baseGoField := naming.SnakeToPascal(baseField.Name)
	baseGoType := resolveGoType(t.Name, *baseField, cfg, rules, usedTypes, nil)
	baseArgName := naming.SnakeToCamel(baseField.Name)

	vc := &GoVariantConstructorType{
		TypeName:    typeName,
		BaseGoField: baseGoField,
		BaseGoType:  baseGoType,
		BaseParam:   baseArgName + " " + baseGoType,
		BaseArg:     baseArgName,
	}

	// Check if field should be excluded.
	isExcluded := func(fieldName string) bool {
		return slices.Contains(vcCfg.Exclude, fieldName)
	}

	// Generate base constructor if requested (e.g., NewKeyboardButton).
	if vcCfg.IncludeBase {
		vc.Variants = append(vc.Variants, GoVariantConstructor{
			Name:    "",
			Comment: "",
		})
	}

	// Generate variant for each optional field.
	for _, f := range t.Fields {
		if !f.Optional || f.Name == vcCfg.BaseField || isExcluded(f.Name) {
			continue
		}

		goFieldName := resolveFieldName(t.Name, f.Name, cfg, usedNames)
		goType := resolveGoType(t.Name, f, cfg, rules, usedTypes, interfaceTypes)
		argName := naming.SnakeToCamel(f.Name)

		variant := GoVariantConstructor{
			Name:    goFieldName,
			GoField: goFieldName,
			GoType:  goType,
			Comment: "with " + goFieldName,
		}

		// Check if this field has a default value configured.
		if defaultVal, ok := vcCfg.Defaults[f.Name]; ok {
			variant.HasDefault = true
			variant.DefaultValue = defaultVal
		} else {
			// For pointer types, accept value and take address in constructor.
			paramType := goType
			arg := argName
			if strings.HasPrefix(goType, "*") {
				paramType = goType[1:] // Remove "*" from param type
				arg = "&" + argName    // Add "&" when assigning
			}
			variant.Param = argName + " " + paramType
			variant.Arg = arg
		}

		vc.Variants = append(vc.Variants, variant)
	}

	return vc
}

// resolveBuilderType creates a GoBuilderType with With* methods for all optional fields of a type.
// Returns nil if the type has no optional fields.
func resolveBuilderType(t *ir.Type, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes, interfaceTypes map[string]bool) *GoBuilderType {
	typeName := naming.NormalizeTypeName(t.Name)
	var methods []GoBuilderMethod

	for _, f := range t.Fields {
		if f.Const != "" || !f.Optional {
			continue
		}
		goFieldName := resolveFieldName(t.Name, f.Name, cfg, usedNames)
		goType := resolveGoType(t.Name, f, cfg, rules, usedTypes, interfaceTypes)

		m := GoBuilderMethod{
			FieldName: goFieldName,
		}
		if goType == "bool" {
			m.IsBool = true
		} else {
			paramName := naming.EscapeReserved(naming.SnakeToCamel(f.Name))
			// For pointer types, accept value and take address internally.
			if strings.HasPrefix(goType, "*") {
				m.GoType = goType[1:]
				m.IsPtr = true
			} else {
				m.GoType = goType
			}
			m.ParamName = paramName
		}
		methods = append(methods, m)
	}

	if len(methods) == 0 {
		return nil
	}

	return &GoBuilderType{
		TypeName: typeName,
		Methods:  methods,
	}
}

// resolveInterfaceConstructor creates a GoInterfaceConstructor for an interface union variant type.
// Required bool/True fields become sentinels (auto-set to true).
// Required fields with ArrayDepth >= 2 become variadic params.
// Other required fields become regular params.
// ReturnPtr is true if the type has any optional fields (for builder chaining).
func resolveInterfaceConstructor(t *ir.Type, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool) *GoInterfaceConstructor {
	typeName := naming.NormalizeTypeName(t.Name)
	ctor := &GoInterfaceConstructor{
		FuncName: "New" + typeName,
		TypeName: typeName,
	}

	hasOptional := false
	for _, f := range t.Fields {
		if f.Const != "" {
			continue
		}
		if f.Optional {
			hasOptional = true
			continue
		}
		goFieldName := resolveFieldName(t.Name, f.Name, cfg, usedNames)
		goType := resolveGoType(t.Name, f, cfg, rules, usedTypes, nil)

		// Required bool/True fields are sentinels.
		if goType == "bool" {
			ctor.Sentinels = append(ctor.Sentinels, GoInterfaceConstructorSentinel{
				GoField: goFieldName,
				Value:   "true",
			})
			continue
		}

		paramName := naming.EscapeReserved(naming.SnakeToCamel(f.Name))

		// Array depth >= 2 → variadic param (e.g. [][]KeyboardButton → ...[]KeyboardButton)
		if f.TypeExpr.Array >= 2 {
			// Variadic: strip one level of []
			variadicType := goType[2:] // "[][]X" → "[]X"
			ctor.Params = append(ctor.Params, GoInterfaceConstructorParam{
				GoField: goFieldName,
				Name:    paramName,
				Type:    "..." + variadicType,
			})
			continue
		}

		ctor.Params = append(ctor.Params, GoInterfaceConstructorParam{
			GoField: goFieldName,
			Name:    paramName,
			Type:    goType,
		})
	}

	ctor.ReturnPtr = hasOptional
	return ctor
}
