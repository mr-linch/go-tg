package typegen

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"text/template"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
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
	Comment       string
	Name          string
	Variants      []GoUnionVariant
	Discriminator string
}

// GoUnionVariant represents one arm of a union type.
type GoUnionVariant struct {
	FieldName string
	TypeName  string
	ConstVal  string
}

// TemplateData is the data passed to the template.
type TemplateData struct {
	Package    string
	Types      []GoType
	UnionTypes []GoUnionType
	NeedJSON   bool
	NeedFmt    bool
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

	log.Info("generating types", "structs", len(data.Types), "unions", len(data.UnionTypes))

	tmpl, err := template.New("types").Parse(typesTmpl)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	return tmpl.Execute(w, data)
}

func buildTemplateData(api *ir.API, cfg *config.TypeGen, rules *CompiledFieldTypeRules, log *slog.Logger) *TemplateData {
	data := &TemplateData{}

	// Track which config entries are actually used.
	usedExcludes := make(map[string]bool)
	usedNameOverrides := make(map[string]bool)
	usedTypeOverrides := make(map[string]bool)

	for _, t := range api.Types {
		if cfg.IsExcluded(t.Name) {
			log.Debug("excluding type", "name", t.Name)
			usedExcludes[t.Name] = true
			continue
		}

		if len(t.Subtypes) > 0 && len(t.Fields) == 0 {
			if u := resolveUnionType(t, api.Types); u != nil {
				log.Debug("generating union", "name", t.Name, "variants", len(u.Variants))
				data.UnionTypes = append(data.UnionTypes, *u)
				data.NeedJSON = true
				data.NeedFmt = true
			}
			continue
		}

		if len(t.Fields) == 0 {
			continue
		}

		log.Debug("generating type", "name", t.Name)
		goType := resolveType(t, cfg, rules, usedNameOverrides, usedTypeOverrides)
		data.Types = append(data.Types, goType)
	}

	// Warn about unused config entries.
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

	return data
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
	return fmt.Sprintf("`json:\"%s\"`", fieldName)
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

func resolveUnionType(t ir.Type, allTypes []ir.Type) *GoUnionType {
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
	union := &GoUnionType{
		Name:          name,
		Comment:       formatTypeComment(name, t.Description),
		Discriminator: discField,
	}

	for _, stName := range t.Subtypes {
		st := findType(allTypes, stName)
		if st == nil {
			continue
		}
		constVal := getConstValue(st, discField)
		normalizedSt := normalizeTypeName(stName)
		fieldName := strings.TrimPrefix(normalizedSt, name)
		union.Variants = append(union.Variants, GoUnionVariant{
			FieldName: fieldName,
			TypeName:  normalizedSt,
			ConstVal:  constVal,
		})
	}

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
