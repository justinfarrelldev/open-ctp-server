package rules

import (
	"fmt"
	"strings"
	"unicode"
)

// RetailCatalog describes an original CTP2 retail database text file. Unlike
// CatalogFile/Import, this is a loss-preserving raw parser layer: it accepts the
// shipped CTP2 database syntax before any server-supported runtime projection is
// attempted.
type RetailCatalog struct {
	Filename string
	Kind     string
	Prefixes []string
}

// RetailRawRecord is a source-order record parsed from a CTP2 retail database
// file. Fields preserve repeated keys, valueless flags, colon-style bit values,
// and nested blocks so later runtime compilers can source values from the retail
// catalog instead of synthetic fixtures.
type RetailRawRecord struct {
	Kind   string           `json:"kind"`
	ID     string           `json:"id"`
	Fields []RetailRawField `json:"fields"`
}

// RetailRawField is a single source-order field within a RetailRawRecord or
// nested block. Values are the tokens following the key on the same source line;
// nested Fields are present for key { ... } blocks.
type RetailRawField struct {
	Key    string           `json:"key"`
	Values []string         `json:"values,omitempty"`
	Fields []RetailRawField `json:"fields,omitempty"`
}

// ParseRetailRawRecords parses original CTP2 database text records while
// preserving syntax the bounded runtime importer intentionally rejects: valueless
// flags, repeated fields, colon bitsets, and nested structs.
func ParseRetailRawRecords(filename string, data []byte, catalog RetailCatalog) ([]RetailRawRecord, error) {
	tokens, err := lexRetailCTP2(string(data), filename)
	if err != nil {
		return nil, err
	}
	var records []RetailRawRecord
	for index := 0; index < len(tokens); {
		if index+1 < len(tokens) && tokens[index+1].value == "{" {
			id := tokens[index]
			if !hasAnyPrefix(id.value, catalog.Prefixes) {
				return nil, importError(ErrorUnsupportedMechanic, filename, id.line, id.value, "record prefix does not match retail catalog file")
			}
			fields, next, err := parseRetailFields(tokens, index+2, filename, id.value)
			if err != nil {
				return nil, err
			}
			records = append(records, RetailRawRecord{Kind: catalog.Kind, ID: id.value, Fields: fields})
			index = next
			continue
		}
		// Some legacy databases start with a standalone record count. Ignore any
		// top-level token that is not a record ID followed by an opening brace.
		index++
	}
	if len(records) == 0 {
		return nil, importError(ErrorMalformedValue, filename, 1, "", "no retail CTP2 records")
	}
	return records, nil
}

func parseRetailFields(tokens []token, index int, filename, recordID string) ([]RetailRawField, int, error) {
	var fields []RetailRawField
	for index < len(tokens) {
		current := tokens[index]
		switch current.value {
		case "}":
			return fields, index + 1, nil
		case "{", ":":
			return nil, index, importError(ErrorMalformedValue, filename, current.line, recordID, "unexpected token "+current.value)
		}

		key := current
		index++
		if index < len(tokens) && tokens[index].value == ":" {
			index++
		}
		if index < len(tokens) && tokens[index].value == "{" {
			child, next, err := parseRetailFields(tokens, index+1, filename, recordID)
			if err != nil {
				return nil, index, err
			}
			fields = append(fields, RetailRawField{Key: key.value, Fields: child})
			index = next
			continue
		}

		var values []string
		for index < len(tokens) && tokens[index].value != "}" && tokens[index].value != "{" && tokens[index].line == key.line {
			if tokens[index].value != ":" {
				values = append(values, tokens[index].value)
			}
			index++
		}
		field := RetailRawField{Key: key.value, Values: values}
		if index < len(tokens) && tokens[index].value == "{" {
			child, next, err := parseRetailFields(tokens, index+1, filename, recordID)
			if err != nil {
				return nil, index, err
			}
			field.Fields = child
			index = next
		}
		fields = append(fields, field)
	}
	return nil, index, importError(ErrorMalformedValue, filename, 1, recordID, "missing closing brace")
}

func (r RetailRawRecord) Values(key string) []string {
	return retailFieldValues(r.Fields, key)
}

func (r RetailRawRecord) FirstValue(key string) string {
	values := r.Values(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (r RetailRawRecord) HasFlag(key string) bool {
	return retailHasFlag(r.Fields, key)
}

func (r RetailRawRecord) FirstBlock(key string) (RetailRawField, bool) {
	return retailFirstBlock(r.Fields, key)
}

func (f RetailRawField) ChildValues(key string) []string {
	return retailFieldValues(f.Fields, key)
}

func (f RetailRawField) FirstValue(key string) string {
	values := f.ChildValues(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (f RetailRawField) HasFlag(key string) bool {
	return retailHasFlag(f.Fields, key)
}

func (f RetailRawField) FirstBlock(key string) (RetailRawField, bool) {
	return retailFirstBlock(f.Fields, key)
}

func retailFieldValues(fields []RetailRawField, key string) []string {
	values := make([]string, 0)
	for _, field := range fields {
		if field.Key == key && len(field.Values) > 0 {
			values = append(values, field.Values[0])
		}
	}
	return values
}

func retailHasFlag(fields []RetailRawField, key string) bool {
	for _, field := range fields {
		if field.Key == key && len(field.Values) == 0 && len(field.Fields) == 0 {
			return true
		}
	}
	return false
}

func retailFirstBlock(fields []RetailRawField, key string) (RetailRawField, bool) {
	for _, field := range fields {
		if field.Key == key && len(field.Fields) > 0 {
			return field, true
		}
	}
	return RetailRawField{}, false
}

func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func lexRetailCTP2(input, filename string) ([]token, error) {
	var tokens []token
	line := 1
	for index := 0; index < len(input); {
		char := rune(input[index])
		if char == '\n' {
			line++
			index++
			continue
		}
		if unicode.IsSpace(char) {
			index++
			continue
		}
		if char == '#' {
			for index < len(input) && input[index] != '\n' {
				index++
			}
			continue
		}
		if char == '{' || char == '}' || char == ':' {
			tokens = append(tokens, token{string(char), line})
			index++
			continue
		}
		if char == '"' {
			start := index
			index++
			for index < len(input) && input[index] != '"' {
				if input[index] == '\n' {
					line++
				}
				index++
			}
			if index == len(input) {
				return nil, importError(ErrorMalformedValue, filename, line, "", "unterminated quoted string")
			}
			tokens = append(tokens, token{input[start+1 : index], line})
			index++
			continue
		}
		if !isRetailTokenCharacter(char) {
			return nil, importError(ErrorMalformedValue, filename, line, "", fmt.Sprintf("invalid CTP2 token character %q", char))
		}
		start := index
		for index < len(input) && isRetailTokenCharacter(rune(input[index])) {
			index++
		}
		tokens = append(tokens, token{input[start:index], line})
	}
	return tokens, nil
}

func isRetailTokenCharacter(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' || char == '.' || char == '+' || char == '-' || char == '/'
}
