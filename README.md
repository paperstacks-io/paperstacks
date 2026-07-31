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

Run `docker compose up` from the repository root to start the application, database, Adminer, and local S3-compatible object storage.

## License

This work (source code) is licensed under [MIT](./LICENSE).
