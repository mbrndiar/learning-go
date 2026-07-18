// Package domain defines the comparative capstone's storage-independent API.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrNotImplemented marks the incomplete boundary used by the starter harness.
// The complete solution retains the symbol for API parity but never returns it.
var ErrNotImplemented = errors.New("comparative kvstore: not implemented")

// Implemented reports whether the harness placeholders have been replaced.
const Implemented = true

// Revision is a global successful-mutation sequence number.
type Revision int64

// MaxRevision is the largest revision permitted by the shared specification.
const MaxRevision Revision = 9007199254740991

const (
	maxValueBytes     = 65536
	maxContainerDepth = 32
)

// ExpectationKind identifies a conditional mutation rule.
type ExpectationKind string

const (
	ExpectAny    ExpectationKind = "any"
	ExpectAbsent ExpectationKind = "absent"
	ExpectExact  ExpectationKind = "exact"
)

// Expectation is the parsed condition attached to a set or delete.
type Expectation struct {
	Kind     ExpectationKind
	Revision Revision
}

// Value is one normalized restricted-JSON value.
type Value = any

// Entry is the observable representation of a stored key.
type Entry struct {
	Key      string   `json:"key"`
	Value    Value    `json:"value"`
	Revision Revision `json:"revision"`
}

// SetResult is returned after a successful set.
type SetResult struct {
	Key      string   `json:"key"`
	Value    Value    `json:"value"`
	Revision Revision `json:"revision"`
	Created  bool     `json:"created"`
}

// DeleteResult is returned after a successful delete.
type DeleteResult struct {
	Key             string   `json:"key"`
	DeletedRevision Revision `json:"deleted_revision"`
	Revision        Revision `json:"revision"`
}

// ListResult is returned by list.
type ListResult struct {
	Entries        []Entry  `json:"entries"`
	GlobalRevision Revision `json:"global_revision"`
}

// Error is a shared-contract error with an optional wrapped cause.
type Error struct {
	Category string         `json:"category"`
	Details  map[string]any `json:"details"`
	Cause    error          `json:"-"`
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Category
	}
	return fmt.Sprintf("%s: %v", e.Category, e.Cause)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// ExitCode returns the normative process exit code for the error.
func (e *Error) ExitCode() int {
	switch e.Category {
	case "usage", "invalid_argument", "invalid_json", "invalid_value":
		return 2
	case "conflict":
		return 3
	case "not_found":
		return 4
	default:
		return 5
	}
}

// ParseKey validates and returns a shared-contract key.
func ParseKey(value string) (string, error) {
	if len(value) < 1 || len(value) > 128 {
		return "", invalidArgument("key", "format")
	}
	for index, character := range []byte(value) {
		if index == 0 {
			if !isASCIIAlphanumeric(character) {
				return "", invalidArgument("key", "format")
			}
			continue
		}
		if !isASCIIAlphanumeric(character) &&
			character != '.' &&
			character != '_' &&
			character != '/' &&
			character != '-' {
			return "", invalidArgument("key", "format")
		}
	}
	return value, nil
}

// ParseExpectation parses a set or delete expectation.
func ParseExpectation(value string, allowAbsent bool) (Expectation, error) {
	switch value {
	case "any":
		return Expectation{Kind: ExpectAny}, nil
	case "absent":
		if allowAbsent {
			return Expectation{Kind: ExpectAbsent}, nil
		}
		return Expectation{}, invalidArgument("expect", "format")
	}
	if value == "" {
		return Expectation{}, invalidArgument("expect", "format")
	}
	if value[0] == '0' {
		return Expectation{}, invalidArgument("expect", "format")
	}
	for _, character := range []byte(value) {
		if character < '0' || character > '9' {
			return Expectation{}, invalidArgument("expect", "format")
		}
	}
	revision, err := strconv.ParseInt(value, 10, 64)
	if err != nil || revision < 1 || revision > int64(MaxRevision) {
		return Expectation{}, invalidArgument("expect", "format")
	}
	return Expectation{Kind: ExpectExact, Revision: Revision(revision)}, nil
}

