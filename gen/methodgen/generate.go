package methodgen

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"text/template"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/naming"
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
	Description   string
	Required      bool
}

// GoConstructor represents a constructor variant for a method.
type GoConstructor struct {
	Suffix         string    // "" for default, "Inline" for inline variants
	RequiredParams []GoParam // params to use in constructor signature
	EmbedType      string    // override embed type for this variant (empty = use method default)
}

// GoMethod represents a resolved method for template rendering.
type GoMethod struct {
	Comment      string          // multi-line doc comment
	APIName      string          // e.g., "sendMessage"
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
	"int":     "Int",
	"int64":   "Int64",
	"float64": "Float64",
	"string":  "String",
	"bool":    "Bool",
	"PeerID":  "PeerID",
	"ChatID":  "ChatID",
	"UserID":  "UserID",
	"FileID":  "FileID",
	"FileArg": "File",
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

	return tmpl.Execute(w, data)
}

func buildTemplateData(api *ir.API, cfg *config.MethodGen, rules *CompiledParamTypeRules, stringerTypes map[string]bool, log *slog.Logger) *TemplateData {
	data := &TemplateData{}

	usedParamOverrides := make(map[string]bool)
	usedReturnOverrides := make(map[string]bool)

	for _, m := range api.Methods {
		goMethod := resolveMethod(m, cfg, rules, stringerTypes, usedParamOverrides, usedReturnOverrides)
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

func resolveMethod(m ir.Method, cfg *config.MethodGen, rules *CompiledParamTypeRules, stringerTypes map[string]bool, usedParamOverrides, usedReturnOverrides map[string]bool) GoMethod {
	goName := naming.MethodName(m.Name)
	callTypeName := goName + "Call"

	// Resolve params.
	var params []GoParam
	for _, p := range m.Params {
		goParam := resolveParam(m.Name, p, cfg, rules, stringerTypes, usedParamOverrides)
		params = append(params, goParam)
	}

	// Build constructor variants.
	constructors := buildConstructors(m.Name, params, cfg.ConstructorVariants)

	// Resolve return type - use first variant's return_type if specified.
	var embedType string
	var isNoResult bool
	if len(constructors) > 0 && constructors[0].EmbedType != "" {
		embedType = constructors[0].EmbedType
		isNoResult = embedType == "CallNoResult"
	} else {
		embedType, isNoResult = resolveReturnType(m, cfg.ReturnTypeOverrides, usedReturnOverrides)
	}

	comment := formatMethodComment(m.Description)

	return GoMethod{
		Comment:      comment,
		APIName:      m.Name,
		CallTypeName: callTypeName,
		GoName:       goName,
		EmbedType:    embedType,
		IsNoResult:   isNoResult,
		Params:       params,
		Constructors: constructors,
	}
}

// buildConstructors builds the constructor variants for a method.
func buildConstructors(methodName string, params []GoParam, variants map[string][]config.ConstructorVariant) []GoConstructor {
	// Check for explicit variants in config.
	if cfgVariants, ok := variants[methodName]; ok {
		var constructors []GoConstructor
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
			constructors = append(constructors, GoConstructor{
				Suffix:         v.Suffix,
				RequiredParams: requiredParams,
				EmbedType:      embedType,
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
	return []GoConstructor{{
		Suffix:         "",
		RequiredParams: requiredParams,
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

func resolveParam(methodName string, p ir.Param, cfg *config.MethodGen, rules *CompiledParamTypeRules, stringerTypes map[string]bool, usedOverrides map[string]bool) GoParam {
	// Resolve names.
	goName := naming.SnakeToPascal(p.Name)
	goArgName := naming.EscapeReserved(naming.SnakeToCamel(p.Name))

	// Resolve type.
	goType := resolveParamType(methodName, p, cfg, rules, usedOverrides)

	// Determine request method.
	requestMethod := resolveRequestMethod(goType, stringerTypes)

	return GoParam{
		Name:          p.Name,
		GoName:        goName,
		GoArgName:     goArgName,
		GoType:        goType,
		RequestMethod: requestMethod,
		Description:   p.Description,
		Required:      p.Required,
	}
}

func resolveParamType(methodName string, p ir.Param, cfg *config.MethodGen, rules *CompiledParamTypeRules, usedOverrides map[string]bool) string {
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

	// For unions with multiple types, default to any (unless matched by rule).
	if len(p.TypeExpr.Types) > 1 {
		return "any"
	}

	baseType := resolveBaseType(p.TypeExpr.Types[0].Type)
	return applyArray(baseType, p.TypeExpr.Array)
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

	// Special case for InputMedia slice.
	if goType == "[]InputMedia" {
		return "InputMediaSlice"
	}

	// Default to JSON for everything else.
	return "JSON"
}

func formatMethodComment(desc []string) string {
	if len(desc) == 0 {
		return ""
	}
	return strings.Join(desc, "\n// ")
}
