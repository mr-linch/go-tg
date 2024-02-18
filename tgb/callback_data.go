package tgb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// CallbackDataParser is a helper for parsing and serializing callback data.
type CallbackDataParser struct {
	delimiter          rune
	intBase            int
	floatFmt           byte
	floatPrec          int
	disableLengthCheck bool
}

const callbackDataMaxLen = 64

// CallbackDataIsTooLongError is returned when callback data length is too long.
type CallbackDataIsTooLongError struct {
	Length int
}

// Error returns a string representation of the error.
func (e *CallbackDataIsTooLongError) Error() string {
	return fmt.Sprintf("callback data length is too long: %v, max: %v", e.Length, callbackDataMaxLen)
}

// NewCallbackDataParser creates a new CallbackDataParser with default options.
type CallbackDataParserOption func(*CallbackDataParser)

// WithCallbackDataDelimiter sets a delimiter for callback data.
// Default is ':'.
func WithCallbackDataDelimiter(delimiter rune) CallbackDataParserOption {
	return func(p *CallbackDataParser) {
		p.delimiter = delimiter
	}
}

// WithCallbackDataIntBase sets a base for integer fields in callback data.
// Default is 36.
func WithCallbackDataIntBase(base int) CallbackDataParserOption {
	return func(p *CallbackDataParser) {
		p.intBase = base
	}
}

// WithCallbackDataFloatFmt sets a format for float fields in callback data.
// Default is 'f'.
func WithCallbackDataFloatFmt(fmt byte) CallbackDataParserOption {
	return func(p *CallbackDataParser) {
		p.floatFmt = fmt
	}
}

// WithCallbackDataFloatPrec sets a precision for float fields in callback data.
// Default is -1.
func WithCallbackDataFloatPrec(prec int) CallbackDataParserOption {
	return func(p *CallbackDataParser) {
		p.floatPrec = prec
	}
}

// WithCallbackDataDisableLengthCheck disables length check for callback data.
// Default is false.
func WithCallbackDataDisableLengthCheck(disable bool) CallbackDataParserOption {
	return func(p *CallbackDataParser) {
		p.disableLengthCheck = disable
	}
}

// NewCallbackDataParser creates a new CallbackDataParser with custom options.
// With no options it will use ':' as a delimiter, 36 as a base for integer fields, 'f' as a format and -1 as a precision for float fields.
func NewCallbackDataParser(opts ...CallbackDataParserOption) *CallbackDataParser {
	parser := &CallbackDataParser{
		delimiter:          ':',
		intBase:            36,
		floatFmt:           'f',
		floatPrec:          -1,
		disableLengthCheck: false,
	}

	for _, opt := range opts {
		opt(parser)
	}

	return parser
}

func (p *CallbackDataParser) getIntFieldBaseOrDefault(field reflect.StructField) (int, error) {
	baseStr, ok := field.Tag.Lookup("tgbase")
	if !ok {
		return p.intBase, nil
	}

	base, err := strconv.Atoi(baseStr)
	if err != nil {
		return 0, fmt.Errorf("invalid base value: %w", err)
	}

	return base, nil
}

func (p *CallbackDataParser) getFloatFieldFmtOrDefault(field reflect.StructField) (byte, error) {
	fmtStr, ok := field.Tag.Lookup("tgfmt")
	if !ok {
		return p.floatFmt, nil
	}

	if len(fmtStr) != 1 {
		return 0, fmt.Errorf("invalid fmt value: %v", fmtStr)
	}

	return fmtStr[0], nil
}

func (p *CallbackDataParser) getFloatFieldPrecOrDefault(field reflect.StructField) (int, error) {
	precStr, ok := field.Tag.Lookup("tgprec")
	if !ok {
		return p.floatPrec, nil
	}

	prec, err := strconv.Atoi(precStr)
	if err != nil {
		return 0, fmt.Errorf("invalid prec value: %w", err)
	}

	return prec, nil
}

