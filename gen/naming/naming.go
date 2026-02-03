package naming

import (
	"strings"
	"unicode"
)

// Initialisms is the set of Go naming convention initialisms.
var Initialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GIF":   true,
	"GUID":  true,
	"MIME":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"MPEG":  true,
	"OK":    true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSS":   true,
}

// NormalizeTypeName applies Go initialism rules to PascalCase type names.
// e.g., "MessageId" → "MessageID", "LoginUrl" → "LoginURL",
// "InlineQueryResultMpeg4Gif" → "InlineQueryResultMPEG4GIF".
func NormalizeTypeName(name string) string {
	for initialism := range Initialisms {
		if len(initialism) < 2 {
			continue
		}
		mixed := strings.ToUpper(initialism[:1]) + strings.ToLower(initialism[1:])
		name = replaceInitialism(name, mixed, initialism)
	}
	return name
}

// replaceInitialism replaces all occurrences of mixed-case initialism form
// at PascalCase word boundaries (followed by uppercase, digit, or end of string).
func replaceInitialism(name, mixed, upper string) string {
	offset := 0
	for {
		idx := strings.Index(name[offset:], mixed)
		if idx < 0 {
			return name
		}
		pos := offset + idx
		end := pos + len(mixed)
		if end == len(name) || unicode.IsUpper(rune(name[end])) || unicode.IsDigit(rune(name[end])) {
			name = name[:pos] + upper + name[end:]
			offset = pos + len(upper)
		} else {
			offset = end
		}
	}
}

// SnakeToPascal converts a snake_case string to PascalCase,
// applying Go initialism rules (e.g., "id" → "ID", "url" → "URL").
func SnakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var sb strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		upper := strings.ToUpper(part)
		switch {
		case Initialisms[upper]:
			sb.WriteString(upper)
		case len(part) > 1 && strings.HasSuffix(part, "s") && Initialisms[strings.ToUpper(part[:len(part)-1])]:
			// Plural of an initialism: "ids" → "IDs", "urls" → "URLs"
			sb.WriteString(strings.ToUpper(part[:len(part)-1]) + "s")
		default:
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			sb.WriteString(string(runes))
		}
	}
	return NormalizeTypeName(sb.String())
}

// SnakeToCamel converts a snake_case string to camelCase,
// applying Go initialism rules (e.g., "chat_id" → "chatID", "from_chat_id" → "fromChatID").
func SnakeToCamel(s string) string {
	pascal := SnakeToPascal(s)
	if pascal == "" {
		return ""
	}
	runes := []rune(pascal)

	// Find where the leading uppercase run ends.
	// For "ChatID" → "chatID", "ID" → "id", "URL" → "url", "HTTPSServer" → "httpsServer"
	i := 0
	for i < len(runes) && unicode.IsUpper(runes[i]) {
		i++
	}

	if i == 0 {
		return pascal
	}

	// If entire string is uppercase, lowercase all
	if i == len(runes) {
		return strings.ToLower(pascal)
	}

	// Lowercase all but the last uppercase (which starts the next word)
	// "ChatID" (i=1) → "chatID"
	// "HTTPSServer" (i=5) → lowercase first 4 → "httpsServer"
	for j := 0; j < i-1; j++ {
		runes[j] = unicode.ToLower(runes[j])
	}
	// Also lowercase the single leading char if i==1
	if i == 1 {
		runes[0] = unicode.ToLower(runes[0])
	}

	return string(runes)
}

// MethodName converts an API method name to a Go method name.
// It capitalizes the first letter (e.g., "sendMessage" → "SendMessage").
func MethodName(apiName string) string {
	if apiName == "" {
		return ""
	}
	runes := []rune(apiName)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// GoReservedWords is the set of Go reserved keywords and predeclared identifiers
// that cannot be used as variable names.
var GoReservedWords = map[string]string{
	// Keywords
	"break":       "break_",
	"case":        "case_",
	"chan":        "chan_",
	"const":       "const_",
	"continue":    "continue_",
	"default":     "default_",
	"defer":       "defer_",
	"else":        "else_",
	"fallthrough": "fallthrough_",
	"for":         "for_",
	"func":        "func_",
	"go":          "go_",
	"goto":        "goto_",
	"if":          "if_",
	"import":      "import_",
	"interface":   "interface_",
	"map":         "map_",
	"package":     "package_",
	"range":       "range_",
	"return":      "return_",
	"select":      "select_",
	"struct":      "struct_",
	"switch":      "switch_",
	"type":        "type_",
	"var":         "var_",
}

// EscapeReserved returns an escaped version of name if it's a Go reserved word.
func EscapeReserved(name string) string {
	if escaped, ok := GoReservedWords[name]; ok {
		return escaped
	}
	return name
}
