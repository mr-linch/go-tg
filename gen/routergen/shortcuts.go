package routergen

import (
	"fmt"
	"maps"
	"path"
	"strings"
	"unicode"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/methodgen"
	"github.com/mr-linch/go-tg/gen/naming"
)

// GoHelperArg represents a function argument in a generated helper method.
type GoHelperArg struct {
	Name     string // camelCase arg name
	Type     string // full type with tg. prefix (e.g., "tg.FileArg")
	Variadic bool   // renders as ...Type
}

// GoChainCall represents a fluent chain call appended to the helper return.
type GoChainCall struct {
	Setter string // method name (e.g., "Emoji")
	Arg    string // arg expression (e.g., "emoji")
	Spread bool   // true → "arg..."
}

// GoUpdateHelper represents one generated helper method on an update wrapper type.
type GoUpdateHelper struct {
	UpdateType    string        // e.g., "MessageUpdate"
	Receiver      string        // e.g., "msg"
	Name          string        // e.g., "AnswerPhoto"
	Comment       string        // doc comment body
	ArgsSignature string        // e.g., "photo tg.FileArg"
	ReturnType    string        // e.g., "SendPhotoCall"
	ClientMethod  string        // e.g., "SendPhoto"
	ConstructArgs string        // e.g., "msg.Chat, photo"
	ChainCalls    []GoChainCall // fluent chain calls
}

// goPrimitives are types that don't need a tg. prefix.
var goPrimitives = map[string]bool{
	"int": true, "int64": true, "float64": true,
	"string": true, "bool": true,
}

// prefixTgType adds "tg." prefix to non-primitive Go types.
func prefixTgType(goType string) string {
	// Handle slice prefix
	if strings.HasPrefix(goType, "[]") {
		inner := goType[2:]
		if goPrimitives[inner] {
			return goType
		}
		return "[]tg." + inner
	}

	if goPrimitives[goType] {
		return goType
	}
	return "tg." + goType
}

// resolveShortcuts builds helper method definitions from the shortcuts config.
func resolveShortcuts(
	methods []methodgen.GoMethod,
	cfg *config.ShortcutsConfig,
	handlerTypes []GoHandlerType,
) ([]GoUpdateHelper, error) {
	if cfg == nil || len(cfg.Methods) == 0 {
		return nil, nil
	}

	// Build method lookup by API name.
	methodByAPI := make(map[string]methodgen.GoMethod, len(methods))
	for _, m := range methods {
		methodByAPI[m.APIName] = m
	}

	// Build handler type lookup.
	handlerByWrapper := make(map[string]GoHandlerType, len(handlerTypes))
	for _, ht := range handlerTypes {
		handlerByWrapper[ht.WrapperName] = ht
	}

	// Collect all API method names for glob matching.
	apiNames := make([]string, 0, len(methods))
	for _, m := range methods {
		apiNames = append(apiNames, m.APIName)
	}

	var helpers []GoUpdateHelper

	for entryIdx, entry := range cfg.Methods {
		// Resolve matching methods.
		matched, err := matchMethods(entry, apiNames)
		if err != nil {
			return nil, fmt.Errorf("shortcuts.methods[%d]: %w", entryIdx, err)
		}

		for _, target := range entry.Targets {
			binding, ok := cfg.Bindings[target]
			if !ok {
				return nil, fmt.Errorf("shortcuts.methods[%d]: no binding defined for target %q", entryIdx, target)
			}

			for _, apiName := range matched {
				goMethod := methodByAPI[apiName]
				if helper := buildHelper(entry, target, binding, goMethod); helper != nil {
					helpers = append(helpers, *helper)
				}
			}
		}
	}

	return helpers, nil
}

