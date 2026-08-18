package bibliography

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

var ErrInvalidBibLaTeX = errors.New("invalid BibLaTeX")

// Diagnostic identifies source data that was ignored or cannot be represented.
type Diagnostic struct {
	EntryKey string
	Field    string
	Message  string
}

// ImportResult is the result of parsing a complete BibLaTeX document.
type ImportResult struct {
	Imported []PaperEntry
	Failed   []PaperEntry
}

// PaperEntry is one BibLaTeX entry translated to a Paper candidate. Paper
// UUIDs are deliberately left empty: allocation belongs to PaperService.Create.
type PaperEntry struct {
	SourceKey string
	Paper     domain.Paper
	Warnings  []Diagnostic
	Errors    []Diagnostic
}

type bibLaTeXEntry struct {
	typeName       string
	key            string
	fields         map[string]string
	parserWarnings []Diagnostic
}

// ImportBibLaTeX parses BibLaTeX source into Paper candidates without reading
// repositories, allocating UUIDs, or writing data. A syntactically malformed
// document returns an error. Candidate errors are returned in Errors; warnings
// remain attached to their relevant candidate.
func ImportBibLaTeX(source []byte) (ImportResult, error) {
	entries, err := parseBibLaTeX(source)
	if err != nil {
		return ImportResult{}, fmt.Errorf("%w: %v", ErrInvalidBibLaTeX, err)
	}

	result := ImportResult{
		Imported: make([]PaperEntry, 0, len(entries)),
		Failed:   make([]PaperEntry, 0),
	}
	for _, entry := range entries {
		imported := importBibLaTeXEntry(entry)
		if len(imported.Errors) > 0 {
			result.Failed = append(result.Failed, imported)
			continue
		}

		result.Imported = append(result.Imported, imported)
	}

	return result, nil
}

func importBibLaTeXEntry(entry bibLaTeXEntry) PaperEntry {
	paper := domain.Paper{
		Title:      bibLaTeXImportField(entry, "title"),
		TitleShort: bibLaTeXImportField(entry, "shorttitle"),
		DOI:        bibLaTeXImportField(entry, "doi"),
		Abstract:   bibLaTeXImportField(entry, "abstract"),
		Authors:    bibLaTeXAuthorsToDomain(entry.fields["author"]),
		Keywords:   splitBibLaTeXList(bibLaTeXImportField(entry, "keywords")),
		Metadata: domain.Metadata{
			JournalTitle:  firstBibLaTeXField(entry, "journaltitle", "journal"),
			JournalAbbrev: bibLaTeXImportField(entry, "shortjournal"),
			BookTitle:     bibLaTeXImportField(entry, "booktitle"),
			SeriesTitle:   bibLaTeXImportField(entry, "series"),
			EventTitle:    bibLaTeXImportField(entry, "eventtitle"),
			EventLocation: firstBibLaTeXField(entry, "eventplace", "location"),
			Institution:   bibLaTeXImportField(entry, "institution"),
			Publisher:     bibLaTeXImportField(entry, "publisher"),
			Volume:        bibLaTeXImportField(entry, "volume"),
			Issue:         bibLaTeXImportField(entry, "number"),
			Pages:         bibLaTeXImportField(entry, "pages"),
			ISBN:          splitBibLaTeXList(bibLaTeXImportField(entry, "isbn")),
			ISSN:          splitBibLaTeXList(bibLaTeXImportField(entry, "issn")),
		},
	}

	warnings := make([]Diagnostic, 0, len(entry.parserWarnings)+2)
	errors := make([]Diagnostic, 0, 4)
	warnings = append(warnings, entry.parserWarnings...)

	if rawURL := bibLaTeXImportField(entry, "url"); rawURL != "" {
		parsed, err := url.ParseRequestURI(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			warnings = append(warnings, entryDiagnostic(entry.key, "url", "URL is not an absolute HTTP(S) URL and was omitted"))
		} else {
			paper.Metadata.References = []string{parsed.String()}
		}
	}

	for _, field := range unsupportedBibLaTeXFields(entry.fields) {
		warnings = append(warnings, entryDiagnostic(entry.key, field, fmt.Sprintf("BibLaTeX field %q is not represented paperstacks.io", field)))
	}

	publicationType, isSupported := bibLaTeXPublicationType(entry.typeName)
	paper.Type = publicationType
	if !isSupported || !paper.Type.IsValid() {
		errors = append(errors, entryDiagnostic(entry.key, "", fmt.Sprintf("BibLaTeX entry type %q is not a supported publication type", entry.typeName)))
	}

	date, err := bibLaTeXDate(entry)
	paper.PublicationDate = date
	if err != nil || !date.IsValid() {
		errors = append(errors, entryDiagnostic(entry.key, "date", err.Error()))
	}

	if paper.DOI == "" {
		errors = append(errors, entryDiagnostic(entry.key, "doi", "DOI is required"))
	}
	if paper.Title == "" {
		errors = append(errors, entryDiagnostic(entry.key, "title", "title is required"))
	}

	hasAuthors := len(paper.Authors) > 0 && paper.Authors[0].NameLast != "" && paper.Authors[0].NameFirst != ""
	if !hasAuthors {
		errors = append(errors, entryDiagnostic(entry.key, "author", "author is required (first + last name)"))
	}

	return PaperEntry{
		SourceKey: entry.key,
		Paper:     paper,
		Warnings:  warnings,
		Errors:    errors,
	}
}

