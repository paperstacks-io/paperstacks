\set ON_ERROR_STOP on

\connect paperstacks

BEGIN;

SET ROLE app_owner;
SET search_path TO public;

CREATE TABLE public.affiliation (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
	name text,
	CONSTRAINT affiliation_pk PRIMARY KEY (key)
);

CREATE TABLE public.author (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
	name_middle text,
	name_first text,
	key_affiliation bigint,
	CONSTRAINT author_pk PRIMARY KEY (key)
);

COMMENT ON TABLE public.author IS E'Contains the basic information of an author';

ALTER TABLE public.author ADD CONSTRAINT affiliation_fk
FOREIGN KEY (key_affiliation)
REFERENCES public.affiliation (key)
ON DELETE SET NULL ON UPDATE CASCADE;

CREATE INDEX author_key_affiliation_idx ON public.author (key_affiliation);

CREATE TYPE public.paper_type AS ENUM (
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

CREATE TYPE public.publication_status AS ENUM (
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

CREATE TABLE public.app_user (
	external_id text NOT NULL,
	email text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT app_user_pk PRIMARY KEY (external_id),
	CONSTRAINT app_user_email_uq UNIQUE (email)
);

COMMENT ON TABLE public.app_user IS
E'Represents a user identified by the external authentication provider';

CREATE TABLE public.paper (
	uuid uuid NOT NULL,
	doi text,
	title text,
	title_short text,
	publication_year smallint,
	paper_type public.paper_type,
	publication_status public.publication_status,
	publication_status_timestamp timestamptz,
	abstract text,
	keywords text[],
	CONSTRAINT paper_doi_uq UNIQUE (doi),
	CONSTRAINT paper_pk PRIMARY KEY (uuid)
);

COMMENT ON TABLE public.paper IS E'represents a paper with multiple possible pdfs';

CREATE TABLE public.stack (
	uuid uuid NOT NULL,
	name text NOT NULL,
	owner_external_id text NOT NULL,
	is_public boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT stack_pk PRIMARY KEY (uuid),
	CONSTRAINT stack_owner_name_uq UNIQUE (owner_external_id, name)
);

COMMENT ON TABLE public.stack IS
E'Represents a named collection of papers owned by a user';

ALTER TABLE public.stack ADD CONSTRAINT stack_owner_fk
FOREIGN KEY (owner_external_id)
REFERENCES public.app_user (external_id)
ON DELETE RESTRICT ON UPDATE CASCADE;

COMMENT ON CONSTRAINT stack_owner_fk ON public.stack IS
E'Prevents deleting a user while that user still owns one or more stacks';

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

CREATE TABLE public.stack_paper (
	uuid_stack uuid NOT NULL,
	uuid_paper uuid NOT NULL,
	CONSTRAINT stack_paper_pk PRIMARY KEY (uuid_stack, uuid_paper)
);

COMMENT ON TABLE public.stack_paper IS
E'Associates papers with stacks';

ALTER TABLE public.stack_paper ADD CONSTRAINT stack_paper_stack_fk
FOREIGN KEY (uuid_stack)
REFERENCES public.stack (uuid)
ON DELETE CASCADE ON UPDATE CASCADE;

COMMENT ON CONSTRAINT stack_paper_stack_fk ON public.stack_paper IS
E'Deleting a stack removes only its paper associations; the papers remain';

ALTER TABLE public.stack_paper ADD CONSTRAINT stack_paper_paper_fk
FOREIGN KEY (uuid_paper)
REFERENCES public.paper (uuid)
ON DELETE CASCADE ON UPDATE CASCADE;

COMMENT ON CONSTRAINT stack_paper_paper_fk ON public.stack_paper IS
E'Deleting a paper removes only its stack associations; the stacks remain';

CREATE INDEX stack_paper_uuid_paper_idx ON public.stack_paper (uuid_paper);

CREATE TABLE public.pdf (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
	pdf_url text,
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
	published_in text,
	volume smallint,
	issue smallint,
	datasource text,
	datasource_timestamp timestamptz,
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

COMMIT;
