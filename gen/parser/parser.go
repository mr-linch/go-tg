package parser

import (
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/mr-linch/go-tg/gen/ir"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var reQuotedValue = regexp.MustCompile(`["\x{201c}]([^"\x{201d}]+)["\x{201d}]`)
var reAlwaysConst = regexp.MustCompile(`always ["\x{201c}]([^"\x{201d}]+)["\x{201d}]`)
var reDefault = regexp.MustCompile(`(?i)\bDefaults to (\w+)`)

// Parse reads Telegram Bot API HTML from r and returns the parsed IR.
func Parse(r io.Reader) (*ir.API, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	sections := extractSections(doc)

	api := &ir.API{}
	for _, sec := range sections {
		switch {
		case isTypeName(sec.name):
			t := parseTypeSection(sec)
			api.Types = append(api.Types, t)
		case isMethodName(sec.name):
			m := parseMethodSection(sec)
			api.Methods = append(api.Methods, m)
		}
	}

	return api, nil
}

// section represents a parsed section delimited by <h4>.
type section struct {
	name     string       // text content of the <h4>
	elements []*html.Node // sibling nodes between this <h4> and the next <h4>/<h3>
}

// extractSections walks the DOM, splitting content by <h4> elements.
func extractSections(doc *html.Node) []section {
	var sections []section
	var h4s []*html.Node
	findH4s(doc, &h4s)

	for _, h4 := range h4s {
		name := extractText(h4)
		var elements []*html.Node
		for sib := h4.NextSibling; sib != nil; sib = sib.NextSibling {
			if sib.Type == html.ElementNode && (sib.DataAtom == atom.H4 || sib.DataAtom == atom.H3) {
				break
			}
			if sib.Type == html.ElementNode {
				elements = append(elements, sib)
			}
		}
		sections = append(sections, section{name: name, elements: elements})
	}

	return sections
}

// findH4s recursively finds all <h4> elements.
func findH4s(n *html.Node, result *[]*html.Node) {
	if n.Type == html.ElementNode && n.DataAtom == atom.H4 {
		*result = append(*result, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findH4s(c, result)
	}
}

const baseURL = "https://core.telegram.org/bots/api"

// extractText returns the concatenated text content of an HTML node tree,
// skipping <i> elements (used for anchor icons) and converting <br> to space.
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Type == html.ElementNode {
		if n.DataAtom == atom.I {
			return ""
		}
		if n.DataAtom == atom.Br {
			return " "
		}
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(extractText(c))
	}
	return sb.String()
}

// extractDescription returns text content with links rendered as markdown.
func extractDescription(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Type == html.ElementNode {
		if n.DataAtom == atom.I {
			return ""
		}
		if n.DataAtom == atom.Br {
			return " "
		}
		if n.DataAtom == atom.A {
			href := getAttr(n, "href")
			text := extractText(n)
			if href != "" && text != "" {
				if strings.HasPrefix(href, "#") {
					href = baseURL + href
				}
				return "[" + text + "](" + href + ")"
			}
			return text
		}
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(extractDescription(c))
	}
	return sb.String()
}

// extractDescriptionLines returns description text split at <br> elements.
// For blockquotes, it processes inner <p> elements.
func extractDescriptionLines(n *html.Node) []string {
	if n.Type == html.ElementNode && n.DataAtom == atom.Blockquote {
		var lines []string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.DataAtom == atom.P {
				lines = append(lines, splitByBr(c)...)
			}
		}
		return lines
	}
	return splitByBr(n)
}

// splitByBr splits the direct children of a node at <br> elements,
// returning each segment as a separate trimmed string.
func splitByBr(n *html.Node) []string {
	var lines []string
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == atom.Br {
			if line := strings.TrimSpace(sb.String()); line != "" {
				lines = append(lines, line)
			}
			sb.Reset()
			continue
		}
		sb.WriteString(extractDescription(c))
	}
	if line := strings.TrimSpace(sb.String()); line != "" {
		lines = append(lines, line)
	}
	return lines
}

