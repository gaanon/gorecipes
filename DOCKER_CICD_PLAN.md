# Dockerization and CI/CD Plan for GoRecipes

**Goal:**
Automate the building of Docker images for the Go backend and SvelteKit frontend, and push them to Docker Hub on every push to the `main` branch. Data for BadgerDB and uploaded images will be managed using Docker named volumes.

**Summary of Configuration:**
*   **Docker Hub Username:** `gaanon`
*   **Backend Image:** `gaanon/gorecipes-backend`
*   **Frontend Image:** `gaanon/gorecipes-frontend`
*   **Tags:** `latest` and `sha-<short_commit_sha>`
*   **SvelteKit Adapter:** `adapter-static`
*   **Data Persistence:** Docker named volumes for BadgerDB data and uploaded images.
*   **GitHub Actions Trigger:** On push to the `main` branch.
*   **GitHub Secrets:** `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN` (to be set up by the user).

---

## Phase 1: Dockerization

### 1. Backend Dockerfile (`backend/Dockerfile`)
*   **Multi-stage build:**
    *   **Build Stage:** Use a Go base image (e.g., `golang:1.22-alpine`) to build the Go binary.
        *   Copy `go.mod` and `go.sum`, download dependencies.
        *   Copy the rest of the backend source code.
        *   Build the application (`CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/gorecipes-backend ./cmd/server/main.go`).
    *   **Runtime Stage:** Use a minimal base image (e.g., `alpine:latest` or `gcr.io/distroless/static-debian11`).
        *   Copy the compiled binary from the build stage.
        *   Set the working directory (e.g., `/app`).
        *   Define the default command to run the binary (e.g., `CMD ["/app/gorecipes-backend"]`).
        *   Expose the backend port (e.g., `8080`).
*   **Data Volume Paths:** The application will use `/app/data/badgerdb` for BadgerDB and `/app/uploads/images` for uploads. These paths will be targeted by Docker volumes.

### 2. Frontend Dockerfile (`frontend/Dockerfile`)
*   **Multi-stage build:**
    *   **Build Stage:** Use a Node.js base image (e.g., `node:20-alpine`) to build the SvelteKit static assets.
        *   Set working directory (e.g., `/app/frontend`).
        *   Copy `package.json`, `package-lock.json`.
        *   Install dependencies (`npm install`).
        *   Copy the rest of the frontend source code.
        *   Build the static site (`npm run build`).
    *   **Runtime Stage:** Use a lightweight web server image (e.g., `nginx:stable-alpine`).
        *   Copy the static assets from the SvelteKit build output (e.g., `build/`) into Nginx's HTML directory (e.g., `/usr/share/nginx/html`).
        *   Copy a custom Nginx configuration file (`nginx.conf`) to handle SvelteKit's SPA routing and proxy API requests.
        *   Expose the Nginx port (e.g., `80`).

### 3. Nginx Configuration (`frontend/nginx.conf`)
*   Configure Nginx to serve the SvelteKit static files from `/usr/share/nginx/html`.
*   Set up a `try_files` directive (e.g., `try_files $uri $uri/ /index.html;`) to ensure all routes are directed to `index.html` for SPA behavior.
*   Configure a `location /api/` block to proxy requests to the backend service (e.g., `proxy_pass http://gorecipes-backend:8080/api/;`). The hostname `gorecipes-backend` will be resolvable when using Docker Compose or a Docker network.

---

## Phase 2: GitHub Actions Workflow

### 1. Workflow File (`.github/workflows/docker-publish.yml`)
*   **Name:** "Build and Push Docker Images"
*   **Trigger:**
    ```yaml
    on:
      push:
        branches:
          - main
    ```