// ParseValue parses and normalizes one restricted JSON value.
func ParseValue(input json.RawMessage) (Value, error) {
	return parseValue(input, false)
}

// ParseStoredValue parses a value already persisted in normalized form.
func ParseStoredValue(input string) (Value, error) {
	return parseValue(json.RawMessage(input), true)
}

func parseValue(input json.RawMessage, requireNormalized bool) (Value, error) {
	if len(input) > maxValueBytes {
		return nil, invalidValue("byte_limit")
	}
	if !utf8.Valid(input) {
		return nil, invalidJSON()
	}

	parser := jsonParser{input: string(input), bytes: input}
	raw, err := parser.parseValue()
	if err != nil {
		return nil, err
	}
	parser.skipWhitespace()
	if parser.position != len(parser.bytes) {
		return nil, invalidJSON()
	}

	metadata := validationMetadata{normalized: !parser.sawWhitespace}
	value, err := normalizeValue(raw, &metadata, 0)
	if err != nil {
		return nil, err
	}
	if requireNormalized && !metadata.normalized {
		return nil, invalidValue("not_normalized")
	}
	return value, nil
}

func isASCIIAlphanumeric(value byte) bool {
	return value >= 'A' && value <= 'Z' ||
		value >= 'a' && value <= 'z' ||
		value >= '0' && value <= '9'
}

func invalidArgument(field, reason string) error {
	return &Error{
		Category: "invalid_argument",
		Details:  map[string]any{"field": field, "reason": reason},
	}
}

func invalidJSON() error {
	return &Error{
		Category: "invalid_json",
		Details:  map[string]any{"reason": "syntax"},
	}
}

func invalidValue(reason string) error {
	return &Error{
		Category: "invalid_value",
		Details:  map[string]any{"reason": reason},
	}
}

type rawKind uint8

const (
	rawNull rawKind = iota
	rawBool
	rawStringValue
	rawNumber
	rawArray
	rawObject
)

type rawValue struct {
	kind      rawKind
	boolean   bool
	text      string
	surrogate bool
	array     []rawValue
	object    []rawMember
}

type rawMember struct {
	name  rawString
	value rawValue
}

type rawString struct {
	value     string
	surrogate bool
}

type jsonParser struct {
	input         string
	bytes         []byte
	position      int
	sawWhitespace bool
}

func (p *jsonParser) parseValue() (rawValue, error) {
	p.skipWhitespace()
	switch p.peek() {
	case 'n':
		if !p.consumeLiteral("null") {
			return rawValue{}, invalidJSON()
		}
		return rawValue{kind: rawNull}, nil
	case 't':
		if !p.consumeLiteral("true") {
			return rawValue{}, invalidJSON()
		}
		return rawValue{kind: rawBool, boolean: true}, nil
	case 'f':
		if !p.consumeLiteral("false") {
			return rawValue{}, invalidJSON()
		}
		return rawValue{kind: rawBool}, nil
	case '"':
		value, err := p.parseString()
		if err != nil {
			return rawValue{}, err
		}
		return rawValue{
			kind:      rawStringValue,
			text:      value.value,
			surrogate: value.surrogate,
		}, nil
	case '[':
		return p.parseArray()
	case '{':
		return p.parseObject()
	case '-':
		return p.parseNumber()
	default:
		if next := p.peek(); next >= '0' && next <= '9' {
			return p.parseNumber()
		}
		return rawValue{}, invalidJSON()
	}
}

