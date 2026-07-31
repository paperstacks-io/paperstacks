<div align="center">
    <img src="https://paperstacks.io/img/logo-300.png" alt="PaperStacks Logo" height="150"/>
    <img src="https://paperstacks.io/img/text.png" alt="PaperStacks Logo" width="600"/>
</div>

## Find, stack, sort, and share papers for your research

The scientific publication ecosystem relies heavily on proprietary platforms for search, indexing, and reference management, resulting in opaque processes, strong vendor lock-in, and limited data sovereignty. At the same time, open alternatives are fragmented, poorly interoperable, and lack reliable linking between different publication versions. Although open metadata sources exist, there is no integrated, transparent, and reproducible infrastructure that supports the full scientific workflow. Consequently, researchers and institutions lack long-term control, interoperability, and clarity over access, versions, and use of scientific publications, especially from a European perspective.

## Mission

Our mission is to empower researchers with open, interoperable, and sovereign infrastructure for scientific knowledge.

Visit our [landing page](https://paperstacks.io/) for more information.

## Project structure

- `docs/`: Contains more technical information and documentation about the project.
- `backend/`: Contains the backend code, including API endpoints and the HTMX frontend (Server Side Rendering).
- `db/`: Contains all database (PostgreSQL) related scripts.
- `objectstorage/`: Contains the local S3-compatible object storage setup.

## Prerequisites

[Docker](https://www.docker.com/products/docker-desktop/) with Compose are required to run the project locally. Please ensure you have them installed before proceeding.

Passwords and credentials for the database are stored in the `secrets` directory that is ignored by git.
Run `make secrets` to create the necessary files with random passwords.
The `paperstacks` service loads runtime configuration, including `HANKO_API_URL`, from `backend/.env`.

Run `docker compose up` from the repository root to start the application, database, Adminer, and local S3-compatible object storage.

## Production deployment

Arcane deploys this application with `compose.yaml` and `compose.prod.yaml`; Traefik is provisioned independently by the platform stack. Configure these deployment variables in Arcane before deploying:

- `PAPERSTACKS_IMAGE_TAG`: a published Paperstacks image tag, for example `v0.2.0`.
- `PAPERSTACKS_HOST`: the public hostname for Paperstacks.
- `RUSTFS_HOST`: the public hostname for the RustFS S3 API; the management console remains private.
- `TRAEFIK_CERT_RESOLVER`: the configured Traefik certificate resolver.
- `TRAEFIK_NETWORK`: optional external Traefik network name; defaults to `edge`.
- `RUSTFS_ACCESS_KEY` and `RUSTFS_SECRET_KEY`: private RustFS credentials.

The platform stack must create the external Traefik network and its certificate resolver before deployment. Provision the untracked `backend/.env` file on the deployment host with `HANKO_API_URL`.

## License

This work (source code) is licensed under [MIT](./LICENSE).
