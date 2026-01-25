package routergen

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"text/template"

	"github.com/mr-linch/go-tg/gen/ir"
	"github.com/mr-linch/go-tg/gen/naming"
)

//go:embed router.go.tmpl
var routerTmpl string

//go:embed handler.go.tmpl
var handlerTmpl string

//go:embed update.go.tmpl
var updateTmpl string

// GoRouterMethod represents one router registration method.
type GoRouterMethod struct {
	FieldName   string // PascalCase field name (e.g., "Message", "EditedMessage")
	HandlerType string // e.g., "MessageHandler"
}

// GoHandlerType represents a unique handler type (grouped by underlying type).
type GoHandlerType struct {
	TypeName    string   // "Message" - underlying tg type
	HandlerName string   // "MessageHandler"
	WrapperName string   // "MessageUpdate"
	EmbedType   string   // "tg.Message"
	FieldName   string   // Field name in wrapper struct (e.g., "Message")
	Fields      []string // Update field names using this handler ["Message", "EditedMessage", ...]
	MultiField  bool     // true if len(Fields) > 1 (uses firstNotNil)
}

// TemplateData passed to templates.
type TemplateData struct {
	Package       string
	RouterMethods []GoRouterMethod // One per Update field
	HandlerTypes  []GoHandlerType  // One per unique underlying type
}

// Options controls generation behavior.
type Options struct {
	Package string
}

// Generate writes the generated router, handler, and update files.
func Generate(api *ir.API, routerW, handlerW, updateW io.Writer, log *slog.Logger, opts Options) error {
	if opts.Package == "" {
		opts.Package = "tgb"
	}

	data, err := buildTemplateData(api, opts.Package, log)
	if err != nil {
		return err
	}

	log.Info("generating tgb infrastructure",
		"router_methods", len(data.RouterMethods),
		"handler_types", len(data.HandlerTypes))

	// Generate router file.
	if err := executeTemplate("router", routerTmpl, routerW, data); err != nil {
		return fmt.Errorf("generate router: %w", err)
	}

	// Generate handler file.
	if err := executeTemplate("handler", handlerTmpl, handlerW, data); err != nil {
		return fmt.Errorf("generate handler: %w", err)
	}

	// Generate update file.
	if err := executeTemplate("update", updateTmpl, updateW, data); err != nil {
		return fmt.Errorf("generate update: %w", err)
	}

	return nil
}

func executeTemplate(name, tmplStr string, w io.Writer, data *TemplateData) error {
	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	return tmpl.Execute(w, data)
}

func buildTemplateData(api *ir.API, pkg string, log *slog.Logger) (*TemplateData, error) {
	// Find Update type.
	var updateType *ir.Type
	for i := range api.Types {
		if api.Types[i].Name == "Update" {
			updateType = &api.Types[i]
			break
		}
	}
	if updateType == nil {
		return nil, fmt.Errorf("Update type not found in API")
	}

	// Group fields by underlying type.
	typeToFields := make(map[string][]string) // type name -> field names
	fieldToType := make(map[string]string)    // field name -> type name

	for _, f := range updateType.Fields {
		if !f.Optional {
			continue // Skip non-optional fields (like update_id)
		}
		if len(f.TypeExpr.Types) == 0 {
			continue
		}
		typeName := f.TypeExpr.Types[0].Type
		fieldName := naming.SnakeToPascal(f.Name)

		typeToFields[typeName] = append(typeToFields[typeName], fieldName)
		fieldToType[fieldName] = typeName
	}

	// Build handler types (one per unique underlying type).
	// Naming rule:
	// - Multi-field types: use TYPE name (e.g., MessageHandler for Message type)
	// - Single-field types: use FIELD name (e.g., ChatBoostHandler for chat_boost field)
	var handlerTypes []GoHandlerType
	typeToHandler := make(map[string]string) // type name -> handler base name

	for typeName, fields := range typeToFields {
		multiField := len(fields) > 1
		var baseName string
		if multiField {
			// Multi-field: use type name
			baseName = typeName
		} else {
			// Single-field: use field name
			baseName = fields[0]
		}
		typeToHandler[typeName] = baseName

		handlerTypes = append(handlerTypes, GoHandlerType{
			TypeName:    typeName,
			HandlerName: baseName + "Handler",
			WrapperName: baseName + "Update",
			EmbedType:   "tg." + typeName,
			FieldName:   typeName, // Field name in struct literal is the type name
			Fields:      fields,
			MultiField:  multiField,
		})
	}

	// Build router methods (one per Update field).
	var routerMethods []GoRouterMethod
	for _, f := range updateType.Fields {
		if !f.Optional {
			continue
		}
		if len(f.TypeExpr.Types) == 0 {
			continue
		}
		fieldName := naming.SnakeToPascal(f.Name)
		typeName := f.TypeExpr.Types[0].Type
		baseName := typeToHandler[typeName]
		handlerType := baseName + "Handler"

		routerMethods = append(routerMethods, GoRouterMethod{
			FieldName:   fieldName,
			HandlerType: handlerType,
		})
	}

	return &TemplateData{
		Package:       pkg,
		RouterMethods: routerMethods,
		HandlerTypes:  handlerTypes,
	}, nil
}
