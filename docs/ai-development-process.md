# AI development process

I used AI to support planning and implementation. I reviewed and edited the output, tested the work, and made the final technical and design decisions.

Personal development skills, agent instructions, and internal reference lists are intentionally omitted below.

## Planning

### Prompt 1

Let's define the calculator behavior first. Propose three variations, and I'll choose one.

Calculator operations: addition, subtraction, multiplication, division, exponentiation, square root, and percentage (like a real calculator, but nothing too complex).

Design:

- Intuitive UI for entering input and displaying results
- Input validation and error handling
- Responsive design (mobile support)

#### My review

I chose my preferred option and requested `calculator-behaviour.md`. I reviewed and edited the document.

### Prompt 2

List websites I can use as design references or for a moodboard. Use my reference list and `calculator-behaviour.md`. If needed, search for other professional websites with ratings or endorsements from people or companies.

*Personal design reference list omitted.*

#### My review

I reviewed all the proposed websites and designs and selected a few to guide the design.

### Prompt 3

Create the layout of the buttons and display for the calculator. Make it a low-fidelity prototype.

#### My review

I reviewed and revised the prototype.

### Prompt 4

Create three high-fidelity design images so I can choose the final design. Use the supplied design reference.

*Internal design reference omitted.*

#### My review

I reviewed the images and selected one. I made changes and created `DESIGN.md`.

### Prompt 5

Create `prd.md` using the technical assessment details, `DESIGN.md`, and `calculator-behaviour.md`.

#### My review

I reviewed and edited the product requirements document.

### Prompt 6

Create `architecture.md` using `DESIGN.md`, `calculator-behaviour.md`, and `prd.md`.

#### My review

I reviewed and edited the architecture document.

### Prompt 7

Create `openapi.yaml` using `DESIGN.md`, `calculator-behaviour.md`, and `prd.md`. Use the supplied OpenAPI documentation references.

*Personal OpenAPI reference list omitted.*

#### My review

I reviewed and edited the API specification.

## Project initialization

- I created most files manually and consulted an AI agent on technical questions.
- I established the rules and development skills for the coding agent. Personal development skills and agent instructions omitted.
- I configured the local environment.

## Backend development

### Prompt 8

Use the relevant parts of the supplied files from a more complex project to develop the backend. Apply the supplied development skills.

*Personal development skills and source files from the other project omitted.*

#### My review

I reviewed and edited the backend implementation.

## Frontend development

### Prompt 9

Implement the frontend using the `docs/` folder, `DESIGN.md`, and the selected high-fidelity design image. Apply the supplied development skills.

*Personal development skills omitted.*

#### My review

I reviewed and edited the frontend implementation.

## Docker and documentation

### Prompt 10

Create a simple, production-quality Docker setup for this repository.

First inspect the existing frontend and backend configuration.

Use the supplied official documentation references as the primary source of truth.

*Personal documentation reference list omitted.*

Create:

- `frontend/Dockerfile`
- `frontend/.dockerignore`
- `backend/Dockerfile`
- `backend/.dockerignore`
- `compose.yaml` at the repository root
- An Nginx configuration for the frontend, only if needed

Requirements:

- Use multi-stage builds.
- Respect the Node, pnpm, and Go versions already pinned by the project.
- Frontend: build Vite with pnpm and serve `dist/` with Nginx, not `vite preview`.
- Backend: build the Go binary in a builder stage and use a minimal production runtime.
- Run production containers as non-root where practical.
- Have Nginx proxy `/calculate` and `/healthcheck` to `backend:4000`, so React can use relative API URLs.
- Publish only the frontend to the host unless backend exposure is needed.
- Use Compose's default network.
- Add sensible `.dockerignore` files.
- Do not add databases, volumes, custom networks, profiles, Docker Bake, Kubernetes, dev containers, or other infrastructure this project does not need.
- Do not refactor unrelated application code.

#### Outcome

I tested the Docker setup and confirmed it worked.

### Prompt 11

Review the whole project and improve code comments and documentation where needed.

Use my personal development skills and documentation guidelines where applicable. Keep application behavior unchanged.

Focus on:

* first-party code comments,
* README setup and project status,
* consistency between the implementation and files in `docs/`,
* removing outdated or misleading documentation.

Do not refactor unrelated code or introduce new features.

#### My review

I reviewed the resulting documentation and comment changes before keeping them.
