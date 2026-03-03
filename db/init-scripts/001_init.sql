\set ON_ERROR_STOP on

-- Create application login roles with least privilege.
DO $$
DECLARE
  app_rw_password text;
  app_ro_password text;
BEGIN
  app_rw_password := regexp_replace(pg_read_file('/run/secrets/db_app_rw_password'), '\s+$', '');
  app_ro_password := regexp_replace(pg_read_file('/run/secrets/db_app_ro_password'), '\s+$', '');

  IF app_rw_password IS NULL OR app_rw_password = '' THEN
    RAISE EXCEPTION 'app_rw_password secret is empty or missing';
  END IF;

  IF app_ro_password IS NULL OR app_ro_password = '' THEN
    RAISE EXCEPTION 'app_ro_password secret is empty or missing';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_rw') THEN
    EXECUTE format(
      'CREATE ROLE app_rw LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION INHERIT CONNECTION LIMIT 50 PASSWORD %L',
      app_rw_password
    );
  ELSE
    EXECUTE format(
      'ALTER ROLE app_rw WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION INHERIT CONNECTION LIMIT 50 PASSWORD %L',
      app_rw_password
    );
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_ro') THEN
    EXECUTE format(
      'CREATE ROLE app_ro LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION INHERIT CONNECTION LIMIT 30 PASSWORD %L',
      app_ro_password
    );
  ELSE
    EXECUTE format(
      'ALTER ROLE app_ro WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION INHERIT CONNECTION LIMIT 30 PASSWORD %L',
      app_ro_password
    );
  END IF;
END
$$;

-- Create a dedicated owner role (no login) for schema object ownership.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_owner') THEN
    CREATE ROLE app_owner
      NOLOGIN
      NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION
      INHERIT;
  END IF;
END
$$;

-- Create the application database if it does not exist.
SELECT 'CREATE DATABASE paperstacks OWNER app_owner TEMPLATE template0 ENCODING ''UTF8'''
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'paperstacks')
\gexec

-- Lock down database-level access and grant only app roles.
REVOKE ALL ON DATABASE paperstacks FROM PUBLIC;
GRANT CONNECT, TEMP ON DATABASE paperstacks TO app_rw;
GRANT CONNECT ON DATABASE paperstacks TO app_ro;

\connect paperstacks

-- Restrict schema-level defaults and grant explicit privileges.
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE, CREATE ON SCHEMA public TO app_owner;
GRANT USAGE ON SCHEMA public TO app_rw;
GRANT USAGE ON SCHEMA public TO app_ro;

-- Default privileges for objects created by the owner role.
ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_rw;

ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA public
  GRANT SELECT ON TABLES TO app_ro;

ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA public
  GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO app_rw;

ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO app_ro;

ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA public
  GRANT EXECUTE ON FUNCTIONS TO app_rw;

ALTER DEFAULT PRIVILEGES FOR ROLE app_owner IN SCHEMA public
  GRANT EXECUTE ON FUNCTIONS TO app_ro;

-- Backfill grants for existing objects in case script is applied to a non-empty DB.
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_rw;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_ro;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO app_rw;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_ro;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO app_rw;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO app_ro;

-- Optional hardening: enforce encrypted transport where supported.
-- ALTER ROLE app_rw SET sslmode = 'require';
-- ALTER ROLE app_ro SET sslmode = 'require';
