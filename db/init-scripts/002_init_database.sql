\connect paperstacks

SET ROLE app_owner;
SET search_path TO public;

CREATE TABLE public.affiliation (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
	name text,
	CONSTRAINT affiliation_pk PRIMARY KEY (key)
);

CREATE TABLE public.author (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
	"nameMiddle" text,
	"nameFirst" text,
	key_affiliation bigint,
	CONSTRAINT author_pk PRIMARY KEY (key)
);

COMMENT ON TABLE public.author IS E'Contains the basic information of an author';

ALTER TABLE public.author ADD CONSTRAINT affiliation_fk
FOREIGN KEY (key_affiliation)
REFERENCES public.affiliation (key)
ON DELETE SET NULL ON UPDATE CASCADE;

CREATE INDEX author_key_affiliation_idx ON public.author (key_affiliation);

CREATE TYPE public."paperType" AS ENUM (
	'journal_article',
	'conference_paper',
	'workshop_paper',
	'preprint',
	'thesis',
	'book',
	'book_chapter',
	'technical_report',
	'survey',
	'review',
	'poster',
	'dataset',
	'software',
	'other'
);

CREATE TYPE public."publicationStatus" AS ENUM (
	'draft',
	'submitted',
	'under_review',
	'accepted',
	'rejected',
	'camera_ready',
	'published',
	'preprint',
	'withdrawn'
);

CREATE TABLE public.paper (
	uuid uuid NOT NULL,
	doi text,
	title text,
	"titleShort" text,
	"publicationYear" smallint,
	"paperType" public."paperType",
	"publicationStatus" public."publicationStatus",
	"publicationStatusTimestamp" timestamptz,
	abstract text,
	keywords text[],
	CONSTRAINT paper_doi_uq UNIQUE (doi),
	CONSTRAINT paper_pk PRIMARY KEY (uuid)
);

COMMENT ON TABLE public.paper IS E'represents a paper with multiple possible pdfs';

CREATE TABLE public.paper_author (
	uuid_paper uuid NOT NULL,
	key_author bigint NOT NULL,
	CONSTRAINT paper_author_pk PRIMARY KEY (uuid_paper, key_author)
);

ALTER TABLE public.paper_author ADD CONSTRAINT paper_author_paper_fk
FOREIGN KEY (uuid_paper)
REFERENCES public.paper (uuid)
ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE public.paper_author ADD CONSTRAINT paper_author_author_fk
FOREIGN KEY (key_author)
REFERENCES public.author (key)
ON DELETE RESTRICT ON UPDATE CASCADE;

CREATE INDEX paper_author_key_author_idx ON public.paper_author (key_author);

CREATE TABLE public.pdf (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
	"pdfUrl" text,
	uuid_paper uuid NOT NULL,
	CONSTRAINT pdf_pk PRIMARY KEY (key)
);

ALTER TABLE public.pdf ADD CONSTRAINT pdf_paper_fk
FOREIGN KEY (uuid_paper)
REFERENCES public.paper (uuid)
ON DELETE CASCADE ON UPDATE CASCADE;

CREATE INDEX pdf_uuid_paper_idx ON public.pdf (uuid_paper);

CREATE TABLE public.metadata (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
	publisher text,
	"publishedIn" text,
	volume smallint,
	issue smallint,
	datasource text,
	"datasourceTimestamp" timestamptz,
	reference text[],
	isbn text[],
	pages smallint,
	license text,
	copyright text,
	funding text,
	uuid_paper uuid NOT NULL,
	CONSTRAINT metadata_pk PRIMARY KEY (key)
);

COMMENT ON COLUMN public.metadata.volume IS E'volume of publication';
COMMENT ON COLUMN public.metadata.issue IS E'issue of publication';
COMMENT ON COLUMN public.metadata.datasource IS E'url of datasource';
COMMENT ON COLUMN public.metadata.reference IS E'references the paper uses';

ALTER TABLE public.metadata ADD CONSTRAINT metadata_paper_fk
FOREIGN KEY (uuid_paper)
REFERENCES public.paper (uuid)
ON DELETE CASCADE ON UPDATE CASCADE;

CREATE INDEX metadata_uuid_paper_idx ON public.metadata (uuid_paper);