// isTypeName returns true if name matches ^[A-Z][a-zA-Z0-9]+$
func isTypeName(name string) bool {
	if len(name) < 2 {
		return false
	}
	if !unicode.IsUpper(rune(name[0])) {
		return false
	}
	for _, r := range name[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// isMethodName returns true if name matches ^[a-z][a-zA-Z]+$
// (all letters, starts with lowercase, at least 2 chars).
func isMethodName(name string) bool {
	if len(name) < 2 {
		return false
	}
	if !unicode.IsLower(rune(name[0])) {
		return false
	}
	for _, r := range name[1:] {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// parseTypeSection parses a section classified as a type.
func parseTypeSection(sec section) ir.Type {
	t := ir.Type{Name: sec.name}

	// Extract description from <p> elements
	var descParts []string
	for _, el := range sec.elements {
		if el.DataAtom == atom.P {
			if text := strings.TrimSpace(extractDescription(el)); text != "" {
				descParts = append(descParts, text)
			}
		}
	}
	t.Description = strings.Join(descParts, "\n")

	// Look for <table>
	table := findElement(sec.elements, atom.Table)
	if table != nil {
		t.Fields = parseFieldTable(table)
		return t
	}

	// Look for <ul> with subtypes
	for _, el := range sec.elements {
		if el.DataAtom == atom.Ul {
			if subtypes, ok := isSubtypesList(el); ok {
				t.Subtypes = subtypes
				return t
			}
		}
	}

	return t
}

// parseFieldTable parses a type's field table.
func parseFieldTable(table *html.Node) []ir.Field {
	var fields []ir.Field
	tbody := findChild(table, atom.Tbody)
	if tbody == nil {
		return nil
	}

	for tr := tbody.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type != html.ElementNode || tr.DataAtom != atom.Tr {
			continue
		}
		tds := collectChildren(tr, atom.Td)
		if len(tds) < 3 {
			continue
		}

		name := strings.TrimSpace(extractText(tds[0]))
		name = strings.TrimSuffix(name, " NEW")

		typeExpr := parseTypeCell(tds[1])
		desc := strings.TrimSpace(extractDescription(tds[2]))
		optional := strings.Contains(desc, "Optional")

		// Check for Integer64 (52-bit IDs)
		if len(typeExpr.Types) == 1 && typeExpr.Types[0].Type == "Integer" &&
			isInt64Description(desc) {
			typeExpr.Types[0].Type = string(ir.TypeInteger64)
		}

		// Detect const/enum values
		constVal := extractConst(tds[2])
		var enumVals []string
		if constVal == "" {
			enumVals = extractEnum(desc)
		}

		fields = append(fields, ir.Field{
			Name:        name,
			TypeExpr:    typeExpr,
			Optional:    optional,
			Description: desc,
			Const:       constVal,
			Enum:        enumVals,
		})
	}

	return fields
}

// parseMethodSection parses a section classified as a method.
func parseMethodSection(sec section) ir.Method {
	m := ir.Method{Name: sec.name}

	// Collect description paragraphs and blockquotes, splitting at <br>
	var descNodes []*html.Node
	for _, el := range sec.elements {
		if el.DataAtom == atom.P || el.DataAtom == atom.Blockquote {
			m.Description = append(m.Description, extractDescriptionLines(el)...)
			descNodes = append(descNodes, el)
		}
	}

	// Parse return type from description HTML
	m.Returns = parseReturnType(descNodes)

	// Look for parameter table
	table := findElement(sec.elements, atom.Table)
	if table != nil {
		m.Params = parseParamTable(table)
	}

	return m
}

// parseParamTable parses a method's parameter table.
func parseParamTable(table *html.Node) []ir.Param {
	var params []ir.Param
	tbody := findChild(table, atom.Tbody)
	if tbody == nil {
		return nil
	}

	for tr := tbody.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type != html.ElementNode || tr.DataAtom != atom.Tr {
			continue
		}
		tds := collectChildren(tr, atom.Td)
		if len(tds) < 4 {
			continue
		}

		name := strings.TrimSpace(extractText(tds[0]))
		typeExpr := parseTypeCell(tds[1])
		required := strings.TrimSpace(extractText(tds[2])) == "Yes"
		desc := strings.TrimSpace(extractDescription(tds[3]))

		// Check for Integer64 (52-bit IDs)
		if len(typeExpr.Types) == 1 && typeExpr.Types[0].Type == "Integer" &&
			isInt64Description(desc) {
			typeExpr.Types[0].Type = string(ir.TypeInteger64)
		}

		params = append(params, ir.Param{
			Name:        name,
			TypeExpr:    typeExpr,
			Required:    required,
			Description: desc,
			Default:     extractDefault(desc),
		})
	}

	return params
}

// parseTypeCell parses a <td> containing a type expression into a TypeExpr.
func parseTypeCell(td *html.Node) ir.TypeExpr {
	text := extractText(td)
	text = strings.TrimSpace(text)

	// Count and strip "Array of " prefixes
	arrayDepth := 0
	for strings.HasPrefix(text, "Array of ") {
		arrayDepth++
		text = strings.TrimPrefix(text, "Array of ")
	}

	// Collect all <a> links in the cell
	var links []ir.TypeRef
	collectLinks(td, &links)

	var types []ir.TypeRef

	if len(links) > 0 {
		if len(links) == 1 {
			// Single link — check if there's a trailing " or Primitive"
			remaining := text
			linkText := links[0].Type
			if idx := strings.Index(remaining, linkText); idx >= 0 {
				remaining = remaining[idx+len(linkText):]
			}
			if strings.HasPrefix(remaining, " or ") {
				// e.g. "InputFile or String"
				types = append(types, links[0])
				rest := strings.TrimSpace(strings.TrimPrefix(remaining, " or "))
				types = append(types, ir.TypeRef{Type: normalizeTypeName(rest)})
			} else {
				types = append(types, ir.TypeRef{
					Type: normalizeTypeName(links[0].Type),
					Ref:  links[0].Ref,
				})
			}
		} else {
			// Multiple links → one TypeRef per link
			for _, l := range links {
				types = append(types, ir.TypeRef{
					Type: normalizeTypeName(l.Type),
					Ref:  l.Ref,
				})
			}
		}
	} else {
		// No links — split by " or " for primitive unions
		parts := strings.Split(text, " or ")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				types = append(types, ir.TypeRef{Type: normalizeTypeName(p)})
			}
		}
	}

	return ir.TypeExpr{Types: types, Array: arrayDepth}
}

