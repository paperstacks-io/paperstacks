# paperstacks.io project description

## Scope

paperstacks.io enables users to

- search for academic papers
- organize papers in named stacks
- view and improve reliable paper metadata
- share stacks publicly or with selected users (later groups or organizations)
- export citations

## Supported user stories

### Anonymous user

- Search for papers
- View paper details and verification
- Browse public stacks
- View public stack contents
- Upload papers for verification for all users

### Registered user

- Create stacks
- Add papers to stacks (by identifier or upload)
- Organize stack contents (add, pin, sort, remove)
- Share stacks (private / public / shared with users)
- Export stack citations (different formats)
- Subscribe and unsubscribe to stacks
- Manage account (including deletion)

### DMCA reviewer

- Review DMCA takedown requests
- Inspect referenced submissions, papers, and stacks
- Temporarily disable access to disputed content
- Permanently remove content when claims are validated
- Restore content when claims are rejected or withdrawn
- Record decisions and rationale for each action
- View an audit trail of takedown actions

### Out of scope for MVP

- Comments or annotations shared across users
- Organization accounts
- Groups

## Implementation

### Engineering Principles

- Build the smallest thing that works.
- If it needs documentation, simplify it.
- Readability outweighs abstraction.
- Prefer clarity over cleverness.
- Prefer the standard library over external dependencies.
- Distribution must pay for itself.
- Measure before optimizing.
- Experiment to learn.
- HTML is the API:
  - URLs name nouns.
  - Forms perform verbs.
  - Links show what can happen next.

### Stack

- Go as programming language
- PostgreSQL as database
- S3 storage for PDFs
- Tailwind CSS as GUI

### Canonical Resource Identifiers

| Resource                | Identifier                      | Description                       | Notes                    | Stability |
| ----------------------- | ------------------------------- | --------------------------------- | ------------------------ | --------- |
| Entry                   | `/`                             | Primary entry point               | Only required URL        | Canonical |
| User                    | `/{username}`                   | User profile / namespace          | Public identity          | Canonical |
| Stack                   | `/{username}/{stack}`           | Paper stack                       | Access-controlled        | Canonical |
| Stack settings          | `/{username}/{stack}/settings`  | Stack configuration view          | Owner-only               | Derived   |
| Stack activity          | `/{username}/{stack}/activity`  | Activity overview for the stack   | Read-only                | Derived   |
| Paper                   | `/papers/{paper-id}`            | Canonical verified paper          | UID                      | Canonical |
| Paper activity          | `/papers/{paper-id}/activity`   | Activity overview for the paper   | Read-only                | Derived   |
| Paper provenance        | `/papers/{paper-id}/provenance` | PDF sources & verification report | Read-only                | Derived   |
| Submission              | `/submissions/{submission-id}`  | Paper submitted for verification  | Stateful workflow object | Transient |
| Account                 | `/account`                      | Current user account              | Authenticated only       | Canonical |
| Subscription collection | `/account/subscriptions`        | Stacks the current user follows   | Authenticated            | Derived   |

## Backend

- Import meta data and PDF from databases
- Extract text and metadata from PDFs
- Match preprints/author versions and papers
