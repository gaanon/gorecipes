# Recipe Manager Webapp Design Plan

### 1. Overview

The project is a web application for managing cooking recipes. Users will be able to list, add, edit, and delete recipes. Recipes can be filtered by tags, where every ingredient automatically becomes a tag. Each recipe will display its ingredients, cooking method, and an optional photo (a generic one will be fetched and stored if not provided by the user).

### 2. Technology Stack

*   **Backend:** Go with the Gin framework.
*   **Frontend:** Svelte with SvelteKit (for routing and a more structured application).
*   **Database:** BadgerDB (embedded key-value store, lightweight and Go-native).
*   **Image Fetching:** A generic image search API (e.g., a free tier of Google Custom Search API, or a library that scrapes public domain image sites. We'll need to be mindful of API terms of service and rate limits).

### 3. Database Design (BadgerDB)

BadgerDB is a key-value store. We'll need to define how we structure our data.

*   **Recipes:**
    *   Each recipe will be stored as a JSON object.
    *   The key could be a unique ID (e.g., UUID) prefixed with `recipe:`. Example: `recipe:uuid-1234-abcd`.
    *   The value will be the JSON string of the recipe object:
        ```json
        {
          "id": "uuid-1234-abcd",
          "name": "Spaghetti Carbonara",
          "ingredients": ["Spaghetti", "Guanciale", "Eggs", "Pecorino Romano", "Black Pepper"],
          "method": "1. Cook spaghetti. 2. Fry guanciale. 3. Mix eggs and cheese. 4. Combine all.",
          "photo_filename": "carbonara.jpg", // or "fetched_image_uuid-5678.jpg"
          "created_at": "2025-05-15T15:00:00Z",
          "updated_at": "2025-05-15T15:00:00Z"
        }
        ```
*   **Ingredients (for autocomplete and tagging):**
    *   We can maintain a set of all unique ingredients.
    *   Key prefix: `ingredient:`. Example: `ingredient:Spaghetti`.
    *   The value could be simple, like `1` or `true`, just to mark its existence, or we could store a list of recipe IDs that use this ingredient if we want to optimize tag-based lookups without iterating all recipes. For simplicity initially, just marking existence is fine.
    *   Alternatively, for tag-based filtering, we might iterate through all recipes and check their `ingredients` array. Given BadgerDB's speed for scans, this might be acceptable for a moderate number of recipes. If performance becomes an issue, we can introduce secondary indexing (e.g., an ingredient key pointing to a list of recipe IDs).
*   **Images:**
    *   User-uploaded images and fetched images will be stored on the server's file system. The `photo_filename` in the recipe JSON will point to this file.

### 4. API Endpoints (Backend - Gin)

All endpoints will be under a base path like `/api/v1`.

*   **Recipes:**
    *   `GET /recipes`: List all recipes.
        *   Optional query parameter `tags` (comma-separated) to filter by ingredients.
    *   `POST /recipes`: Create a new recipe.
        *   Request body: JSON with recipe data (name, ingredients, method). Photo can be uploaded via multipart/form-data.
        *   If no photo is provided, the backend will trigger an image fetch.
    *   `GET /recipes/:id`: Get a specific recipe by ID.
    *   `PUT /recipes/:id`: Update an existing recipe.
        *   Request body: JSON with recipe data.
    *   `DELETE /recipes/:id`: Delete a recipe.
*   **Ingredients (for autocomplete):**
    *   `GET /ingredients?q=<query>`: Get a list of ingredient names matching the query for autocomplete.
*   **Images:**
    *   `POST /recipes/:id/image`: (Alternative to including in `POST /recipes`) Upload an image for a specific recipe.
    *   Images will likely be served statically by Gin from a dedicated folder (e.g., `/uploads/images/:filename`).

### 5. Frontend Components (Svelte/SvelteKit)

*   **Layouts:**
    *   `+layout.svelte`: Main layout (e.g., navbar, footer).
*   **Routes (Pages):**
    *   `/ (routes/+page.svelte)`: Main page.
        *   Displays `RecipeList`.
        *   Button/link to "Create New Recipe" (`/recipes/new`).
    *   `/recipes/new (routes/recipes/new/+page.svelte)`: Page for creating a new recipe.
        *   Uses `RecipeForm` component.
    *   `/recipes/[id] (routes/recipes/[id]/+page.svelte)`: Recipe detail page.
        *   Displays `RecipeDetail` component.
    *   `/recipes/[id]/edit (routes/recipes/[id]/edit/+page.svelte)`: Page for editing an existing recipe.
        *   Uses `RecipeForm` component, pre-filled with recipe data.
*   **Components (src/lib/components):**
    *   `RecipeList.svelte`: Fetches and displays a list of recipes (cards or list items).
        *   Includes filtering UI (e.g., tag selection/search).
    *   `RecipeCard.svelte`: Displays a single recipe summary in the list.
    *   `RecipeDetail.svelte`: Displays full details of a recipe.
    *   `RecipeForm.svelte`: Form for creating/editing recipes.
        *   Input fields for name, ingredients (with autocomplete), method.
        *   File input for photo.
    *   `TagInput.svelte`: Reusable component for inputting ingredients with autocomplete suggestions.
    *   `ImageUploader.svelte`: Component to handle image preview and upload.

### 6. Key Features Breakdown

*   **CRUD for Recipes:**
    *   **Create:** `RecipeForm` submits to `POST /api/v1/recipes`.
    *   **Read (List):** `RecipeList` fetches from `GET /api/v1/recipes`.
    *   **Read (Detail):** `RecipeDetail` fetches from `GET /api/v1/recipes/:id`.
    *   **Update:** `RecipeForm` (in edit mode) submits to `PUT /api/v1/recipes/:id`.
    *   **Delete:** Button in `RecipeDetail` or `RecipeList` calls `DELETE /api/v1/recipes/:id`.
*   **Tagging and Filtering:**
    *   Ingredients in the `ingredients` array of a recipe are the tags.
    *   Backend `GET /recipes` endpoint will support a `?tags=foo,bar` query parameter.
    *   Frontend `RecipeList` will have UI elements (e.g., a multi-select dropdown, or clickable tags) to build this query.
*   **Main Page:**
    *   Lists all recipes using `RecipeList`.
    *   "Create New Recipe" button navigates to `/recipes/new`.
*   **Recipe Detail Page:**
    *   Displays name, ingredients (as a list, potentially clickable as tags), cooking method, and photo.
*   **Photo Handling:**
    *   `RecipeForm` includes an `<input type="file">`.
    *   If a photo is uploaded, it's sent to the backend with the recipe data. Backend saves it to a designated folder (e.g., `./uploads/images/`) and stores the filename.
    *   If no photo is uploaded, the backend, after saving the recipe, attempts to fetch an image using the recipe name or primary ingredients from a generic image search API.
    *   The fetched image is then saved (e.g., `./uploads/images/fetched_RECIPE_ID.jpg`) and its filename is associated with the recipe.
    *   Images are served statically by Gin.
*   **Recipe Structure:** As defined in the Database Design section.
*   **Ingredient Input with Autocomplete:**
    *   `TagInput.svelte` component.
    *   As the user types, it queries `GET /api/v1/ingredients?q=<typed_text>`.
    *   Backend searches BadgerDB for keys prefixed with `ingredient:` that match the query.
    *   Frontend displays suggestions. When an ingredient is added, it's added to a list for the current recipe.

### 7. Project Structure (Suggested)

```
gorecipes/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         // Main application entry point
│   ├── internal/
│   │   ├── config/             // Configuration loading
│   │   ├── database/           // BadgerDB interaction logic
│   │   │   └── badgerdb.go
│   │   ├── handlers/           // Gin HTTP handlers (recipes, ingredients)
│   │   │   ├── recipes.go
│   │   │   └── ingredients.go
│   │   ├── models/             // Go structs for recipes, etc.
│   │   │   └── recipe.go
│   │   ├── services/           // Business logic (e.g., image fetching)
│   │   │   └── image_fetcher.go
│   │   └── router/             // Gin router setup
│   │       └── router.go
│   ├── go.mod
│   ├── go.sum
│   └── uploads/                // Stored images (add to .gitignore if not versioned)
│       └── images/
├── frontend/ (SvelteKit project)
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/     // Reusable Svelte components
│   │   │   │   ├── RecipeList.svelte
│   │   │   │   ├── RecipeCard.svelte
│   │   │   │   ├── RecipeDetail.svelte
│   │   │   │   ├── RecipeForm.svelte
│   │   │   │   ├── TagInput.svelte
│   │   │   │   └── ImageUploader.svelte
│   │   │   └── services/       // API call functions
│   │   │       └── api.js
│   │   ├── routes/             // SvelteKit page routes
│   │   │   ├── +page.svelte    // Home page
│   │   │   ├── recipes/
│   │   │   │   ├── new/
│   │   │   │   │   └── +page.svelte
│   │   │   │   ├── [id]/
│   │   │   │   │   ├── +page.svelte
│   │   │   │   │   └── edit/
│   │   │   │   │       └── +page.svelte
│   │   │   └── +layout.svelte  // Main layout
│   │   ├── app.html
│   ├── static/                 // Static assets
│   ├── svelte.config.js
│   ├── package.json
│   └── vite.config.js
├── .gitignore
└── README.md
```

### 8. Deployment Considerations

*   **Backend:** The Go application can be compiled into a single binary.
*   **Frontend:** The SvelteKit app can be built into static assets or run with a Node.js adapter.
*   **Containerization:** Both backend (with the BadgerDB data directory and uploads) and frontend can be containerized using Docker. A `docker-compose.yml` can manage both services. BadgerDB data would be persisted using a Docker volume.

### 9. Mermaid Diagrams

#### a. Component Interaction (Simplified Frontend)

```mermaid
graph TD
    A[User] --> B[/ (HomePage)];
    B -- Clicks 'New Recipe' --> D[/recipes/new (RecipeFormPage)];
    B -- Views List --> E[RecipeList];
    E -- Clicks Recipe --> F[/recipes/:id (RecipeDetailPage)];
    F -- Clicks 'Edit' --> G[/recipes/:id/edit (RecipeFormPage)];
    D -- Submits Form --> H{API Call: Create Recipe};
    G -- Submits Form --> I{API Call: Update Recipe};
    F -- Clicks 'Delete' --> J{API Call: Delete Recipe};

    subgraph "Svelte Components"
        E
        F
        D
        G
    end
```

#### b. Backend Request Flow (Example: Create Recipe)

```mermaid
graph TD
    User --> FE[Frontend: RecipeForm];
    FE -- Submits Data (Name, Ingredients, Method, Optional Photo) --> API[Gin API: POST /api/v1/recipes];
    API --> Handler[Recipe Handler];
    Handler --> Validation[Validate Data];
    Validation -- Valid --> DBLogic[Database Logic];
    DBLogic -- Save Recipe --> BadgerDB[(BadgerDB)];
    alt No Photo Provided
        Handler --> ImgService[Image Fetching Service];
        ImgService -- Fetch Image --> ExtAPI[External Image API];
        ExtAPI -- Image Data --> ImgService;
        ImgService -- Save Image --> FileSystem[(Server Filesystem)];
        ImgService -- Update Recipe with Image Filename --> DBLogic;
    end
    alt Photo Provided
        Handler -- Save Image --> FileSystem[(Server Filesystem)];
        Handler -- Update Recipe with Image Filename --> DBLogic;
    end
    DBLogic -- Success --> Handler;
    Handler -- Recipe Data (incl. ID, Photo Filename) --> API;
    API -- JSON Response --> FE;
```

### 10. Next Steps

1.  **Setup Project Structure:** Create the main directories and initialize Go modules and SvelteKit project.
2.  **Backend Development:**
    *   Implement BadgerDB wrapper/service.
    *   Define recipe models.
    *   Create Gin handlers for CRUD operations on recipes.
    *   Implement ingredient autocomplete endpoint.
    *   Implement image upload and basic image fetching logic.
3.  **Frontend Development:**
    *   Set up SvelteKit routing.
    *   Develop Svelte components for listing, viewing, creating, and editing recipes.
    *   Integrate API calls.
4.  **Styling:** Apply CSS for a pleasant user experience.
5.  **Testing:** Unit tests for backend logic and potentially E2E tests for frontend interactions.