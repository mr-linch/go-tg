package methodgen

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"text/template"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/docutil"
	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/naming"
	"mvdan.cc/gofumpt/format"
)

//go:embed methods.go.tmpl
var methodsTmpl string

// GoParam represents a resolved method parameter for template rendering.
type GoParam struct {
	Name          string // API param name ("chat_id")
	GoName        string // PascalCase setter name ("ChatID")
	GoArgName     string // camelCase arg name ("chatID")
	GoType        string // resolved type ("PeerID")
	RequestMethod string // request builder method ("PeerID")
	ClassConvert  string // scalar: "AsInputMedia" (method); variadic: "InputMediaOf" (function)
	Variadic      bool   // true → renders as ...XClass, ClassConvert is a function name
	Description   string
	Required      bool
}

// GoConstructor represents a constructor variant for a method.
type GoConstructor struct {
	Suffix         string    // "" for default, "Inline" for inline variants
	RequiredParams []GoParam // params to use in constructor signature
	EmbedType      string    // override embed type for this variant (empty = use method default)
	LinkDefs       []string  // URL link target definitions from param descriptions
}

// GoMethod represents a resolved method for template rendering.
type GoMethod struct {
	Comment      string          // multi-line doc comment
	APIName      string          // e.g., "sendMessage"
	APIDocURL    string          // e.g., "https://core.telegram.org/bots/api#sendmessage"
	CallTypeName string          // e.g., "SendMessageCall"
	GoName       string          // e.g., "SendMessage" (client method name)
	EmbedType    string          // "Call[Message]" or "CallNoResult"
	IsNoResult   bool            // true when returns True
	Params       []GoParam       // all params
	Constructors []GoConstructor // constructor variants (at least one)
}

// RequiredParams returns the required params for the default constructor.
func (m GoMethod) RequiredParams() []GoParam {
	if len(m.Constructors) > 0 {
		return m.Constructors[0].RequiredParams
	}
	var result []GoParam
	for _, p := range m.Params {
		if p.Required {
			result = append(result, p)
		}
	}
	return result
}

// TemplateData is the data passed to the template.
type TemplateData struct {
	Package string
	ir.Metadata
	Methods []GoMethod
}

// Options controls generation behavior.
type Options struct {
	Package string
}

// primitiveGoType maps IR primitive type names to Go type strings.
var primitiveGoType = map[ir.PrimitiveType]string{
	ir.TypeInteger:   "int",
	ir.TypeInteger64: "int64",
	ir.TypeFloat:     "float64",
	ir.TypeString:    "string",
	ir.TypeBoolean:   "bool",
	ir.TypeTrue:      "bool",
	"Int":            "int", // API docs use "Int" in return types
}

// requestMethodMap maps Go types to Request builder method names.
var requestMethodMap = map[string]string{
	"int":        "Int",
	"int64":      "Int64",
	"float64":    "Float64",
	"string":     "String",
	"bool":       "Bool",
	"PeerID":     "PeerID",
	"ChatID":     "ChatID",
	"UserID":     "UserID",
	"FileID":     "FileID",
	"FileArg":    "File",
	"InputMedia": "InputMedia",
}

