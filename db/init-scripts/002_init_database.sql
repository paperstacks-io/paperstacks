-- ** Database generated with pgModeler (PostgreSQL Database Modeler).
-- ** pgModeler version: 2.0.0-beta
-- ** PostgreSQL version: 18.0
-- ** Project Site: pgmodeler.io
-- ** Model Author: ---

-- ** Database creation must be performed outside a multi lined SQL file. 
-- ** These commands were put in this file only as a convenience.

-- object: new_database | type: DATABASE --
-- DROP DATABASE IF EXISTS new_database;
CREATE DATABASE new_database;
-- ddl-end --


-- object: "paperData" | type: SCHEMA --
-- DROP SCHEMA IF EXISTS "paperData" CASCADE;
CREATE SCHEMA "paperData";
-- ddl-end --
ALTER SCHEMA "paperData" OWNER TO postgres;
-- ddl-end --

SET search_path TO pg_catalog,public,"paperData";
-- ddl-end --

-- object: "paperData".affiliation | type: TABLE --
-- DROP TABLE IF EXISTS "paperData".affiliation CASCADE;
CREATE TABLE "paperData".affiliation (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY ,
	name text,
	CONSTRAINT affiliation_pk PRIMARY KEY (key)
);
-- ddl-end --
ALTER TABLE "paperData".affiliation OWNER TO postgres;
-- ddl-end --

-- object: "paperData".author | type: TABLE --
-- DROP TABLE IF EXISTS "paperData".author CASCADE;
CREATE TABLE "paperData".author (
	key bigint NOT NULL GENERATED ALWAYS AS IDENTITY ,
	"nameMiddle" text,
	"nameFirst" text,
	key_affiliation bigint,
	CONSTRAINT author_pk PRIMARY KEY (key)
);
-- ddl-end --
COMMENT ON TABLE "paperData".author IS E'Contains the basic information of an author';
-- ddl-end --
ALTER TABLE "paperData".author OWNER TO postgres;
-- ddl-end --

-- object: affiliation_fk | type: CONSTRAINT --
-- ALTER TABLE "paperData".author DROP CONSTRAINT IF EXISTS affiliation_fk CASCADE;
ALTER TABLE "paperData".author ADD CONSTRAINT affiliation_fk FOREIGN KEY (key_affiliation)
REFERENCES "paperData".affiliation (key) MATCH FULL
ON DELETE SET NULL ON UPDATE CASCADE;
-- ddl-end --

-- object: author_uq | type: CONSTRAINT --
-- ALTER TABLE "paperData".author DROP CONSTRAINT IF EXISTS author_uq CASCADE;
ALTER TABLE "paperData".author ADD CONSTRAINT author_uq UNIQUE (key_affiliation);
-- ddl-end --

-- object: public."paperType" | type: TYPE --
-- DROP TYPE IF EXISTS public."paperType" CASCADE;
CREATE TYPE public."paperType" AS
ENUM ('journal_article','conference_paper','workshop_paper','preprint','thesis','book','book_chapter','technical_report','survey','review','poster','dataset','software','other');
-- ddl-end --
ALTER TYPE public."paperType" OWNER TO postgres;
-- ddl-end --

-- object: public."publicationStatus" | type: TYPE --
-- DROP TYPE IF EXISTS public."publicationStatus" CASCADE;
CREATE TYPE public."publicationStatus" AS
ENUM ('draft','submitted','under_review','accepted','rejected','camera_ready','published','preprint','withdrawn');
-- ddl-end --
ALTER TYPE public."publicationStatus" OWNER TO postgres;
-- ddl-end --

-- object: "paperData".paper | type: TABLE --
-- DROP TABLE IF EXISTS "paperData".paper CASCADE;
CREATE TABLE "paperData".paper (
	uuid char(36) NOT NULL,
	doi char(255),
	title char(500),
	"titleShort" char(200),
	"publicationYear" smallint,
	"paperType" public."paperType",
	"publicationStatus" public."publicationStatus",
	"publicationStatusTimestamp" char(14),
	abstract text,
	keywords text[],
	CONSTRAINT "Unique Doi" UNIQUE NULLS NOT DISTINCT (doi),
	CONSTRAINT paper_pk PRIMARY KEY (uuid)
);
-- ddl-end --
COMMENT ON TABLE "paperData".paper IS E'represents a paper with multiple possible pdfs';
-- ddl-end --
ALTER TABLE "paperData".paper OWNER TO postgres;
-- ddl-end --