// MarshalCallbackData serializes a struct into callback data.
// This data will be in format prefix:field_value_1:field_value_2:...:field_value_n
// Only plain structures are supported.
func (p *CallbackDataParser) Encode(src any) (string, error) {
	structValue := reflect.ValueOf(src)

	if structValue.Type().Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	if !structValue.IsValid() {
		return "", fmt.Errorf("src is nil")
	}

	if structValue.Kind() != reflect.Struct {
		return "", fmt.Errorf("src should be a struct")
	}

	var result strings.Builder

	fieldsCount := structValue.NumField()

	structType := structValue.Type()

	for i := 0; i < fieldsCount; i++ {
		if i > 0 {
			result.WriteRune(p.delimiter)
		}

		field := structValue.Field(i)
		structField := structType.Field(i)

		switch field.Kind() {
		case reflect.Bool:
			if field.Bool() {
				result.WriteString("1")
			} else {
				result.WriteString("0")
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			base, err := p.getIntFieldBaseOrDefault(structField)
			if err != nil {
				return "", fmt.Errorf("field %v: %w", structField.Name, err)
			}

			result.WriteString(strconv.FormatInt(field.Int(), base))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			base, err := p.getIntFieldBaseOrDefault(structField)
			if err != nil {
				return "", fmt.Errorf("field %v: %w", structField.Name, err)
			}

			result.WriteString(strconv.FormatUint(field.Uint(), base))
		case reflect.String:
			result.WriteString(field.String())
		case reflect.Float32, reflect.Float64:
			format, err := p.getFloatFieldFmtOrDefault(structField)
			if err != nil {
				return "", fmt.Errorf("field %v: %w", structField.Name, err)
			}

			prec, err := p.getFloatFieldPrecOrDefault(structField)
			if err != nil {
				return "", fmt.Errorf("field %v: %w", structField.Name, err)
			}

			result.WriteString(strconv.FormatFloat(field.Float(), format, prec, 64))
		default:
			return "", fmt.Errorf("unsupported field type: %v", field.Kind())
		}
	}

	if !p.disableLengthCheck && result.Len() > callbackDataMaxLen {
		return "", &CallbackDataIsTooLongError{Length: result.Len()}
	}

	return result.String(), nil
}

func (p *CallbackDataParser) Decode(data string, dst any) error {
	structValue := reflect.ValueOf(dst)

	if structValue.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("dst should be a pointer to a struct")
	}

	structValue = structValue.Elem()

	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dst should be a pointer to a struct")
	}

	fieldsCount := structValue.NumField()

	structType := structValue.Type()

	var values []string
	if len(data) > 0 {
		values = strings.Split(data, string(p.delimiter))
	}

	if len(values) != fieldsCount {
		return fmt.Errorf("invalid data length: expected %v, got %v", fieldsCount, len(values))
	}

	for i := 0; i < fieldsCount; i++ {
		field := structValue.Field(i)
		structField := structType.Field(i)

		switch field.Kind() {
		case reflect.Bool:
			if values[i] == "1" {
				field.SetBool(true)
			} else if values[i] == "0" {
				field.SetBool(false)
			} else {
				return fmt.Errorf("invalid bool value: %v", values[i])
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			base, err := p.getIntFieldBaseOrDefault(structField)
			if err != nil {
				return fmt.Errorf("field %v: %w", structField.Name, err)
			}

			value, err := strconv.ParseInt(values[i], base, 64)
			if err != nil {
				return fmt.Errorf("field %v: %w", structField.Name, err)
			}

			field.SetInt(value)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			base, err := p.getIntFieldBaseOrDefault(structField)
			if err != nil {
				return fmt.Errorf("field %v: %w", structField.Name, err)
			}

			value, err := strconv.ParseUint(values[i], base, 64)
			if err != nil {
				return fmt.Errorf("field %v: %w", structField.Name, err)
			}

			field.SetUint(value)
		case reflect.String:
			field.SetString(values[i])
		case reflect.Float32, reflect.Float64:
			value, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				return fmt.Errorf("field %v: %w", structField.Name, err)
			}

			field.SetFloat(value)
		default:
			return fmt.Errorf("unsupported field type: %v", field.Kind())
		}
	}

	return nil
}

var DefaultCallbackDataParser = NewCallbackDataParser()

// EncodeCallbackData serializes a struct into callback data using default parser.
func EncodeCallbackData(src any) (string, error) {
	return DefaultCallbackDataParser.Encode(src)
}

// DecodeCallbackData deserializes callback data into a struct using default parser.
func DecodeCallbackData(data string, dst any) error {
	return DefaultCallbackDataParser.Decode(data, dst)
}