func (p *jsonParser) parseArray() (rawValue, error) {
	p.position++
	p.skipWhitespace()
	values := make([]rawValue, 0)
	if p.consumeIf(']') {
		return rawValue{kind: rawArray, array: values}, nil
	}
	for {
		value, err := p.parseValue()
		if err != nil {
			return rawValue{}, err
		}
		values = append(values, value)
		p.skipWhitespace()
		if p.consumeIf(']') {
			return rawValue{kind: rawArray, array: values}, nil
		}
		if !p.consumeIf(',') {
			return rawValue{}, invalidJSON()
		}
	}
}

func (p *jsonParser) parseObject() (rawValue, error) {
	p.position++
	p.skipWhitespace()
	members := make([]rawMember, 0)
	if p.consumeIf('}') {
		return rawValue{kind: rawObject, object: members}, nil
	}
	for {
		p.skipWhitespace()
		if p.peek() != '"' {
			return rawValue{}, invalidJSON()
		}
		name, err := p.parseString()
		if err != nil {
			return rawValue{}, err
		}
		p.skipWhitespace()
		if !p.consumeIf(':') {
			return rawValue{}, invalidJSON()
		}
		value, err := p.parseValue()
		if err != nil {
			return rawValue{}, err
		}
		members = append(members, rawMember{name: name, value: value})
		p.skipWhitespace()
		if p.consumeIf('}') {
			return rawValue{kind: rawObject, object: members}, nil
		}
		if !p.consumeIf(',') {
			return rawValue{}, invalidJSON()
		}
	}
}

func (p *jsonParser) parseString() (rawString, error) {
	p.position++
	var output strings.Builder
	segmentStart := p.position
	unpaired := false

	for p.position < len(p.bytes) {
		switch character := p.bytes[p.position]; {
		case character == '"':
			output.WriteString(p.input[segmentStart:p.position])
			p.position++
			return rawString{value: output.String(), surrogate: unpaired}, nil
		case character == '\\':
			output.WriteString(p.input[segmentStart:p.position])
			p.position++
			if p.position >= len(p.bytes) {
				return rawString{}, invalidJSON()
			}
			escaped := p.bytes[p.position]
			p.position++
			switch escaped {
			case '"', '\\', '/':
				output.WriteByte(escaped)
			case 'b':
				output.WriteByte('\b')
			case 'f':
				output.WriteByte('\f')
			case 'n':
				output.WriteByte('\n')
			case 'r':
				output.WriteByte('\r')
			case 't':
				output.WriteByte('\t')
			case 'u':
				if err := p.parseUnicodeEscape(&output, &unpaired); err != nil {
					return rawString{}, err
				}
			default:
				return rawString{}, invalidJSON()
			}
			segmentStart = p.position
		case character < 0x20:
			return rawString{}, invalidJSON()
		case character < utf8.RuneSelf:
			p.position++
		default:
			_, size := utf8.DecodeRune(p.bytes[p.position:])
			if size == 1 {
				return rawString{}, invalidJSON()
			}
			p.position += size
		}
	}
	return rawString{}, invalidJSON()
}

func (p *jsonParser) parseUnicodeEscape(output *strings.Builder, unpaired *bool) error {
	first, err := p.parseHexQuad()
	if err != nil {
		return err
	}
	switch {
	case first >= 0xd800 && first <= 0xdbff:
		if p.position+2 > len(p.bytes) ||
			p.bytes[p.position] != '\\' ||
			p.bytes[p.position+1] != 'u' {
			*unpaired = true
			output.WriteRune(utf8.RuneError)
			return nil
		}
		p.position += 2
		second, err := p.parseHexQuad()
		if err != nil {
			return err
		}
		if second < 0xdc00 || second > 0xdfff {
			*unpaired = true
			output.WriteRune(utf8.RuneError)
			if second < 0xd800 || second > 0xdfff {
				output.WriteRune(rune(second))
			}
			return nil
		}
		scalar := rune(0x10000 + (uint32(first)-0xd800)<<10 + (uint32(second) - 0xdc00))
		output.WriteRune(scalar)
	case first >= 0xdc00 && first <= 0xdfff:
		*unpaired = true
		output.WriteRune(utf8.RuneError)
	default:
		output.WriteRune(rune(first))
	}
	return nil
}

