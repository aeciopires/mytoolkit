// Package jsontoon implements the JSON to TOON Converter tool's pure logic:
// converting a JSON document into TOON (Token-Oriented Object Notation), a
// compact, indentation-based text format designed to reduce LLM token usage.
// Spec: https://github.com/toon-format/spec (v3.2).
package jsontoon

import (
	"bytes"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Delimiter  string `json:"delimiter"`   // "comma" (default) | "tab" | "pipe"
	IndentSize int    `json:"indent_size"` // default 2
}

// Convert converts JSON input into TOON text.
func Convert(input []byte, opts Options) (string, error) {
	delim := opts.Delimiter
	if delim == "" {
		delim = "comma"
	}
	if err := apperr.OneOf("delimiter", delim, "comma", "tab", "pipe"); err != nil {
		return "", err
	}
	indentSize := opts.IndentSize
	if indentSize <= 0 {
		indentSize = 2
	}

	if len(bytes.TrimSpace(input)) == 0 {
		return "", apperr.ErrEmptyInput
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return "", apperr.Newf(400, "INVALID_JSON", "%s", err.Error())
	}

	e := &emitter{
		delimChar:  delimiterChar(delim),
		delimSym:   delimiterSymbol(delim),
		indentSize: indentSize,
	}
	var buf strings.Builder
	e.emitRoot(&buf, v)
	return buf.String(), nil
}

// --- ordered JSON value model (order-preserving decode, same streaming
// technique as internal/tools/jsontree; duplicated rather than imported,
// since internal/tools/<name> packages may only depend on apperr) ---

type kind int

const (
	kindObject kind = iota
	kindArray
	kindString
	kindNumber
	kindBool
	kindNull
)

type value struct {
	kind    kind
	str     string // string kind: raw value; number kind: raw json.Number text
	boolean bool
	fields  []field // object kind, order preserved
	items   []value // array kind
}

type field struct {
	key string
	val value
}

func decodeValue(dec *json.Decoder) (value, error) {
	tok, err := dec.Token()
	if err != nil {
		return value{}, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			var fields []field
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return value{}, err
				}
				key, _ := keyTok.(string)
				child, err := decodeValue(dec)
				if err != nil {
					return value{}, err
				}
				fields = append(fields, field{key: key, val: child})
			}
			if _, err := dec.Token(); err != nil { // consume '}'
				return value{}, err
			}
			return value{kind: kindObject, fields: fields}, nil
		case '[':
			var items []value
			for dec.More() {
				child, err := decodeValue(dec)
				if err != nil {
					return value{}, err
				}
				items = append(items, child)
			}
			if _, err := dec.Token(); err != nil { // consume ']'
				return value{}, err
			}
			return value{kind: kindArray, items: items}, nil
		}
	case string:
		return value{kind: kindString, str: t}, nil
	case json.Number:
		return value{kind: kindNumber, str: t.String()}, nil
	case bool:
		return value{kind: kindBool, boolean: t}, nil
	case nil:
		return value{kind: kindNull}, nil
	}
	return value{kind: kindNull}, nil
}

// --- TOON emission ---

type emitter struct {
	delimChar  byte
	delimSym   string
	indentSize int
}

func delimiterChar(d string) byte {
	switch d {
	case "tab":
		return '\t'
	case "pipe":
		return '|'
	default:
		return ','
	}
}

// delimiterSymbol is the header marker per spec §9.1 for non-comma delimiters.
func delimiterSymbol(d string) string {
	switch d {
	case "tab":
		return "~"
	case "pipe":
		return "|"
	default:
		return ""
	}
}

func (e *emitter) indent(depth int) string {
	return strings.Repeat(" ", depth*e.indentSize)
}

func (e *emitter) emitRoot(buf *strings.Builder, v value) {
	switch v.kind {
	case kindObject:
		for _, f := range v.fields {
			e.emitField(buf, f.key, f.val, 0)
		}
	case kindArray:
		e.emitArrayField(buf, "", v.items, 0)
	default:
		buf.WriteString(e.formatPrimitive(v))
		buf.WriteString("\n")
	}
}

func (e *emitter) emitField(buf *strings.Builder, key string, v value, depth int) {
	ind := e.indent(depth)
	switch v.kind {
	case kindObject:
		buf.WriteString(ind)
		buf.WriteString(key)
		buf.WriteString(":\n")
		for _, f := range v.fields {
			e.emitField(buf, f.key, f.val, depth+1)
		}
	case kindArray:
		e.emitArrayField(buf, key, v.items, depth)
	default:
		buf.WriteString(ind)
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.WriteString(e.formatPrimitive(v))
		buf.WriteString("\n")
	}
}