*   **Jobs:**
    *   **`build_and_push_backend` Job:**
        *   Runs on `ubuntu-latest`.
        *   **Steps:**
            1.  `actions/checkout@v4`: Checkout code.
            2.  `docker/setup-qemu-action@v3`: Set up QEMU.
            3.  `docker/setup-buildx-action@v3`: Set up Docker Buildx.
            4.  `docker/login-action@v3`: Login to Docker Hub using `secrets.DOCKERHUB_USERNAME` and `secrets.DOCKERHUB_TOKEN`.
            5.  `docker/metadata-action@v5`: Extract metadata (tags: `latest`, `sha-<short_commit_sha>`; image name: `gaanon/gorecipes-backend`).
            6.  `docker/build-push-action@v5`: Build and push the backend image.
                *   `context: ./backend`
                *   `file: ./backend/Dockerfile`
                *   `push: true`
                *   `tags: ${% raw %}{{ steps.meta.outputs.tags }}{% endraw %}`
                *   `labels: ${% raw %}{{ steps.meta.outputs.labels }}{% endraw %}`
    *   **`build_and_push_frontend` Job:**
        *   Runs on `ubuntu-latest`.
        *   Similar steps to the backend job, but for the frontend:
            1.  Checkout code.
            2.  Set up QEMU.
            3.  Set up Buildx.
            4.  Login to Docker Hub.
            5.  Extract metadata for `gaanon/gorecipes-frontend`.
            6.  Build and push the frontend image.
                *   `context: ./frontend`
                *   `file: ./frontend/Dockerfile`
                *   `push: true`
                *   `tags: ${% raw %}{{ steps.meta.outputs.tags }}{% endraw %}`
                *   `labels: ${% raw %}{{ steps.meta.outputs.labels }}{% endraw %}`

---

## Phase 3: Local Development and Deployment (Recommendations)

### 1. `.dockerignore` files:
*   Create `.dockerignore` in `./backend` (e.g., to ignore `.git`, `tmp/`, local dev files, `uploads/`, `data/`).
*   Create `.dockerignore` in `./frontend` (e.g., to ignore `.git`, `node_modules/`, `build/`, `.svelte-kit/`).

### 2. `docker-compose.yml` (for local development & testing):
*   Define services for `backend` and `frontend`.
*   **`backend` service:**
    *   Build from `./backend/Dockerfile`.
    *   Image name (optional for local): `gorecipes-backend-dev`.
    *   Ports: e.g., `8080:8080`.
    *   Volumes:
        *   `gorecipes_badger_data:/app/data/badgerdb`
        *   `gorecipes_uploads:/app/uploads/images`
*   **`frontend` service:**
    *   Build from `./frontend/Dockerfile`.
    *   Image name (optional for local): `gorecipes-frontend-dev`.
    *   Ports: e.g., `5173:80` (mapping host 5173 to container Nginx port 80).
    *   Depends on: `backend`.
*   **`volumes` top-level key:**
    *   `gorecipes_badger_data: {}`
    *   `gorecipes_uploads: {}`

---

## Workflow Diagram

```mermaid
graph TD
    A[Push to main branch on GitHub] --> B{GitHub Actions Workflow};
    B --> C[Job: Build & Push Backend];
    B --> D[Job: Build & Push Frontend];

    C --> E[1. Checkout Code];
    C --> F[2. Setup QEMU & Buildx];
    C --> G[3. Login to Docker Hub];
    C --> H[4. Extract Metadata (gaanon/gorecipes-backend:latest, gaanon/gorecipes-backend:sha-xxxx)];
    C --> I[5. Build backend/Dockerfile];
    C --> J[6. Push Image to Docker Hub];

    D --> K[1. Checkout Code];
    D --> L[2. Setup QEMU & Buildx];
    D --> M[3. Login to Docker Hub];
    D --> N[4. Extract Metadata (gaanon/gorecipes-frontend:latest, gaanon/gorecipes-frontend:sha-xxxx)];
    D --> O[5. Build frontend/Dockerfile];
    D --> P[6. Push Image to Docker Hub];

    J --> Q([Docker Hub: gaanon/gorecipes-backend]);
    P --> R([Docker Hub: gaanon/gorecipes-frontend]);

    subgraph Dockerization
        S[backend/Dockerfile]
        T[frontend/Dockerfile]
        U[frontend/nginx.conf]
    end

    subgraph Local Development / Deployment
        V[docker-compose.yml]
        W[Named Volume: BadgerDB Data]
        X[Named Volume: Uploaded Images]
    end
    V --> W;
    V --> X;
    Q -.-> V;
    R -.-> V;
```

---

**Next Steps for User:**
1.  Set up the `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets in the GitHub repository settings. The token should be a Docker Hub Personal Access Token with read/write permissions.