func entryDiagnostic(entryKey, field, message string) Diagnostic {
	return Diagnostic{
		Message:  message,
		EntryKey: entryKey,
		Field:    field,
	}
}

func bibLaTeXPublicationType(entryType string) (domain.PublicationType, bool) {
	switch strings.ToLower(entryType) {
	case "article":
		return domain.PublicationTypeJournalArticle, true
	case "inproceedings", "conference":
		return domain.PublicationTypeConferenceArticle, true
	case "book":
		return domain.PublicationTypeBook, true
	case "inbook", "incollection":
		return domain.PublicationTypeBookChapter, true
	case "thesis", "mastersthesis", "phdthesis":
		return domain.PublicationTypeThesis, true
	case "report":
		return domain.PublicationTypeReport, true
	case "dataset":
		return domain.PublicationTypeDataset, true
	case "online":
		return domain.PublicationTypeWebPage, true
	default:
		return "", false
	}
}

func bibLaTeXImportField(entry bibLaTeXEntry, name string) string {
	return removeBibLaTeXGrouping(strings.TrimSpace(entry.fields[name]))
}

func firstBibLaTeXField(entry bibLaTeXEntry, names ...string) string {
	for _, name := range names {
		if value := bibLaTeXImportField(entry, name); value != "" {
			return value
		}
	}

	return ""
}

func bibLaTeXDate(entry bibLaTeXEntry) (domain.Date, error) {
	var result domain.Date

	date, _ := parseBibLaTeXDate(bibLaTeXImportField(entry, "date"))
	year, _ := parseBibLaTeXDate(bibLaTeXImportField(entry, "year"))
	month := bibLaTeXMonth(bibLaTeXImportField(entry, "month"))

	if !date.IsZero() && date.IsValid() {
		result = date
	} else if !year.IsZero() && year.IsValid() {
		result = year
	} else {
		return domain.Date{}, errors.New("Unable to parse a date from the fileds 'date' or 'year'")
	}

	if result.Month == 0 && month >= 1 {
		result.Month = month
	}

	return result, nil
}

func parseBibLaTeXDate(value string) (domain.Date, error) {
	for _, candidate := range [...]struct {
		layout    string
		precision int
	}{
		{layout: "2006-01-02", precision: 3},
		{layout: "2006-01", precision: 2},
		{layout: "2006", precision: 1},
	} {
		parsed, err := time.Parse(candidate.layout, value)
		if err != nil {
			continue
		}

		date := domain.Date{Year: parsed.Year()}
		if candidate.precision >= 2 {
			date.Month = int(parsed.Month())
		}
		if candidate.precision == 3 {
			date.Day = parsed.Day()
		}
		if date.Year > 0 && date.IsValid() {
			return date, nil
		}
	}

	return domain.Date{}, fmt.Errorf("invalid BibLaTeX date %q, supported date types are ['2006-01-0', '2006-01', '2006']", value)
}

func bibLaTeXMonth(value string) int {
	months := map[string]int{
		"jan": 1, "january": 1,
		"feb": 2, "february": 2,
		"mar": 3, "march": 3,
		"apr": 4, "april": 4,
		"may": 5,
		"jun": 6, "june": 6,
		"jul": 7, "july": 7,
		"aug": 8, "august": 8,
		"sep": 9, "sept": 9, "september": 9,
		"oct": 10, "october": 10,
		"nov": 11, "november": 11,
		"dec": 12, "december": 12,
	}
	if month, ok := months[strings.ToLower(strings.TrimSpace(value))]; ok {
		return month
	}

	month, _ := strconv.Atoi(value)
	if month >= 1 && month <= 12 {
		return month
	}
	return 0
}