// parseReturnType extracts the return TypeExpr from method description nodes.
func parseReturnType(nodes []*html.Node) ir.TypeExpr {
	// Gather all text to check patterns
	var fullText string
	for _, n := range nodes {
		fullText += extractText(n) + " "
	}

	// Check for "Returns True on success" or "True is returned" or "returns True"
	lower := strings.ToLower(fullText)
	if strings.Contains(lower, "returns true") || strings.Contains(lower, "true is returned") {
		return ir.TypeExpr{Types: []ir.TypeRef{{Type: "True"}}}
	}

	// Collect type links (PascalCase text with href starting with #)
	var typeLinks []ir.TypeRef
	for _, n := range nodes {
		collectTypeLinks(n, &typeLinks)
	}

	// Collect <em> with PascalCase text as scalars
	var emTypes []ir.TypeRef
	for _, n := range nodes {
		collectEmTypes(n, &emTypes)
	}

	if len(typeLinks) > 0 {
		// Use the last type link
		last := typeLinks[len(typeLinks)-1]

		// Check for "Array of" before this type in the text
		arrayDepth := countArrayPrefix(fullText, last.Type)

		return ir.TypeExpr{
			Types: []ir.TypeRef{last},
			Array: arrayDepth,
		}
	}

	if len(emTypes) > 0 {
		// Use the last scalar
		return ir.TypeExpr{Types: []ir.TypeRef{emTypes[len(emTypes)-1]}}
	}

	return ir.TypeExpr{}
}

// countArrayPrefix counts "Array of" occurrences before the type name in text.
func countArrayPrefix(text, typeName string) int {
	idx := strings.LastIndex(text, typeName)
	if idx < 0 {
		return 0
	}
	prefix := text[:idx]
	count := 0
	for strings.HasSuffix(strings.TrimSpace(prefix), "of") {
		trimmed := strings.TrimSpace(prefix)
		trimmed = strings.TrimSuffix(trimmed, "of")
		trimmed = strings.TrimSpace(trimmed)
		if strings.HasSuffix(trimmed, "Array") || strings.HasSuffix(strings.ToLower(trimmed), "array") {
			count++
			prefix = strings.TrimSuffix(trimmed, "Array")
			if prefix == trimmed {
				prefix = strings.TrimSuffix(trimmed, "array")
			}
		} else {
			break
		}
	}
	return count
}

