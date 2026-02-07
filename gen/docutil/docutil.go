package docutil

import (
	"regexp"
	"strings"
)

const apiBaseURL = "https://core.telegram.org/bots/api#"

var markdownLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// ConvertLinks replaces Markdown-style links [text](url) in s with Go doc comment links.
//
//   - If text matches a key in knownTypes, it becomes a Go doc link [text] (no URL definition needed).
//   - If text matches a key in knownMethods, it becomes a Go doc link [GoName] where GoName
//     is the value from the map (e.g., "Client.GetMe").
//   - Otherwise, it becomes [text] with a link target definition [text]: url appended
//     at the end of the returned string.
func ConvertLinks(s string, knownTypes map[string]bool, knownMethods map[string]string) string {
	converted, linkDefs := ExtractLinks(s, knownTypes, knownMethods)
	if len(linkDefs) == 0 {
		return converted
	}
	return converted + "\n\n" + strings.Join(linkDefs, "\n")
}

// ExtractLinks replaces Markdown-style links [text](url) in s with Go doc comment links
// and returns the converted text and collected link target definitions separately.
// See [ConvertLinks] for the resolution rules.
func ExtractLinks(s string, knownTypes map[string]bool, knownMethods map[string]string) (converted string, linkDefs []string) {
	matches := markdownLinkRe.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return s, nil
	}

	// Build anchor → Go doc link lookup from known types and methods.
	// Type anchors are lowercase type names; method anchors are the API name (already lowercase).
	anchorToDoc := make(map[string]string, len(knownTypes)+len(knownMethods))
	for t := range knownTypes {
		anchorToDoc[strings.ToLower(t)] = "[" + t + "]"
	}
	for apiName, goDoc := range knownMethods {
		anchorToDoc[strings.ToLower(apiName)] = "[" + goDoc + "]"
	}

	seen := map[string]bool{}
	var result strings.Builder
	last := 0

	for _, loc := range matches {
		// loc[0:2] = full match, loc[2:4] = text group, loc[4:6] = url group
		text := s[loc[2]:loc[3]]
		url := s[loc[4]:loc[5]]

		result.WriteString(s[last:loc[0]])

		switch {
		case knownTypes[text]:
			result.WriteString("[" + text + "]")
		case knownMethods[text] != "":
			result.WriteString("[" + knownMethods[text] + "]")
		default:
			// Try to resolve by URL anchor (e.g., #unbanchatsenderchat → [Client.UnbanChatSenderChat]).
			if anchor, ok := strings.CutPrefix(url, apiBaseURL); ok {
				if docLink, found := anchorToDoc[anchor]; found {
					result.WriteString(docLink)
					last = loc[1]
					continue
				}
			}
			result.WriteString("[" + text + "]")
			if !seen[text] {
				seen[text] = true
				linkDefs = append(linkDefs, "["+text+"]: "+url)
			}
		}

		last = loc[1]
	}
	result.WriteString(s[last:])

	return result.String(), linkDefs
}