func (p *jsonParser) parseHexQuad() (uint16, error) {
	if p.position+4 > len(p.bytes) {
		return 0, invalidJSON()
	}
	var value uint16
	for _, character := range p.bytes[p.position : p.position+4] {
		var digit uint16
		switch {
		case character >= '0' && character <= '9':
			digit = uint16(character - '0')
		case character >= 'a' && character <= 'f':
			digit = uint16(character-'a') + 10
		case character >= 'A' && character <= 'F':
			digit = uint16(character-'A') + 10
		default:
			return 0, invalidJSON()
		}
		value = value*16 + digit
	}
	p.position += 4
	return value, nil
}

func (p *jsonParser) parseNumber() (rawValue, error) {
	start := p.position
	p.consumeIf('-')
	switch next := p.peek(); {
	case next == '0':
		p.position++
		if following := p.peek(); following >= '0' && following <= '9' {
			return rawValue{}, invalidJSON()
		}
	case next >= '1' && next <= '9':
		p.position++
		for next := p.peek(); next >= '0' && next <= '9'; next = p.peek() {
			p.position++
		}
	default:
		return rawValue{}, invalidJSON()
	}
	if p.consumeIf('.') {
		if next := p.peek(); next < '0' || next > '9' {
			return rawValue{}, invalidJSON()
		}
		for next := p.peek(); next >= '0' && next <= '9'; next = p.peek() {
			p.position++
		}
	}
	if next := p.peek(); next == 'e' || next == 'E' {
		p.position++
		if next = p.peek(); next == '+' || next == '-' {
			p.position++
		}
		if next = p.peek(); next < '0' || next > '9' {
			return rawValue{}, invalidJSON()
		}
		for next = p.peek(); next >= '0' && next <= '9'; next = p.peek() {
			p.position++
		}
	}
	return rawValue{kind: rawNumber, text: p.input[start:p.position]}, nil
}

func (p *jsonParser) skipWhitespace() {
	for p.position < len(p.bytes) {
		switch p.bytes[p.position] {
		case ' ', '\n', '\r', '\t':
			p.sawWhitespace = true
			p.position++
		default:
			return
		}
	}
}

func (p *jsonParser) consumeLiteral(literal string) bool {
	if !strings.HasPrefix(p.input[p.position:], literal) {
		return false
	}
	p.position += len(literal)
	return true
}

func (p *jsonParser) consumeIf(expected byte) bool {
	if p.peek() != expected {
		return false
	}
	p.position++
	return true
}

func (p *jsonParser) peek() byte {
	if p.position >= len(p.bytes) {
		return 0
	}
	return p.bytes[p.position]
}

type validationMetadata struct {
	normalized bool
}

