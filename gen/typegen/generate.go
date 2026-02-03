package typegen

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"text/template"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
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
	FieldName string
	TypeName  string
	ConstVal  string
	EnumConst string // Enum constant name, e.g. "MessageOriginTypeUser"
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
	Enums               []GoEnum
	TypeMethods         []GoTypeMethod
	VariantConstructors []GoVariantConstructorType
	NeedJSON            bool
	NeedFmt             bool
}

// Options controls generation behavior.
type Options struct {
	// Package is the Go package name for the generated file (default: "tg").
	Package string
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

	for _, t := range api.Types {
		if cfg.IsExcluded(t.Name) {
			log.Debug("excluding type", "name", t.Name)
			usedExcludes[t.Name] = true
			continue
		}

		if len(t.Subtypes) > 0 && len(t.Fields) == 0 {
			if u := resolveUnionType(t, api.Types, inputTypes); u != nil {
				log.Debug("generating union", "name", t.Name, "variants", len(u.Variants), "hasConstructors", u.HasConstructors)
				data.UnionTypes = append(data.UnionTypes, *u)
				data.NeedJSON = true
				data.NeedFmt = true
			}
			continue
		}

		log.Debug("generating type", "name", t.Name)
		goType := resolveType(t, cfg, rules, usedNameOverrides, usedTypeOverrides)
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
		vc := buildVariantConstructors(irType, &vcCfg, cfg, rules, usedNameOverrides, usedTypeOverrides, log)
		if vc != nil {
			data.VariantConstructors = append(data.VariantConstructors, *vc)
			log.Debug("generating variant constructors", "type", vcCfg.Type, "variants", len(vc.Variants))
		}
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

func resolveType(t ir.Type, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool) GoType {
	name := normalizeTypeName(t.Name)
	gt := GoType{
		Comment: formatTypeComment(name, t.Description),
		Name:    name,
	}

	for _, f := range t.Fields {
		gf := resolveField(t.Name, f, cfg, rules, usedNames, usedTypes)
		gt.Fields = append(gt.Fields, gf)
	}

	return gt
}

func resolveField(typeName string, f ir.Field, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool) GoField {
	goName := resolveFieldName(typeName, f.Name, cfg, usedNames)
	goType := resolveGoType(typeName, f, cfg, rules, usedTypes)
	jsonTag := buildJSONTag(f.Name, f.Optional)
	comment := formatFieldComment(f.Description)

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
	return snakeToPascal(fieldName)
}

func buildJSONTag(fieldName string, optional bool) string {
	if optional {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", fieldName)
	}
	return fmt.Sprintf("`json:%q`", fieldName)
}

func formatTypeComment(name, desc string) string {
	if desc == "" {
		return name
	}
	runes := []rune(desc)
	first := strings.ToLower(string(runes[0]))
	text := name + " " + first + string(runes[1:])
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
func formatFieldComment(s string) string {
	return wrapComment(s)
}

func resolveUnionType(t ir.Type, allTypes []ir.Type, inputTypes map[string]bool) *GoUnionType {
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

	name := normalizeTypeName(t.Name)
	discGoField := snakeToPascal(discField)
	discTypeName := name + discGoField
	union := &GoUnionType{
		Name:                    name,
		Comment:                 formatTypeComment(name, t.Description),
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
		normalizedSt := normalizeTypeName(stName)
		fieldName := strings.TrimPrefix(normalizedSt, name)
		// Use fieldName for enum constant to ensure uniqueness (some unions have
		// variants with duplicate discriminator values, e.g. InlineQueryResult)
		enumConst := discTypeName + fieldName
		union.Variants = append(union.Variants, GoUnionVariant{
			FieldName: fieldName,
			TypeName:  normalizedSt,
			ConstVal:  constVal,
			EnumConst: enumConst,
		})

		if seenConstVals[constVal] {
			hasDuplicates = true
		}
		seenConstVals[constVal] = true
	}

	// Can only generate UnmarshalJSON if discriminator values are unique
	union.CanUnmarshal = !hasDuplicates

	return union
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
		constName := irEnum.Name + snakeToPascal(val)
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

// resolveTypeMethod builds a GoTypeMethod from a type and its corresponding enum.
func resolveTypeMethod(t *ir.Type, enum *ir.Enum, cfg *config.TypeMethodDef) GoTypeMethod {
	method := GoTypeMethod{
		TypeName:   normalizeTypeName(t.Name),
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

		goFieldName := snakeToPascal(val)
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
func buildVariantConstructors(t *ir.Type, vcCfg *config.VariantConstructorDef, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedNames, usedTypes map[string]bool, log *slog.Logger) *GoVariantConstructorType {
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

	typeName := normalizeTypeName(t.Name)
	baseGoField := snakeToPascal(baseField.Name)
	baseGoType := resolveGoType(t.Name, *baseField, cfg, rules, usedTypes)
	baseArgName := snakeToCamel(baseField.Name)

	vc := &GoVariantConstructorType{
		TypeName:    typeName,
		BaseGoField: baseGoField,
		BaseGoType:  baseGoType,
		BaseParam:   baseArgName + " " + baseGoType,
		BaseArg:     baseArgName,
	}

	// Check if field should be excluded.
	isExcluded := func(fieldName string) bool {
		for _, ex := range vcCfg.Exclude {
			if ex == fieldName {
				return true
			}
		}
		return false
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
		goType := resolveGoType(t.Name, f, cfg, rules, usedTypes)
		argName := snakeToCamel(f.Name)

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