// collectTypeLinks finds all <a> elements with PascalCase text and internal refs.
func collectTypeLinks(n *html.Node, result *[]ir.TypeRef) {
	if n.Type == html.ElementNode && n.DataAtom == atom.A {
		href := getAttr(n, "href")
		text := extractText(n)
		if strings.HasPrefix(href, "#") && isTypeName(text) {
			*result = append(*result, ir.TypeRef{
				Type: text,
				Ref:  strings.TrimPrefix(href, "#"),
			})
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectTypeLinks(c, result)
	}
}

// collectEmTypes finds all <em> elements with PascalCase text.
func collectEmTypes(n *html.Node, result *[]ir.TypeRef) {
	if n.Type == html.ElementNode && n.DataAtom == atom.Em {
		text := strings.TrimSpace(extractText(n))
		if isTypeName(text) {
			*result = append(*result, ir.TypeRef{Type: text})
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectEmTypes(c, result)
	}
}

// isSubtypesList checks if a <ul> contains only <li> with single <a> type links.
func isSubtypesList(ul *html.Node) ([]string, bool) {
	var names []string
	liCount := 0
	for li := ul.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.DataAtom != atom.Li {
			continue
		}
		liCount++
		// Check that the <li> contains exactly one <a> child with a type name
		a := findSingleLink(li)
		if a == nil {
			return nil, false
		}
		text := extractText(a)
		if !isTypeName(text) {
			return nil, false
		}
		href := getAttr(a, "href")
		if !strings.HasPrefix(href, "#") {
			return nil, false
		}
		names = append(names, text)
	}
	if liCount == 0 {
		return nil, false
	}
	return names, true
}

// findSingleLink finds a single <a> element that is the sole meaningful content of a node.
func findSingleLink(n *html.Node) *html.Node {
	var links []*html.Node
	var hasOtherContent bool
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == atom.A {
			links = append(links, c)
		} else if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
			hasOtherContent = true
		} else if c.Type == html.ElementNode {
			hasOtherContent = true
		}
	}
	if len(links) == 1 && !hasOtherContent {
		return links[0]
	}
	return nil
}

// isInt64Description checks if description contains the 52-bit integer note.
func isInt64Description(desc string) bool {
	return strings.Contains(desc, "52 significant bits")
}

// isTimestampDescription checks if description indicates a Unix timestamp field.
func isTimestampDescription(desc string) bool {
	lower := strings.ToLower(desc)
	return strings.Contains(lower, "unix time") || strings.Contains(lower, "unix timestamp")
}

// extractDefault extracts a default value from description text.
// Detects "Defaults to X" pattern and strips trailing punctuation.
func extractDefault(desc string) string {
	m := reDefault.FindStringSubmatch(desc)
	if m == nil {
		return ""
	}
	return m[1]
}

// extractConst detects a discriminator constant from a description <td> node.
// Detects: always "value" and must be <em>value</em>.
func extractConst(td *html.Node) string {
	// Pattern A: always "value" in text content
	text := extractText(td)
	if m := reAlwaysConst.FindStringSubmatch(text); m != nil {
		return m[1]
	}

	// Pattern B: must be <em>value</em>
	// Check that <em> is preceded by text ending with "must be "
	for c := td.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == atom.Em {
			if prev := c.PrevSibling; prev != nil && prev.Type == html.TextNode {
				if strings.HasSuffix(strings.TrimRight(prev.Data, " "), "must be") {
					val := strings.TrimSpace(extractText(c))
					if val != "" {
						return val
					}
				}
			}
		}
	}

	return ""
}

// extractEnum detects enum values from description text.
// Triggers on "one of" or "can be" patterns, extracts all quoted strings.
func extractEnum(desc string) []string {
	lower := strings.ToLower(desc)
	if !strings.Contains(lower, "one of") && !strings.Contains(lower, "can be") {
		return nil
	}

	matches := reQuotedValue.FindAllStringSubmatch(desc, -1)
	if len(matches) < 2 {
		return nil
	}

	values := make([]string, len(matches))
	for i, m := range matches {
		values[i] = m[1]
	}
	return values
}

// collectLinks recursively collects <a> elements with internal refs from a node.
func collectLinks(n *html.Node, result *[]ir.TypeRef) {
	if n.Type == html.ElementNode && n.DataAtom == atom.A {
		href := getAttr(n, "href")
		text := strings.TrimSpace(extractText(n))
		if strings.HasPrefix(href, "#") && text != "" {
			*result = append(*result, ir.TypeRef{
				Type: text,
				Ref:  strings.TrimPrefix(href, "#"),
			})
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectLinks(c, result)
	}
}

// normalizeTypeName normalizes type names ("Float number" → "Float").
func normalizeTypeName(name string) string {
	if name == "Float number" {
		return string(ir.TypeFloat)
	}
	return name
}

// findElement finds the first element with the given atom in a list.
func findElement(elements []*html.Node, a atom.Atom) *html.Node {
	for _, el := range elements {
		if el.DataAtom == a {
			return el
		}
	}
	return nil
}

// findChild finds the first child element with the given atom.
func findChild(n *html.Node, a atom.Atom) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == a {
			return c
		}
	}
	return nil
}

// collectChildren collects all child elements with the given atom.
func collectChildren(n *html.Node, a atom.Atom) []*html.Node {
	var result []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.DataAtom == a {
			result = append(result, c)
		}
	}
	return result
}

// getAttr returns the value of the named attribute.
func getAttr(n *html.Node, name string) string {
	for _, attr := range n.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}