// matchMethods resolves which API methods match the entry's criteria.
func matchMethods(entry config.ShortcutMethod, apiNames []string) ([]string, error) {
	excludeSet := make(map[string]bool, len(entry.Exclude))
	for _, e := range entry.Exclude {
		excludeSet[e] = true
	}

	if entry.Method != "" {
		// Exact match.
		if excludeSet[entry.Method] {
			return nil, nil
		}
		return []string{entry.Method}, nil
	}

	if entry.Match != "" {
		// Glob match.
		var result []string
		for _, name := range apiNames {
			if excludeSet[name] {
				continue
			}
			ok, err := path.Match(entry.Match, name)
			if err != nil {
				return nil, fmt.Errorf("invalid glob pattern %q: %w", entry.Match, err)
			}
			if ok {
				result = append(result, name)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("must specify either 'method' or 'match'")
}

// buildHelper constructs a single GoUpdateHelper for a (entry, target, method) triple.
// Returns nil if the method has no bindable params for this target.
func buildHelper(
	entry config.ShortcutMethod,
	target string,
	binding config.ShortcutBinding,
	goMethod methodgen.GoMethod,
) *GoUpdateHelper {
	receiver := binding.Receiver

	// Get required params from the default constructor.
	requiredParams := goMethod.RequiredParams()

	// Resolve bind map: per-shortcut overrides take priority, then target bindings.
	bindMap := make(map[string]string, len(binding.Params)+len(entry.Bind))
	maps.Copy(bindMap, binding.Params)
	maps.Copy(bindMap, entry.Bind)

	// Build constructor args and function args.
	var funcArgs []GoHelperArg
	constructParts := make([]string, 0, len(requiredParams))
	boundCount := 0

	for _, p := range requiredParams {
		arg, construct, bound := resolveParamBinding(p, bindMap)
		if arg != nil {
			funcArgs = append(funcArgs, *arg)
		}
		constructParts = append(constructParts, construct)
		if bound {
			boundCount++
		}
	}

	// Skip methods where no params were bound to context
	// (they wouldn't benefit from being on the update type).
	if boundCount == 0 && len(entry.Chain) == 0 && entry.Bind == nil {
		return nil
	}

	// Resolve chain calls.
	var chainCalls []GoChainCall
	for _, ch := range entry.Chain {
		variadic := strings.HasPrefix(ch.Type, "...")
		argType := ch.Type
		if variadic {
			argType = argType[3:]
		}

		funcArgs = append(funcArgs, GoHelperArg{
			Name:     ch.Param,
			Type:     prefixTgType(argType),
			Variadic: variadic,
		})

		chainCalls = append(chainCalls, GoChainCall{
			Setter: ch.Setter,
			Arg:    ch.Param,
			Spread: variadic,
		})
	}

	// Derive helper name.
	helperName := deriveHelperName(entry, goMethod.GoName)

	// Build doc comment.
	comment := fmt.Sprintf("%s calls [tg.Client.%s]", helperName, goMethod.GoName)

	return &GoUpdateHelper{
		UpdateType:    target,
		Receiver:      receiver,
		Name:          helperName,
		Comment:       comment,
		ArgsSignature: helperArgsSignature(funcArgs),
		ReturnType:    goMethod.CallTypeName,
		ClientMethod:  goMethod.GoName,
		ConstructArgs: strings.Join(constructParts, ", "),
		ChainCalls:    chainCalls,
	}
}

// resolveParamBinding resolves how a single param should be handled.
// Returns the function arg (nil if bound to expression), the constructor expression,
// and whether the param was bound to a context expression.
func resolveParamBinding(p methodgen.GoParam, bindMap map[string]string) (arg *GoHelperArg, construct string, bound bool) {
	expr, ok := bindMap[p.Name]
	if !ok {
		// Unbound — becomes a function arg.
		a := GoHelperArg{Name: p.GoArgName, Type: prefixTgType(p.GoType), Variadic: p.Variadic}
		construct = p.GoArgName
		if p.Variadic {
			construct += "..."
		}
		return &a, construct, false
	}

	// Check for {argName} syntax — creates a function arg with custom name.
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		argName := expr[1 : len(expr)-1]
		a := GoHelperArg{Name: argName, Type: prefixTgType(p.GoType), Variadic: p.Variadic}
		construct = argName
		if p.Variadic {
			construct += "..."
		}
		return &a, construct, false
	}

	// Bound to a Go expression.
	return nil, expr, true
}

// deriveHelperName computes the helper method name from config and the Go method name.
func deriveHelperName(entry config.ShortcutMethod, goName string) string {
	if entry.HelperName != "" {
		return entry.HelperName
	}

	name := goName

	if entry.HelperStrip != "" {
		name = strings.Replace(name, entry.HelperStrip, "", 1)
	}

	if entry.HelperPrefix != "" {
		// Strip the glob pattern's constant prefix from GoName.
		if entry.Match != "" {
			globPrefix := globConstantPrefix(entry.Match)
			goPrefix := capitalizeFirst(globPrefix)
			name = strings.TrimPrefix(name, goPrefix)
		}
		name = entry.HelperPrefix + name
	}

	return name
}

// globConstantPrefix returns the part of a glob pattern before the first wildcard.
func globConstantPrefix(pattern string) string {
	for i, r := range pattern {
		if r == '*' || r == '?' || r == '[' {
			return pattern[:i]
		}
	}
	return pattern
}

// capitalizeFirst uppercases the first rune of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// helperArgsSignature builds the function signature string for a helper's arguments.
func helperArgsSignature(args []GoHelperArg) string {
	if len(args) == 0 {
		return ""
	}

	// Try to merge consecutive args with the same type.
	var parts []string
	i := 0
	for i < len(args) {
		j := i + 1
		// Group consecutive non-variadic args with the same type.
		if !args[i].Variadic {
			for j < len(args) && !args[j].Variadic && args[j].Type == args[i].Type {
				j++
			}
		}
		// Build the name list.
		var names []string
		for k := i; k < j; k++ {
			names = append(names, naming.EscapeReserved(args[k].Name))
		}
		typeStr := args[i].Type
		if args[i].Variadic {
			typeStr = "..." + typeStr
		}
		parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		i = j
	}
	return strings.Join(parts, ", ")
}
