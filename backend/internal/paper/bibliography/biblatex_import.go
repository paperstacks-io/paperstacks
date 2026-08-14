package bibliography

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
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
	Entries []ImportedPaper
	Errors  []Diagnostic
}

// ImportedPaper is one BibLaTeX entry translated to a Paper candidate. Paper
// UUIDs are deliberately left empty: allocation belongs to PaperService.Create.
type ImportedPaper struct {
	SourceKey string
	Paper     domain.Paper
	Warnings  []Diagnostic
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
		Entries: make([]ImportedPaper, 0, len(entries)),
		Errors:  make([]Diagnostic, 0),
	}
	for _, entry := range entries {
		imported, errors := importBibLaTeXEntry(entry)
		result.Entries = append(result.Entries, imported)
		result.Errors = append(result.Errors, errors...)
	}

	return result, nil
}

func importBibLaTeXEntry(entry bibLaTeXEntry) (ImportedPaper, []Diagnostic) {
	paper := domain.Paper{
		Title:      bibLaTeXImportField(entry, "title"),
		TitleShort: bibLaTeXImportField(entry, "shorttitle"),
		DOI:        bibLaTeXImportField(entry, "doi"),
		Abstract:   bibLaTeXImportField(entry, "abstract"),
		Authors:    bibLaTeXAuthorsToDomain(bibLaTeXRawField(entry, "author")),
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

	publicationType, supportedType := bibLaTeXPublicationType(entry.typeName)
	if !supportedType {
		paper.Type = domain.PublicationType(entry.typeName)
		errors = append(errors, entryDiagnostic(entry.key, "", fmt.Sprintf("BibLaTeX entry type %q has no Paper type mapping", entry.typeName)))
	} else {
		paper.Type = publicationType
	}

	date, dateDiagnostic := bibLaTeXDate(entry)
	paper.PublicationDate = date
	if dateDiagnostic != nil {
		dateDiagnostic.EntryKey = entry.key
		errors = append(errors, *dateDiagnostic)
	}

	if rawURL := bibLaTeXImportField(entry, "url"); rawURL != "" {
		parsed, err := url.ParseRequestURI(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			warnings = append(warnings, entryDiagnostic(entry.key, "url", "URL is not an absolute HTTP(S) URL and was omitted"))
		} else {
			paper.Metadata.References = []string{parsed.String()}
		}
	}

	for _, field := range unsupportedBibLaTeXFields(entry.fields) {
		warnings = append(warnings, entryDiagnostic(entry.key, field, fmt.Sprintf("BibLaTeX field %q is not represented by Paper", field)))
	}

	paper = paper.Normalize()
	if paper.DOI == "" {
		errors = append(errors, entryDiagnostic(entry.key, "doi", "DOI is required"))
	}
	if paper.Title == "" {
		errors = append(errors, entryDiagnostic(entry.key, "title", "title is required"))
	}
	if !paper.PublicationDate.IsValid() {
		errors = append(errors, entryDiagnostic(entry.key, "date", "publication date cannot be parsed"))
	}
	if !paper.Type.IsValid() {
		errors = append(errors, entryDiagnostic(entry.key, "", "publication type not supported"))
	}

	return ImportedPaper{
		SourceKey: entry.key,
		Paper:     paper,
		Warnings:  warnings,
	}, errors
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
	return removeBibLaTeXGrouping(bibLaTeXRawField(entry, name))
}

func bibLaTeXRawField(entry bibLaTeXEntry, name string) string {
	return strings.TrimSpace(entry.fields[name])
}

func firstBibLaTeXField(entry bibLaTeXEntry, names ...string) string {
	for _, name := range names {
		if value := bibLaTeXImportField(entry, name); value != "" {
			return value
		}
	}

	return ""
}

func bibLaTeXDate(entry bibLaTeXEntry) (domain.Date, *Diagnostic) {
	if value := bibLaTeXImportField(entry, "date"); value != "" {
		date, err := parseBibLaTeXDate(value)
		if err != nil {
			diagnostic := entryDiagnostic("", "date", err.Error())
			return domain.Date{}, &diagnostic
		}
		return date, nil
	}

	year := bibLaTeXImportField(entry, "year")
	if year == "" {
		return domain.Date{}, nil
	}

	parsedYear, err := strconv.Atoi(year)
	if err != nil || parsedYear < 1 {
		diagnostic := entryDiagnostic("", "year", "year is not a positive integer")
		return domain.Date{}, &diagnostic
	}

	date := domain.Date{Year: parsedYear}
	if month := bibLaTeXMonth(bibLaTeXImportField(entry, "month")); month != 0 {
		date.Month = month
	} else if bibLaTeXImportField(entry, "month") != "" {
		diagnostic := entryDiagnostic("", "month", "month is not a valid BibLaTeX month")
		return domain.Date{}, &diagnostic
	}

	return date, nil
}

func parseBibLaTeXDate(value string) (domain.Date, error) {
	if strings.Contains(value, "/") {
		return domain.Date{}, errors.New("date ranges cannot be represented by Paper")
	}

	parts := strings.Split(value, "-")
	if len(parts) < 1 || len(parts) > 3 {
		return domain.Date{}, fmt.Errorf("date %q is not year, year-month, or full-date", value)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 1 {
		return domain.Date{}, fmt.Errorf("date %q has an invalid year", value)
	}
	date := domain.Date{Year: year}
	if len(parts) >= 2 {
		month, err := strconv.Atoi(parts[1])
		if err != nil || month < 1 || month > 12 {
			return domain.Date{}, fmt.Errorf("date %q has an invalid month", value)
		}
		date.Month = month
	}
	if len(parts) == 3 {
		day, err := strconv.Atoi(parts[2])
		if err != nil || day < 1 {
			return domain.Date{}, fmt.Errorf("date %q has an invalid day", value)
		}
		date.Day = day
	}
	if !date.IsValid() {
		return domain.Date{}, fmt.Errorf("date %q is not a calendar date", value)
	}

	return date, nil
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
			return nil, parser.errorf("expected entry type after @")
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
		return parser.errorf("expected string name")
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
		return bibLaTeXEntry{}, parser.errorf("expected entry key")
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
			return bibLaTeXEntry{}, parser.errorf("expected field name")
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
		return "", nil, parser.errorf("expected field value")
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
			return "", nil, parser.errorf("expected field value")
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
	return "", parser.errorf("unterminated braced value")
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
	return "", parser.errorf("unterminated quoted value")
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
	return parser.errorf("unterminated @comment")
}

func (parser *bibLaTeXParser) closingDelimiter() (byte, error) {
	if parser.eof() {
		return 0, parser.errorf("expected { or (")
	}
	switch parser.peek() {
	case '{':
		parser.position++
		return '}', nil
	case '(':
		parser.position++
		return ')', nil
	default:
		return 0, parser.errorf("expected { or (")
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
		return parser.errorf("expected %q", expected)
	}
	parser.position++
	return nil
}

func (parser *bibLaTeXParser) errorf(format string, args ...any) error {
	return fmt.Errorf("byte %d: %s", parser.position, fmt.Sprintf(format, args...))
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
