// Package parser decodes Crossref public-data-export JSON lines into the
// crawler's domain model.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/domain"
)

// Parse decodes a single JSON-encoded Crossref work record and converts
// it into the crawler's domain model.
//
// Missing or malformed optional fields are silently skipped; Parse only
// returns an error when the line can't be decoded as JSON at all, or the
// record has no DOI, since the DOI is what the rest of the pipeline keys
// on.
func Parse(line []byte) (*domain.Paper, error) {
	var raw rawRecord
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil, fmt.Errorf("decode record: %w", err)
	}

	if raw.DOI == "" {
		return nil, fmt.Errorf("record has no DOI")
	}

	p := &domain.Paper{
		DOI:                 raw.DOI,
		Type:                raw.Type,
		Source:              raw.Source,
		Title:               firstOf(raw.Title),
		Subtitle:            firstOf(raw.Subtitle),
		ContainerTitle:      firstOf(raw.ContainerTitle),
		ShortContainerTitle: firstOf(raw.ShortContainerTitle),
		Abstract:            raw.Abstract,
		Publisher:           raw.Publisher,
		PublisherLocation:   raw.PublisherLocation,
		Member:              raw.Member,
		Prefix:              raw.Prefix,
		Language:            raw.Language,
		URL:                 raw.URL,
		ResourceURL:         raw.Resource.Primary.URL,
		Volume:              raw.Volume,
		Issue:               raw.Issue,
		Page:                raw.Page,
		SpecialNumbering:    raw.SpecialNumbering,
		AlternativeIDs:      raw.AlternativeID,
		ReferenceCount:      raw.ReferenceCount,
		ReferencesCount:     raw.ReferencesCount,
		IsReferencedByCount: raw.IsReferencedByCount,
		Issued:              toPartialDate(raw.Issued),
		Published:           toPartialDate(raw.Published),
		PublishedPrint:      toPartialDate(raw.PublishedPrint),
		PublishedOnline:     toPartialDate(raw.PublishedOnline),
		Created:             toTime(raw.Created),
		Deposited:           toTime(raw.Deposited),
		Indexed:             toTime(raw.Indexed),
		ISSNs:               toIdentifiers(raw.ISSN, raw.ISSNType),
		ISBNs:               toISBNs(raw.ISBN, raw.ISBNType),
		Authors:             toContributors(raw.Author),
		Editors:             toContributors(raw.Editor),
	}

	for _, f := range raw.Funder {
		p.Funders = append(p.Funders, domain.Funder{
			Name:   f.Name,
			DOI:    f.DOI,
			Awards: f.Award,
		})
	}

	for _, ref := range raw.Reference {
		p.References = append(p.References, domain.Reference{
			Key:          ref.Key,
			DOI:          ref.DOI,
			ArticleTitle: ref.ArticleTitle,
			Author:       ref.Author,
			JournalTitle: ref.JournalTitle,
			Volume:       ref.Volume,
			FirstPage:    ref.FirstPage,
			Year:         ref.Year,
			Unstructured: ref.Unstructured,
		})
	}

	for _, lic := range raw.License {
		p.Licenses = append(p.Licenses, domain.License{
			URL:            lic.URL,
			ContentVersion: lic.ContentVersion,
			DelayInDays:    lic.DelayInDays,
			Start:          toPartialDate(lic.Start),
		})
	}

	return p, nil
}

// toIdentifiers merges a plain identifier list (e.g. "ISSN") with its
// typed counterpart (e.g. "issn-type") into domain.ISSN values. The
// typed list is preferred when present since it also carries the medium
// (print/electronic); values only present in the plain list are kept
// with an empty Type.
func toIdentifiers(plain []string, typed []rawTypedID) []domain.ISSN {
	if len(typed) == 0 {
		out := make([]domain.ISSN, 0, len(plain))
		for _, v := range plain {
			out = append(out, domain.ISSN{Value: v})
		}
		return out
	}

	out := make([]domain.ISSN, 0, len(typed))
	for _, t := range typed {
		out = append(out, domain.ISSN{Value: t.Value, Type: t.Type})
	}
	return out
}

func toISBNs(plain []string, typed []rawTypedID) []domain.ISBN {
	if len(typed) == 0 {
		out := make([]domain.ISBN, 0, len(plain))
		for _, v := range plain {
			out = append(out, domain.ISBN{Value: v})
		}
		return out
	}

	out := make([]domain.ISBN, 0, len(typed))
	for _, t := range typed {
		out = append(out, domain.ISBN{Value: t.Value, Type: t.Type})
	}
	return out
}

func toContributors(raw []rawContributor) []domain.Contributor {
	if len(raw) == 0 {
		return nil
	}

	out := make([]domain.Contributor, 0, len(raw))
	for _, c := range raw {
		contributor := domain.Contributor{
			GivenName:  c.Given,
			FamilyName: c.Family,
			Sequence:   c.Sequence,
			ORCID:      cleanORCID(c.ORCID),
		}
		for _, a := range c.Affiliation {
			if a.Name != "" {
				contributor.Affiliations = append(contributor.Affiliations, a.Name)
			}
		}
		out = append(out, contributor)
	}
	return out
}

// toPartialDate extracts a domain.PartialDate from a rawDateField's
// date-parts. Returns the zero value if no usable date-parts are present.
func toPartialDate(f rawDateField) domain.PartialDate {
	if len(f.DateParts) == 0 || len(f.DateParts[0]) == 0 {
		return domain.PartialDate{}
	}
	parts := f.DateParts[0]
	pd := domain.PartialDate{Year: parts[0]}
	if len(parts) >= 2 {
		pd.Month = parts[1]
	}
	if len(parts) >= 3 {
		pd.Day = parts[2]
	}
	return pd
}

// toTime converts a rawDateField to a time.Time. Prefers the date-time
// string, falls back to the millisecond timestamp, then to date-parts.
// Returns the zero time.Time if nothing usable is available.
func toTime(f rawDateField) time.Time {
	if f.DateTime != "" {
		if t, err := time.Parse(time.RFC3339, f.DateTime); err == nil {
			return t
		}
	}
	if f.Timestamp != nil {
		return time.UnixMilli(*f.Timestamp).UTC()
	}
	return toPartialDate(f).Time()
}

// firstOf returns the first non-empty string from a slice, or "" if the
// slice is empty. Crossref represents single-valued fields like "title"
// as one-element arrays.
func firstOf(ss []string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// cleanORCID strips the common ORCID URL prefix so only the bare
// identifier is stored (e.g. "0000-0002-1825-0097").
func cleanORCID(raw string) string {
	raw = strings.TrimPrefix(raw, "http://orcid.org/")
	raw = strings.TrimPrefix(raw, "https://orcid.org/")
	return raw
}