// Generate writes the generated methods to w.
func Generate(api *ir.API, w io.Writer, cfg *config.MethodGen, log *slog.Logger, opts Options) error {
	if opts.Package == "" {
		opts.Package = "tg"
	}

	rules, err := CompileParamTypeRules(cfg.ParamTypeRules)
	if err != nil {
		return fmt.Errorf("compile param type rules: %w", err)
	}

	stringerTypes := make(map[string]bool)
	for _, t := range cfg.StringerTypes {
		stringerTypes[t] = true
	}

	data := buildTemplateData(api, cfg, rules, stringerTypes, log)
	data.Package = opts.Package
	data.Metadata = api.Metadata

	log.Info("generating methods", "count", len(data.Methods))

	funcMap := template.FuncMap{
		"last": func(i int, slice []GoParam) bool {
			return i == len(slice)-1
		},
	}

	tmpl, err := template.New("methods").Funcs(funcMap).Parse(ir.HeaderTemplate)
	if err != nil {
		return fmt.Errorf("parse header template: %w", err)
	}
	tmpl, err = tmpl.Parse(methodsTmpl)
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

func buildTemplateData(api *ir.API, cfg *config.MethodGen, rules *CompiledParamTypeRules, stringerTypes map[string]bool, log *slog.Logger) *TemplateData {
	data := &TemplateData{}

	// Build subtype → union parent map for resolving multi-type params.
	subtypeToUnion := make(map[string]string)
	for _, t := range api.Types {
		for _, st := range t.Subtypes {
			subtypeToUnion[st] = t.Name
		}
	}

	// Build lookup maps for Go doc link resolution.
	knownTypes := make(map[string]bool, len(api.Types))
	for _, t := range api.Types {
		knownTypes[naming.NormalizeTypeName(t.Name)] = true
	}
	knownMethods := make(map[string]string, len(api.Methods))
	for _, m := range api.Methods {
		knownMethods[m.Name] = "Client." + naming.MethodName(m.Name)
	}

	// Compute discriminator union types that have Class interfaces.
	classTypes := collectClassTypes(api)

	usedParamOverrides := make(map[string]bool)
	usedReturnOverrides := make(map[string]bool)

	for _, m := range api.Methods {
		goMethod := resolveMethod(m, cfg, rules, stringerTypes, subtypeToUnion, classTypes, usedParamOverrides, usedReturnOverrides, knownTypes, knownMethods)
		data.Methods = append(data.Methods, goMethod)
	}

	// Warn about unused config entries.
	for key := range cfg.ParamTypeOverrides {
		if !usedParamOverrides[key] {
			log.Warn("unused param_type_override", "key", key)
		}
	}
	for key := range cfg.ReturnTypeOverrides {
		if !usedReturnOverrides[key] {
			log.Warn("unused return_type_override", "key", key)
		}
	}
	for _, expr := range rules.Unmatched() {
		log.Warn("unmatched param_type_rule", "expr", expr)
	}

	return data
}

func resolveMethod(m ir.Method, cfg *config.MethodGen, rules *CompiledParamTypeRules, stringerTypes map[string]bool, subtypeToUnion map[string]string, classTypes, usedParamOverrides, usedReturnOverrides, knownTypes map[string]bool, knownMethods map[string]string) GoMethod {
	goName := naming.MethodName(m.Name)
	callTypeName := goName + "Call"

	// Resolve params.
	params := make([]GoParam, 0, len(m.Params))
	for _, p := range m.Params {
		goParam := resolveParam(m.Name, p, cfg, rules, stringerTypes, subtypeToUnion, classTypes, usedParamOverrides)
		params = append(params, goParam)
	}

	// Build constructor variants with link extraction from param descriptions.
	constructors := buildConstructors(m.Name, params, cfg.ConstructorVariants, knownTypes, knownMethods)

	// Resolve return type - use first variant's return_type if specified.
	var embedType string
	var isNoResult bool
	if len(constructors) > 0 && constructors[0].EmbedType != "" {
		embedType = constructors[0].EmbedType
		isNoResult = embedType == "CallNoResult"
	} else {
		embedType, isNoResult = resolveReturnType(m, cfg.ReturnTypeOverrides, usedReturnOverrides)
	}

	comment := formatMethodComment(m.Description, knownTypes, knownMethods)

	apiDocURL := "https://core.telegram.org/bots/api#" + strings.ToLower(m.Name)

	return GoMethod{
		Comment:      comment,
		APIName:      m.Name,
		APIDocURL:    apiDocURL,
		CallTypeName: callTypeName,
		GoName:       goName,
		EmbedType:    embedType,
		IsNoResult:   isNoResult,
		Params:       params,
		Constructors: constructors,
	}
}

// buildConstructors builds the constructor variants for a method.
func buildConstructors(methodName string, params []GoParam, variants map[string][]config.ConstructorVariant, knownTypes map[string]bool, knownMethods map[string]string) []GoConstructor {
	// Check for explicit variants in config.
	if cfgVariants, ok := variants[methodName]; ok {
		constructors := make([]GoConstructor, 0, len(cfgVariants))
		for _, v := range cfgVariants {
			var requiredParams []GoParam
			for _, paramName := range v.RequiredParams {
				for _, p := range params {
					if p.Name == paramName {
						requiredParams = append(requiredParams, p)
						break
					}
				}
			}
			// Resolve embed type from variant's return_type if specified.
			var embedType string
			if v.ReturnType != "" {
				embedType = returnTypeToEmbed(v.ReturnType)
			}
			rp, linkDefs := convertParamLinks(requiredParams, knownTypes, knownMethods)
			constructors = append(constructors, GoConstructor{
				Suffix:         v.Suffix,
				RequiredParams: rp,
				EmbedType:      embedType,
				LinkDefs:       linkDefs,
			})
		}
		return constructors
	}

	// Default: single constructor with API-specified required params.
	var requiredParams []GoParam
	for _, p := range params {
		if p.Required {
			requiredParams = append(requiredParams, p)
		}
	}
	rp, linkDefs := convertParamLinks(requiredParams, knownTypes, knownMethods)
	return []GoConstructor{{
		Suffix:         "",
		RequiredParams: rp,
		LinkDefs:       linkDefs,
	}}
}

// returnTypeToEmbed converts a return type name to an embed type string.
func returnTypeToEmbed(returnType string) string {
	if returnType == "True" || returnType == "" {
		return "CallNoResult"
	}
	return "Call[" + returnType + "]"
}

func resolveReturnType(m ir.Method, overrides map[string]string, usedOverrides map[string]bool) (embedType string, isNoResult bool) {
	// Check override first.
	if override, ok := overrides[m.Name]; ok {
		usedOverrides[m.Name] = true
		return "Call[" + override + "]", false
	}

	// Check for True (no result).
	if len(m.Returns.Types) > 0 && m.Returns.Types[0].Type == "True" {
		return "CallNoResult", true
	}

	// Resolve the return type.
	if len(m.Returns.Types) == 0 {
		return "CallNoResult", true
	}

	baseType := resolveBaseType(m.Returns.Types[0].Type)
	returnType := applyArray(baseType, m.Returns.Array)

	return "Call[" + returnType + "]", false
}

func resolveParam(methodName string, p ir.Param, cfg *config.MethodGen, rules *CompiledParamTypeRules, stringerTypes map[string]bool, subtypeToUnion map[string]string, classTypes, usedOverrides map[string]bool) GoParam {
	// Resolve names.
	goName := naming.SnakeToPascal(p.Name)
	goArgName := naming.EscapeReserved(naming.SnakeToCamel(p.Name))

	// Resolve type.
	goType := resolveParamType(methodName, p, cfg, rules, subtypeToUnion, usedOverrides)

	// Determine request method from the concrete type (before Class interface conversion).
	requestMethod := resolveRequestMethod(goType, stringerTypes)

	// Check for class type conversion.
	var classConvert string
	var variadic bool
	if classTypes[goType] {
		// Scalar union param → XClass interface with .AsX() method call
		classConvert = "As" + goType
		goType += "Class"
	} else if strings.HasPrefix(goType, "[]") {
		baseType := goType[2:]
		if classTypes[baseType] {
			// Slice union param → variadic ...XClass with XOf() conversion
			classConvert = baseType + "Of"
			goType = baseType + "Class"
			variadic = true
		}
	}

	return GoParam{
		Name:          p.Name,
		GoName:        goName,
		GoArgName:     goArgName,
		GoType:        goType,
		RequestMethod: requestMethod,
		ClassConvert:  classConvert,
		Variadic:      variadic,
		Description:   p.Description,
		Required:      p.Required,
	}
}

func resolveParamType(methodName string, p ir.Param, cfg *config.MethodGen, rules *CompiledParamTypeRules, subtypeToUnion map[string]string, usedOverrides map[string]bool) string {
	// Check explicit override first.
	key := methodName + "." + p.Name
	if override, ok := cfg.ParamTypeOverrides[key]; ok {
		usedOverrides[key] = true
		return override
	}

	// Check compiled rules.
	if matched, ok := rules.Match(methodName, p); ok {
		return matched
	}

	// Default resolution.
	if len(p.TypeExpr.Types) == 0 {
		return "any"
	}

	// For unions with multiple types, check if all are subtypes of a single union.
	if len(p.TypeExpr.Types) > 1 {
		if union := findCommonUnion(p.TypeExpr.Types, subtypeToUnion); union != "" {
			return applyArray(resolveBaseType(union), p.TypeExpr.Array)
		}
		return "any"
	}

	baseType := resolveBaseType(p.TypeExpr.Types[0].Type)
	return applyArray(baseType, p.TypeExpr.Array)
}

// findCommonUnion returns the union parent type name if all types are subtypes of the same union.
func findCommonUnion(types []ir.TypeRef, subtypeToUnion map[string]string) string {
	if len(types) == 0 {
		return ""
	}
	union := subtypeToUnion[types[0].Type]
	if union == "" {
		return ""
	}
	for _, t := range types[1:] {
		if subtypeToUnion[t.Type] != union {
			return ""
		}
	}
	return union
}

func resolveBaseType(name string) string {
	if goType, ok := primitiveGoType[ir.PrimitiveType(name)]; ok {
		return goType
	}
	return naming.NormalizeTypeName(name)
}

func applyArray(base string, arrayDepth int) string {
	if arrayDepth > 0 {
		result := base
		for range arrayDepth {
			result = "[]" + result
		}
		return result
	}
	return base
}

func resolveRequestMethod(goType string, stringerTypes map[string]bool) string {
	// Check direct mapping first.
	if method, ok := requestMethodMap[goType]; ok {
		return method
	}

	// Check stringer types.
	if stringerTypes[goType] {
		return "Stringer"
	}

	// Special cases for media slices that need file extraction.
	switch goType {
	case "[]InputMedia":
		return "InputMediaSlice"
	case "[]InputPaidMedia":
		return "InputPaidMediaSlice"
	}

	// Default to JSON for everything else.
	return "JSON"
}

func formatMethodComment(desc []string, knownTypes map[string]bool, knownMethods map[string]string) string {
	if len(desc) == 0 {
		return ""
	}
	joined := strings.Join(desc, "\n")
	converted := docutil.ConvertLinks(joined, knownTypes, knownMethods)
	// Re-split and re-join with "// " prefix for continuation lines.
	lines := strings.Split(converted, "\n")
	if len(lines) <= 1 {
		return converted
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

// convertParamLinks processes markdown links in param descriptions.
// It converts links in-place and collects URL link target definitions separately.
func convertParamLinks(params []GoParam, knownTypes map[string]bool, knownMethods map[string]string) (result []GoParam, allLinkDefs []string) {
	seen := map[string]bool{}
	result = make([]GoParam, len(params))
	copy(result, params)
	for i := range result {
		converted, linkDefs := docutil.ExtractLinks(result[i].Description, knownTypes, knownMethods)
		result[i].Description = converted
		for _, ld := range linkDefs {
			if !seen[ld] {
				seen[ld] = true
				allLinkDefs = append(allLinkDefs, ld)
			}
		}
	}
	return result, allLinkDefs
}

// collectClassTypes returns the set of type names that are discriminator unions
// with constructors (i.e., used in method parameters). These types have
// corresponding XClass interfaces generated in types_gen.go.
func collectClassTypes(api *ir.API) map[string]bool {
	typeMap := make(map[string]*ir.Type, len(api.Types))
	for i := range api.Types {
		typeMap[api.Types[i].Name] = &api.Types[i]
	}

	inputTypes := collectInputTypes(api, typeMap)

	// Filter to discriminator unions: has subtypes, no own fields,
	// subtypes have a Const field (discriminator).
	result := make(map[string]bool)
	for _, t := range api.Types {
		if len(t.Subtypes) == 0 || len(t.Fields) > 0 || !inputTypes[t.Name] {
			continue
		}
		if isDiscriminatorUnion(t, typeMap) {
			result[t.Name] = true
		}
	}
	return result
}

// collectInputTypes returns the set of type names reachable from method parameters.
func collectInputTypes(api *ir.API, typeMap map[string]*ir.Type) map[string]bool {
	inputTypes := make(map[string]bool)
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
	for len(queue) > 0 {
		typeName := queue[0]
		queue = queue[1:]
		t := typeMap[typeName]
		if t == nil {
			continue
		}
		for _, st := range t.Subtypes {
			if !inputTypes[st] {
				inputTypes[st] = true
				queue = append(queue, st)
			}
		}
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

// isDiscriminatorUnion checks if a union type has a discriminator field (Const) in any subtype.
func isDiscriminatorUnion(t ir.Type, typeMap map[string]*ir.Type) bool {
	for _, stName := range t.Subtypes {
		st := typeMap[stName]
		if st == nil {
			continue
		}
		for _, f := range st.Fields {
			if f.Const != "" {
				return true
			}
		}
	}
	return false
}
