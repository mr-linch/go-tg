package typegen

import "github.com/mr-linch/go-tg/gen/naming"

// normalizeTypeName applies Go initialism rules to PascalCase type names.
func normalizeTypeName(name string) string {
	return naming.NormalizeTypeName(name)
}

// snakeToPascal converts a snake_case string to PascalCase.
func snakeToPascal(s string) string {
	return naming.SnakeToPascal(s)
}