func bibLaTeXAuthorsToDomain(value string) []domain.Author {
	parts := splitBibLaTeXNames(value)
	authors := make([]domain.Author, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			authors = append(authors, domain.Author{NameLast: strings.TrimSpace(removeBibLaTeXGrouping(part))})
			continue
		}

		nameParts := splitBibLaTeXCommas(part)
		if len(nameParts) >= 2 {
			last := strings.TrimSpace(removeBibLaTeXGrouping(nameParts[0]))
			given := strings.TrimSpace(removeBibLaTeXGrouping(nameParts[len(nameParts)-1]))
			first, middle := splitGivenName(given)
			if len(nameParts) > 2 {
				last = strings.TrimSpace(last + ", " + strings.Join(nameParts[1:len(nameParts)-1], ", "))
			}
			authors = append(authors, domain.Author{NameFirst: first, NameMiddle: middle, NameLast: last})
			continue
		}

		words := strings.Fields(removeBibLaTeXGrouping(part))
		switch len(words) {
		case 0:
		case 1:
			authors = append(authors, domain.Author{NameLast: words[0]})
		default:
			lastStart := len(words) - 1
			for lastStart > 0 && startsLower(words[lastStart-1]) {
				lastStart--
			}
			first, middle := splitGivenName(strings.Join(words[:lastStart], " "))
			authors = append(authors, domain.Author{NameFirst: first, NameMiddle: middle, NameLast: strings.Join(words[lastStart:], " ")})
		}
	}

	return authors
}

func splitBibLaTeXNames(value string) []string {
	return splitBibLaTeXDelimited(value, func(value string, index int) (int, bool) {
		if index+5 > len(value) || !strings.EqualFold(value[index:index+5], " and ") {
			return 0, false
		}
		return 5, true
	})
}

func splitBibLaTeXCommas(value string) []string {
	return splitBibLaTeXDelimited(value, func(value string, index int) (int, bool) {
		if value[index] != ',' {
			return 0, false
		}
		return 1, true
	})
}

func splitBibLaTeXList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitBibLaTeXDelimited(value string, delimiter func(string, int) (int, bool)) []string {
	parts := make([]string, 0, 1)
	start := 0
	depth := 0
	for index := 0; index < len(value); {
		switch value[index] {
		case '\\':
			index += 2
			continue
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				if width, ok := delimiter(value, index); ok {
					parts = append(parts, value[start:index])
					index += width
					start = index
					continue
				}
			}
		}
		_, width := utf8.DecodeRuneInString(value[index:])
		index += width
	}
	parts = append(parts, value[start:])
	return parts
}

func splitGivenName(value string) (string, string) {
	words := strings.Fields(value)
	if len(words) == 0 {
		return "", ""
	}
	return words[0], strings.Join(words[1:], " ")
}

func startsLower(value string) bool {
	for _, r := range value {
		return unicode.IsLower(r)
	}
	return false
}

