package pdfextractor

import (
	"reflect"
	"strings"
	"testing"

	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestParseTEIFullDocument(t *testing.T) {
	t.Parallel()

	rawXML := `<?xml version="1.0" encoding="UTF-8"?>
<TEI xmlns="http://www.tei-c.org/ns/1.0">
  <teiHeader>
    <fileDesc>
      <titleStmt>
        <title level="a" type="main">Attention Is All You Need</title>
      </titleStmt>
      <publicationStmt>
        <publisher>Curran Associates</publisher>
        <date type="published" when="2017-12-06">6 Dec 2017</date>
      </publicationStmt>
      <sourceDesc>
        <biblStruct>
          <analytic>
            <title level="a" type="main">Attention Is All You Need</title>
            <author>
              <persName>
                <forename type="first">Ashish</forename>
                <forename type="middle">K</forename>
                <surname>Vaswani</surname>
              </persName>
              <idno type="ORCID">0000-0001-2345-6789</idno>
              <affiliation key="aff0">
                <orgName type="institution">Google Brain</orgName>
              </affiliation>
            </author>
            <author>
              <persName>
                <forename type="first">Noam</forename>
                <surname>Shazeer</surname>
              </persName>
            </author>
            <idno type="DOI">10.48550/arXiv.1706.03762</idno>
          </analytic>
          <monogr>
            <title level="j" type="main">Advances in Neural Information Processing Systems</title>
            <title level="j" type="abbrev">NeurIPS</title>
            <idno type="ISSN">1049-5258</idno>
            <imprint>
              <publisher>NeurIPS Foundation</publisher>
              <biblScope unit="volume">30</biblScope>
              <biblScope unit="issue">1</biblScope>
              <biblScope unit="page" from="5998" to="6008"/>
            </imprint>
          </monogr>
        </biblStruct>
      </sourceDesc>
    </fileDesc>
    <profileDesc>
      <textClass>
        <keywords>
          <term>transformers</term>
          <term>self-attention</term>
        </keywords>
      </textClass>
      <abstract>
        <div xmlns="http://www.tei-c.org/ns/1.0">
          <p>The dominant sequence transduction models are based on complex recurrent or convolutional neural networks.</p>
        </div>
      </abstract>
    </profileDesc>
  </teiHeader>
  <text>
    <body>
      <div><p>Introduction body text.</p></div>
    </body>
    <back>
      <div type="references">
        <listBibl>
          <biblStruct xml:id="b0">
            <analytic>
              <title level="a" type="main">Neural Machine Translation by Jointly Learning to Align and Translate</title>
              <idno type="DOI">10.1234/bahdanau2014</idno>
            </analytic>
          </biblStruct>
          <biblStruct xml:id="b1">
            <analytic>
              <title level="a" type="main">Long Short-Term Memory</title>
            </analytic>
          </biblStruct>
        </listBibl>
      </div>
    </back>
  </text>
</TEI>`

	got, err := ParseTEI(strings.NewReader(rawXML))
	if err != nil {
		t.Fatalf("ParseTEI() error = %v", err)
	}

	if got.Title != "Attention Is All You Need" {
		t.Errorf("Title = %q, want %q", got.Title, "Attention Is All You Need")
	}
	if got.DOI != "10.48550/arXiv.1706.03762" {
		t.Errorf("DOI = %q, want %q", got.DOI, "10.48550/arXiv.1706.03762")
	}
	if got.PublicationDate != (paper.Date{Year: 2017, Month: 12, Day: 6}) {
		t.Errorf("PublicationDate = %+v, want 2017-12-06", got.PublicationDate)
	}
	if got.Abstract != "The dominant sequence transduction models are based on complex recurrent or convolutional neural networks." {
		t.Errorf("Abstract = %q", got.Abstract)
	}

	wantKeywords := []string{"transformers", "self-attention"}
	if !reflect.DeepEqual(got.Keywords, wantKeywords) {
		t.Errorf("Keywords = %v, want %v", got.Keywords, wantKeywords)
	}

	wantAuthors := []paper.Author{
		{
			NameFirst:   "Ashish",
			NameMiddle:  "K",
			NameLast:    "Vaswani",
			Affiliation: "Google Brain",
			ORCID:       "0000-0001-2345-6789",
		},
		{
			NameFirst: "Noam",
			NameLast:  "Shazeer",
		},
	}
	if !reflect.DeepEqual(got.Authors, wantAuthors) {
		t.Errorf("Authors = %+v, want %+v", got.Authors, wantAuthors)
	}

	if got.Metadata.Publisher != "NeurIPS Foundation" {
		t.Errorf("Publisher = %q, want %q", got.Metadata.Publisher, "NeurIPS Foundation")
	}
	if got.Metadata.JournalTitle != "Advances in Neural Information Processing Systems" {
		t.Errorf("JournalTitle = %q", got.Metadata.JournalTitle)
	}
	if got.Metadata.JournalAbbrev != "NeurIPS" {
		t.Errorf("JournalAbbrev = %q", got.Metadata.JournalAbbrev)
	}
	if got.Metadata.Volume != "30" || got.Metadata.Issue != "1" || got.Metadata.Pages != "5998-6008" {
		t.Errorf("Volume/Issue/Pages = %q/%q/%q", got.Metadata.Volume, got.Metadata.Issue, got.Metadata.Pages)
	}

	wantRefs := []string{"10.1234/bahdanau2014", "Long Short-Term Memory"}
	if !reflect.DeepEqual(got.Metadata.References, wantRefs) {
		t.Errorf("References = %v, want %v", got.Metadata.References, wantRefs)
	}

	if missing := MissingFields(got); len(missing) != 0 {
		t.Errorf("MissingFields() = %v, want empty", missing)
	}
}

func TestMissingFields(t *testing.T) {
	t.Parallel()

	empty := paper.Paper{}
	want := []string{"title", "authors", "abstract", "doi", "publication_date", "keywords"}
	if got := MissingFields(empty); !reflect.DeepEqual(got, want) {
		t.Fatalf("MissingFields(empty) = %v, want %v", got, want)
	}
}

func TestParseTEIErrors(t *testing.T) {
	t.Parallel()

	if _, err := ParseTEI(nil); err == nil {
		t.Fatal("ParseTEI(nil) expected error, got nil")
	}
	if _, err := ParseTEI(strings.NewReader("<<<not valid xml")); err == nil {
		t.Fatal("ParseTEI(invalid) expected error, got nil")
	}
}