func (e *emitter) emitArrayField(buf *strings.Builder, key string, items []value, depth int) {
	ind := e.indent(depth)
	n := len(items)

	if n == 0 {
		buf.WriteString(ind)
		buf.WriteString(key)
		buf.WriteString("[0]:\n")
		return
	}

	if fields, ok := uniformObjectFields(items); ok {
		buf.WriteString(ind)
		buf.WriteString(key)
		buf.WriteString("[")
		buf.WriteString(strconv.Itoa(n))
		buf.WriteString(e.delimSym)
		buf.WriteString("]{")
		buf.WriteString(strings.Join(fields, string(e.delimChar)))
		buf.WriteString("}:\n")
		rowIndent := e.indent(depth + 1)
		for _, item := range items {
			buf.WriteString(rowIndent)
			row := make([]string, len(fields))
			byKey := map[string]value{}
			for _, f := range item.fields {
				byKey[f.key] = f.val
			}
			for i, fk := range fields {
				row[i] = e.formatPrimitiveForArray(byKey[fk])
			}
			buf.WriteString(strings.Join(row, string(e.delimChar)))
			buf.WriteString("\n")
		}
		return
	}

	if allPrimitive(items) {
		buf.WriteString(ind)
		buf.WriteString(key)
		buf.WriteString("[")
		buf.WriteString(strconv.Itoa(n))
		buf.WriteString(e.delimSym)
		buf.WriteString("]: ")
		parts := make([]string, n)
		for i, it := range items {
			parts[i] = e.formatPrimitiveForArray(it)
		}
		buf.WriteString(strings.Join(parts, string(e.delimChar)))
		buf.WriteString("\n")
		return
	}

	// Fallback list form for mixed/non-uniform/nested content.
	buf.WriteString(ind)
	buf.WriteString(key)
	buf.WriteString("[")
	buf.WriteString(strconv.Itoa(n))
	buf.WriteString("]:\n")
	elemIndent := e.indent(depth + 1)
	for _, item := range items {
		e.emitListElement(buf, item, elemIndent)
	}
}

func (e *emitter) emitListElement(buf *strings.Builder, v value, ind string) {
	switch v.kind {
	case kindArray:
		buf.WriteString(ind)
		buf.WriteString("- [")
		buf.WriteString(strconv.Itoa(len(v.items)))
		buf.WriteString("]: ")
		parts := make([]string, len(v.items))
		for i, it := range v.items {
			parts[i] = e.formatPrimitiveForArray(it)
		}
		buf.WriteString(strings.Join(parts, string(e.delimChar)))
		buf.WriteString("\n")
	case kindObject:
		buf.WriteString(ind)
		buf.WriteString("-:\n")
		for _, f := range v.fields {
			e.emitField(buf, f.key, f.val, len(ind)/e.indentSize+1)
		}
	default:
		buf.WriteString(ind)
		buf.WriteString("- ")
		buf.WriteString(e.formatPrimitive(v))
		buf.WriteString("\n")
	}
}

// uniformObjectFields returns the shared, ordered field-key list if every
// item is an object with identical keys (same order) and primitive values.
func uniformObjectFields(items []value) ([]string, bool) {
	if len(items) == 0 || items[0].kind != kindObject {
		return nil, false
	}
	var fields []string
	for _, f := range items[0].fields {
		if !isPrimitive(f.val) {
			return nil, false
		}
		fields = append(fields, f.key)
	}
	for _, item := range items {
		if item.kind != kindObject || len(item.fields) != len(fields) {
			return nil, false
		}
		for i, f := range item.fields {
			if f.key != fields[i] || !isPrimitive(f.val) {
				return nil, false
			}
		}
	}
	return fields, true
}

func allPrimitive(items []value) bool {
	for _, it := range items {
		if !isPrimitive(it) {
			return false
		}
	}
	return true
}

func isPrimitive(v value) bool {
	return v.kind == kindString || v.kind == kindNumber || v.kind == kindBool || v.kind == kindNull
}

// --- primitive formatting ---

func (e *emitter) formatPrimitive(v value) string {
	switch v.kind {
	case kindString:
		return e.formatString(v.str, false)
	case kindNumber:
		return formatNumber(v.str)
	case kindBool:
		if v.boolean {
			return "true"
		}
		return "false"
	default:
		return "null"
	}
}

// formatPrimitiveForArray is like formatPrimitive but also quotes strings
// that contain the active delimiter (spec §7.2).
func (e *emitter) formatPrimitiveForArray(v value) string {
	if v.kind == kindString {
		return e.formatString(v.str, true)
	}
	return e.formatPrimitive(v)
}

var numericPattern = regexp.MustCompile(`^-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?$`)

func (e *emitter) formatString(s string, inArray bool) string {
	if needsQuoting(s, e.delimChar, inArray) {
		return quoteString(s)
	}
	return s
}

func needsQuoting(s string, delim byte, inArray bool) bool {
	if s == "" {
		return true
	}
	if strings.TrimSpace(s) != s {
		return true
	}
	if s == "true" || s == "false" || s == "null" {
		return true
	}
	if numericPattern.MatchString(s) {
		return true
	}
	if strings.HasPrefix(s, "-") {
		return true
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ':' || c == '"' || c == '\\' || c == '[' || c == ']' || c == '{' || c == '}' || c < 0x20 {
			return true
		}
		if inArray && c == delim {
			return true
		}
	}
	return false
}

func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// formatNumber canonicalizes a JSON number literal per spec §2: no
// unnecessary exponents in the normal range, no leading zeros, no trailing
// fractional zeros, "-0" normalized to "0".
func formatNumber(raw string) string {
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	if f == 0 {
		return "0"
	}
	abs := math.Abs(f)
	if abs >= 1e-6 && abs < 1e21 {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		return s
	}
	return strconv.FormatFloat(f, 'e', -1, 64)
}