func normalizeValue(raw rawValue, metadata *validationMetadata, depth int) (Value, error) {
	switch raw.kind {
	case rawNull:
		return nil, nil
	case rawBool:
		return raw.boolean, nil
	case rawStringValue:
		if raw.surrogate {
			return nil, invalidValue("unpaired_surrogate")
		}
		return raw.text, nil
	case rawNumber:
		number, err := normalizeNumber(raw.text)
		if err != nil {
			return nil, err
		}
		if raw.text != strconv.FormatInt(number, 10) {
			metadata.normalized = false
		}
		return number, nil
	case rawArray:
		nextDepth, err := checkedDepth(depth)
		if err != nil {
			return nil, err
		}
		values := make([]Value, 0, len(raw.array))
		for _, item := range raw.array {
			value, err := normalizeValue(item, metadata, nextDepth)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	case rawObject:
		nextDepth, err := checkedDepth(depth)
		if err != nil {
			return nil, err
		}
		lastIndices := make(map[string]int, len(raw.object))
		for index, member := range raw.object {
			if _, exists := lastIndices[member.name.value]; exists {
				metadata.normalized = false
			}
			lastIndices[member.name.value] = index
		}
		value := make(map[string]Value, len(lastIndices))
		for index, member := range raw.object {
			if lastIndices[member.name.value] != index {
				continue
			}
			if member.name.surrogate {
				return nil, invalidValue("unpaired_surrogate")
			}
			normalized, err := normalizeValue(member.value, metadata, nextDepth)
			if err != nil {
				return nil, err
			}
			value[member.name.value] = normalized
		}
		return value, nil
	default:
		panic("unreachable raw JSON kind")
	}
}

func checkedDepth(depth int) (int, error) {
	depth++
	if depth > maxContainerDepth {
		return 0, invalidValue("depth_limit")
	}
	return depth, nil
}

func normalizeNumber(token string) (int64, error) {
	binary64, _ := strconv.ParseFloat(token, 64)
	if math.IsInf(binary64, 0) {
		return 0, invalidValue("non_finite_number")
	}

	unsigned := strings.TrimPrefix(token, "-")
	mantissa := unsigned
	exponentText := "0"
	if index := strings.IndexAny(unsigned, "eE"); index >= 0 {
		mantissa = unsigned[:index]
		exponentText = unsigned[index+1:]
	}
	exponent := saturatingExponent(exponentText)
	integerPart := mantissa
	fraction := ""
	if index := strings.IndexByte(mantissa, '.'); index >= 0 {
		integerPart = mantissa[:index]
		fraction = mantissa[index+1:]
	}
	digits := integerPart + fraction
	if strings.Trim(digits, "0") == "" {
		return 0, nil
	}

	scale := saturatingSubtract(int64(len(fraction)), exponent)
	var integerDigits string
	if scale <= 0 {
		zeroCount := saturatingNegate(scale)
		significant := strings.TrimLeft(digits, "0")
		if int64(len(significant))+zeroCount > 16 {
			return 0, invalidValue("number_out_of_range")
		}
		integerDigits = digits + strings.Repeat("0", int(zeroCount))
	} else {
		if scale > int64(len(digits)) {
			return 0, invalidValue("non_integral_number")
		}
		split := len(digits) - int(scale)
		if strings.Trim(digits[split:], "0") != "" {
			return 0, invalidValue("non_integral_number")
		}
		integerDigits = digits[:split]
	}

	magnitudeText := strings.TrimLeft(integerDigits, "0")
	if magnitudeText == "" {
		return 0, nil
	}
	maximum := strconv.FormatInt(int64(MaxRevision), 10)
	if len(magnitudeText) > len(maximum) ||
		len(magnitudeText) == len(maximum) && magnitudeText > maximum {
		return 0, invalidValue("number_out_of_range")
	}
	magnitude, err := strconv.ParseInt(magnitudeText, 10, 64)
	if err != nil {
		return 0, invalidValue("number_out_of_range")
	}
	if strings.HasPrefix(token, "-") {
		return -magnitude, nil
	}
	return magnitude, nil
}

func saturatingExponent(text string) int64 {
	negative := strings.HasPrefix(text, "-")
	digits := strings.TrimPrefix(strings.TrimPrefix(text, "-"), "+")
	var magnitude int64
	for _, digit := range []byte(digits) {
		if magnitude > (math.MaxInt64-int64(digit-'0'))/10 {
			magnitude = math.MaxInt64
			break
		}
		magnitude = magnitude*10 + int64(digit-'0')
	}
	if negative {
		return -magnitude
	}
	return magnitude
}

func saturatingSubtract(left, right int64) int64 {
	if right < 0 && left > math.MaxInt64+right {
		return math.MaxInt64
	}
	if right > 0 && left < math.MinInt64+right {
		return math.MinInt64
	}
	return left - right
}

func saturatingNegate(value int64) int64 {
	if value == math.MinInt64 {
		return math.MaxInt64
	}
	return -value
}