// removeBibLaTeXGrouping removes the outer field delimiters.
// Example: "some {grouped} value" becomes "some grouped value"
func removeBibLaTeXGrouping(value string) string {
	var out strings.Builder
	out.Grow(len(value))
	escaped := false
	for _, r := range value {
		if escaped {
			out.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			out.WriteRune(r)
			escaped = true
			continue
		}
		if r != '{' && r != '}' {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func unsupportedBibLaTeXFields(fields map[string]string) []string {
	supported := map[string]struct{}{
		"title": {}, "shorttitle": {}, "author": {}, "date": {}, "year": {}, "month": {},
		"doi": {}, "url": {}, "abstract": {}, "keywords": {}, "isbn": {}, "issn": {},
		"journaltitle": {}, "journal": {}, "shortjournal": {}, "booktitle": {}, "series": {},
		"eventtitle": {}, "eventplace": {}, "location": {}, "institution": {}, "publisher": {},
		"volume": {}, "number": {}, "pages": {},
	}

	unsupported := make([]string, 0)
	for field := range fields {
		if _, ok := supported[field]; !ok {
			unsupported = append(unsupported, field)
		}
	}
	sort.Strings(unsupported)
	return unsupported
}

type bibLaTeXParser struct {
	source   []byte
	position int
	macros   map[string]string
}

func parseBibLaTeX(source []byte) ([]bibLaTeXEntry, error) {
	parser := bibLaTeXParser{source: source, macros: make(map[string]string)}
	var entries []bibLaTeXEntry
	for {
		parser.skipSpaceAndComments()
		if parser.eof() {
			return entries, nil
		}
		if parser.source[parser.position] != '@' {
			parser.position++
			continue
		}
		parser.position++
		entryType := strings.ToLower(parser.readIdentifier())
		if entryType == "" {
			return nil, fmt.Errorf("byte %d: expected entry type after @", parser.position)
		}
		parser.skipSpace()
		close, err := parser.closingDelimiter()
		if err != nil {
			return nil, err
		}

		switch entryType {
		case "comment":
			if err := parser.skipBalanced(close); err != nil {
				return nil, err
			}
		case "preamble":
			if _, _, err := parser.readValue(close); err != nil {
				return nil, err
			}
			parser.skipSpace()
			if err := parser.expect(close); err != nil {
				return nil, err
			}
		case "string":
			if err := parser.parseString(close); err != nil {
				return nil, err
			}
		default:
			entry, err := parser.parseEntry(entryType, close)
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}
}

func (parser *bibLaTeXParser) parseString(close byte) error {
	parser.skipSpaceAndComments()
	name := strings.ToLower(parser.readIdentifier())
	if name == "" {
		return fmt.Errorf("byte %d: expected string name", parser.position)
	}
	parser.skipSpace()
	if err := parser.expect('='); err != nil {
		return err
	}
	value, _, err := parser.readValue(close)
	if err != nil {
		return err
	}
	parser.macros[name] = value
	parser.skipSpace()
	if parser.peek() == ',' {
		parser.position++
		parser.skipSpace()
	}
	return parser.expect(close)
}

func (parser *bibLaTeXParser) parseEntry(entryType string, close byte) (bibLaTeXEntry, error) {
	parser.skipSpaceAndComments()
	keyStart := parser.position
	for !parser.eof() && parser.peek() != ',' && parser.peek() != close {
		parser.position++
	}
	key := strings.TrimSpace(string(parser.source[keyStart:parser.position]))
	if key == "" {
		return bibLaTeXEntry{}, fmt.Errorf("byte %d: expected entry key", parser.position)
	}

	entry := bibLaTeXEntry{typeName: entryType, key: key, fields: make(map[string]string)}
	if parser.peek() == close {
		parser.position++
		return entry, nil
	}
	if err := parser.expect(','); err != nil {
		return bibLaTeXEntry{}, err
	}

	for {
		parser.skipSpaceAndComments()
		if parser.peek() == close {
			parser.position++
			return entry, nil
		}
		field := strings.ToLower(parser.readIdentifier())
		if field == "" {
			return bibLaTeXEntry{}, fmt.Errorf("byte %d: expected field name", parser.position)
		}
		parser.skipSpace()
		if err := parser.expect('='); err != nil {
			return bibLaTeXEntry{}, err
		}
		value, diagnostics, err := parser.readValue(close)
		if err != nil {
			return bibLaTeXEntry{}, err
		}
		if _, exists := entry.fields[field]; exists {
			entry.parserWarnings = append(entry.parserWarnings, Diagnostic{
				EntryKey: entry.key,
				Field:    field,
				Message:  fmt.Sprintf("BibLaTeX field %q appears more than once; the last value was used", field),
			})
		}
		entry.fields[field] = value
		for _, diagnostic := range diagnostics {
			if diagnostic.Field == "" {
				diagnostic.Field = field
			}
			diagnostic.EntryKey = entry.key
			entry.parserWarnings = append(entry.parserWarnings, diagnostic)
		}

		parser.skipSpaceAndComments()
		if parser.peek() == close {
			parser.position++
			return entry, nil
		}
		if err := parser.expect(','); err != nil {
			return bibLaTeXEntry{}, err
		}
	}
}

func (parser *bibLaTeXParser) readValue(close byte) (string, []Diagnostic, error) {
	parser.skipSpace()
	var value strings.Builder
	diagnostics := make([]Diagnostic, 0)
	for {
		part, diagnostic, err := parser.readValuePart(close)
		if err != nil {
			return "", nil, err
		}
		value.WriteString(part)
		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
		parser.skipSpace()
		if parser.peek() != '#' {
			return value.String(), diagnostics, nil
		}
		parser.position++
		parser.skipSpace()
	}
}

func (parser *bibLaTeXParser) readValuePart(close byte) (string, *Diagnostic, error) {
	if parser.eof() || parser.peek() == close || parser.peek() == ',' {
		return "", nil, fmt.Errorf("byte %d: expected field value", parser.position)
	}
	switch parser.peek() {
	case '{':
		value, err := parser.readBracedValue()
		return value, nil, err
	case '"':
		value, err := parser.readQuotedValue()
		return value, nil, err
	default:
		start := parser.position
		for !parser.eof() && !isBibLaTeXValueDelimiter(parser.peek(), close) {
			parser.position++
		}
		name := strings.TrimSpace(string(parser.source[start:parser.position]))
		if name == "" {
			return "", nil, fmt.Errorf("byte %d: expected field value", parser.position)
		}
		if value, ok := parser.macros[strings.ToLower(name)]; ok {
			return value, nil, nil
		}
		if _, err := strconv.Atoi(name); err == nil || bibLaTeXMonth(name) != 0 {
			return name, nil, nil
		}
		return name, &Diagnostic{Message: fmt.Sprintf("BibLaTeX string %q is undefined and was preserved literally", name)}, nil
	}
}

func (parser *bibLaTeXParser) readBracedValue() (string, error) {
	parser.position++
	depth := 1
	start := parser.position
	for !parser.eof() {
		current := parser.peek()
		if current == '\\' {
			parser.position++
			if !parser.eof() {
				parser.position++
			}
			continue
		}
		parser.position++
		switch current {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return string(parser.source[start : parser.position-1]), nil
			}
		}
	}
	return "", fmt.Errorf("byte %d: unterminated braced value", parser.position)
}

func (parser *bibLaTeXParser) readQuotedValue() (string, error) {
	parser.position++
	start := parser.position
	for !parser.eof() {
		current := parser.peek()
		if current == '\\' {
			parser.position++
			if !parser.eof() {
				parser.position++
			}
			continue
		}
		parser.position++
		if current == '"' {
			return string(parser.source[start : parser.position-1]), nil
		}
	}
	return "", fmt.Errorf("byte %d: unterminated quoted value", parser.position)
}

func (parser *bibLaTeXParser) skipBalanced(close byte) error {
	depth := 1
	for !parser.eof() {
		current := parser.peek()
		parser.position++
		if current == '\\' && !parser.eof() {
			parser.position++
			continue
		}
		if current == close {
			depth--
			if depth == 0 {
				return nil
			}
		}
		if (close == '}' && current == '{') || (close == ')' && current == '(') {
			depth++
		}
	}
	return fmt.Errorf("byte %d: unterminated @comment", parser.position)
}

func (parser *bibLaTeXParser) closingDelimiter() (byte, error) {
	if parser.eof() {
		return 0, fmt.Errorf("byte %d: expected { or (", parser.position)
	}
	switch parser.peek() {
	case '{':
		parser.position++
		return '}', nil
	case '(':
		parser.position++
		return ')', nil
	default:
		return 0, fmt.Errorf("byte %d: expected { or (", parser.position)
	}
}

func (parser *bibLaTeXParser) readIdentifier() string {
	start := parser.position
	for !parser.eof() {
		current := parser.peek()
		if current == '=' || current == ',' || current == '{' || current == '}' || current == '(' || current == ')' || current == '#' || current == '"' || isBibLaTeXSpace(current) {
			break
		}
		parser.position++
	}
	return strings.TrimSpace(string(parser.source[start:parser.position]))
}

func (parser *bibLaTeXParser) skipSpaceAndComments() {
	for {
		parser.skipSpace()
		if parser.eof() || parser.peek() != '%' {
			return
		}
		for !parser.eof() && parser.peek() != '\n' {
			parser.position++
		}
	}
}

func (parser *bibLaTeXParser) skipSpace() {
	for !parser.eof() && isBibLaTeXSpace(parser.peek()) {
		parser.position++
	}
}

func (parser *bibLaTeXParser) expect(expected byte) error {
	if parser.eof() || parser.peek() != expected {
		return fmt.Errorf("byte %d: expected %q", parser.position, expected)
	}
	parser.position++
	return nil
}

func (parser *bibLaTeXParser) eof() bool {
	return parser.position >= len(parser.source)
}

func (parser *bibLaTeXParser) peek() byte {
	if parser.eof() {
		return 0
	}
	return parser.source[parser.position]
}

func isBibLaTeXValueDelimiter(value, close byte) bool {
	return value == ',' || value == close || value == '#' || isBibLaTeXSpace(value)
}

func isBibLaTeXSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}
