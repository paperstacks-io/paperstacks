// Package tei defines XML-serializable structs representing GROBID TEI-XML documents.
package tei

import "encoding/xml"

type Document struct {
	XMLName xml.Name `xml:"TEI"`
	Header  Header   `xml:"teiHeader"`
	Text    Text     `xml:"text"`
}

type Header struct {
	FileDesc    FileDesc    `xml:"fileDesc"`
	ProfileDesc ProfileDesc `xml:"profileDesc"`
}

type FileDesc struct {
	TitleStmt       TitleStmt       `xml:"titleStmt"`
	PublicationStmt PublicationStmt `xml:"publicationStmt"`
	SourceDesc      SourceDesc      `xml:"sourceDesc"`
}

type TitleStmt struct {
	Titles []Title `xml:"title"`
}

type Title struct {
	Level string `xml:"level,attr"`
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type PublicationStmt struct {
	Publisher string `xml:"publisher"`
	Dates     []Date `xml:"date"`
}

type Date struct {
	Type  string `xml:"type,attr"`
	When  string `xml:"when,attr"`
	Value string `xml:",chardata"`
}

type SourceDesc struct {
	BiblStruct BiblStruct `xml:"biblStruct"`
}

type BiblStruct struct {
	Analytic Analytic `xml:"analytic"`
	Monogr   Monogr   `xml:"monogr"`
	Idnos    []Idno   `xml:"idno"`
}

type Analytic struct {
	Titles  []Title  `xml:"title"`
	Authors []Author `xml:"author"`
	Idnos   []Idno   `xml:"idno"`
}

type Monogr struct {
	Titles  []Title  `xml:"title"`
	Authors []Author `xml:"author"`
	Idnos   []Idno   `xml:"idno"`
	Imprint Imprint  `xml:"imprint"`
}

type Imprint struct {
	Publisher  string      `xml:"publisher"`
	Dates      []Date      `xml:"date"`
	BiblScopes []BiblScope `xml:"biblScope"`
}

type BiblScope struct {
	Unit  string `xml:"unit,attr"`
	From  string `xml:"from,attr"`
	To    string `xml:"to,attr"`
	Value string `xml:",chardata"`
}

type Author struct {
	PersName     PersName      `xml:"persName"`
	Idnos        []Idno        `xml:"idno"`
	Affiliations []Affiliation `xml:"affiliation"`
}

type PersName struct {
	Forenames []Forename `xml:"forename"`
	Surname   string     `xml:"surname"`
}

type Forename struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type Affiliation struct {
	OrgNames []string `xml:"orgName"`
}

type Idno struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type ProfileDesc struct {
	TextClass TextClass `xml:"textClass"`
	Abstract  Abstract  `xml:"abstract"`
}

type TextClass struct {
	Keywords Keywords `xml:"keywords"`
}

type Keywords struct {
	Terms []string `xml:"term"`
	Raw   string   `xml:",chardata"`
}

type Abstract struct {
	Paragraphs []string `xml:"p"`
	Divs       []DivP   `xml:"div"`
}

type DivP struct {
	Paragraphs []string `xml:"p"`
}

type Text struct {
	Back Back `xml:"back"`
}

type Back struct {
	Divs     []BackDiv  `xml:"div"`
	ListBibl []BiblItem `xml:"listBibl>biblStruct"`
}

type BackDiv struct {
	Type     string     `xml:"type,attr"`
	ListBibl []BiblItem `xml:"listBibl>biblStruct"`
}

type BiblItem struct {
	Analytic Analytic `xml:"analytic"`
	Monogr   Monogr   `xml:"monogr"`
	Idnos    []Idno   `xml:"idno"`
}