-- object: public.many_paper_has_many_author | type: TABLE --
-- DROP TABLE IF EXISTS public.many_paper_has_many_author CASCADE;
CREATE TABLE public.many_paper_has_many_author (
	uuid_paper char(36) NOT NULL,
	key_author bigint NOT NULL,
	CONSTRAINT many_paper_has_many_author_pk PRIMARY KEY (uuid_paper,key_author)
);
-- ddl-end --

-- object: paper_fk | type: CONSTRAINT --
-- ALTER TABLE public.many_paper_has_many_author DROP CONSTRAINT IF EXISTS paper_fk CASCADE;
ALTER TABLE public.many_paper_has_many_author ADD CONSTRAINT paper_fk FOREIGN KEY (uuid_paper)
REFERENCES "paperData".paper (uuid) MATCH FULL
ON DELETE RESTRICT ON UPDATE CASCADE;
-- ddl-end --

-- object: author_fk | type: CONSTRAINT --
-- ALTER TABLE public.many_paper_has_many_author DROP CONSTRAINT IF EXISTS author_fk CASCADE;
ALTER TABLE public.many_paper_has_many_author ADD CONSTRAINT author_fk FOREIGN KEY (key_author)
REFERENCES "paperData".author (key) MATCH FULL
ON DELETE RESTRICT ON UPDATE CASCADE;
-- ddl-end --

-- object: "paperData".pdf | type: TABLE --
-- DROP TABLE IF EXISTS "paperData".pdf CASCADE;
CREATE TABLE "paperData".pdf (
	key smallint NOT NULL,
	"pdfUrl" text,
	uuid_paper char(36),
	CONSTRAINT pdf_pk PRIMARY KEY (key)
);
-- ddl-end --
ALTER TABLE "paperData".pdf OWNER TO postgres;
-- ddl-end --

-- object: paper_fk | type: CONSTRAINT --
-- ALTER TABLE "paperData".pdf DROP CONSTRAINT IF EXISTS paper_fk CASCADE;
ALTER TABLE "paperData".pdf ADD CONSTRAINT paper_fk FOREIGN KEY (uuid_paper)
REFERENCES "paperData".paper (uuid) MATCH FULL
ON DELETE SET NULL ON UPDATE CASCADE;
-- ddl-end --

-- object: "paperData".metadata | type: TABLE --
-- DROP TABLE IF EXISTS "paperData".metadata CASCADE;
CREATE TABLE "paperData".metadata (
	key integer NOT NULL,
	publisher text,
	"publishedIn" text,
	volume smallint,
	issue smallint,
	datasource text,
	"datasourceTimestamp" char(14),
	reference text[],
	isbn char(13)[],
	pages smallint,
	license text,
	copyright text,
	funding text,
	uuid_paper char(36),
	CONSTRAINT metadata_pk PRIMARY KEY (key)
);
-- ddl-end --
COMMENT ON COLUMN "paperData".metadata.volume IS E'volume of publication';
-- ddl-end --
COMMENT ON COLUMN "paperData".metadata.issue IS E'issue of publication';
-- ddl-end --
COMMENT ON COLUMN "paperData".metadata.datasource IS E'url of datasource';
-- ddl-end --
COMMENT ON COLUMN "paperData".metadata.reference IS E'references the paper uses';
-- ddl-end --
ALTER TABLE "paperData".metadata OWNER TO postgres;
-- ddl-end --

-- object: paper_fk | type: CONSTRAINT --
-- ALTER TABLE "paperData".metadata DROP CONSTRAINT IF EXISTS paper_fk CASCADE;
ALTER TABLE "paperData".metadata ADD CONSTRAINT paper_fk FOREIGN KEY (uuid_paper)
REFERENCES "paperData".paper (uuid) MATCH FULL
ON DELETE SET NULL ON UPDATE CASCADE;
-- ddl-end --


