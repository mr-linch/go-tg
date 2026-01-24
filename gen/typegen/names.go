package typegen

import (
	"strings"
	"unicode"
)

// initialisms is the set of Go naming convention initialisms.
var initialisms = map[string]bool{
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

// normalizeTypeName applies Go initialism rules to PascalCase type names.
// e.g., "MessageId" → "MessageID", "LoginUrl" → "LoginURL",
// "InlineQueryResultMpeg4Gif" → "InlineQueryResultMPEG4GIF".
func normalizeTypeName(name string) string {
	for initialism := range initialisms {
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

// snakeToPascal converts a snake_case string to PascalCase,
// applying Go initialism rules (e.g., "id" → "ID", "url" → "URL").
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var sb strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		upper := strings.ToUpper(part)
		if initialisms[upper] {
			sb.WriteString(upper)
		} else if len(part) > 1 && strings.HasSuffix(part, "s") && initialisms[strings.ToUpper(part[:len(part)-1])] {
			// Plural of an initialism: "ids" → "IDs", "urls" → "URLs"
			sb.WriteString(strings.ToUpper(part[:len(part)-1]) + "s")
		} else {
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			sb.WriteString(string(runes))
		}
	}
	return normalizeTypeName(sb.String())
}
