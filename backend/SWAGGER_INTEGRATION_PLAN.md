# Swagger/OpenAPI Integration Plan for GoRecipes Backend

This document outlines the detailed plan for integrating Swagger/OpenAPI documentation into the existing GoRecipes backend application.

### **1. Chosen Swagger/OpenAPI Package: `swag`**

For integrating Swagger/OpenAPI documentation into the existing Go backend application, the `swag` package (github.com/swaggo/swag) is the primary choice for documentation generation.

**Reasoning:**
*   **Annotation-based:** `swag` generates OpenAPI documentation directly from Go source code annotations, making it easy to integrate into existing projects without significant refactoring.
*   **Popularity and Community Support:** It's widely used and well-maintained, with good community support.
*   **Integration with Routers:** It integrates seamlessly with popular Go web frameworks like Gin, which is now confirmed to be the underlying framework for the router.

### **2. Integration Steps**

The integration will involve three main phases: **Installation**, **Documentation Generation**, and **Serving the UI**.

#### **Phase 1: Installation**

1.  **Install `swag` CLI:**
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
2.  **Install `gin-swagger`:**
    ```bash
    go get -u github.com/swaggo/gin-swagger
    go get -u github.com/gin-gonic/gin
    ```

#### **Phase 2: Documentation Generation**

`swag` generates the `docs` directory containing `docs.go`, `swagger.json`, and `swagger.yaml`.

1.  **Add General API Information (in `main.go` or a dedicated `docs.go` file):**
    Add annotations at the package level (e.g., in `backend/cmd/server/main.go`) to define global API information like title, version, description, host, and base path.

    ```go
    // @title GoRecipes API
    // @version 1.0
    // @description This is the API documentation for the GoRecipes application.
    // @host localhost:8080 // Adjust based on your application's host and port
    // @BasePath /api/v1 // Adjust based on your API's base path
    // @schemes http https
    // @securityDefinitions.apikey ApiKeyAuth
    // @in header
    // @name Authorization
    // @externalDocs.description OpenAPI
    // @externalDocs.url https://swagger.io/resources/open-api/
    package main
    ```

2.  **Annotate API Endpoints (in handler files):**
    For each API endpoint in files like `backend/internal/handlers/*.go`, add `swag` annotations to define:
    *   HTTP method (`@Router`)
    *   Path (`@Router`)
    *   Summary (`@Summary`)
    *   Description (`@Description`)
    *   Parameters (`@Param`)
    *   Responses (`@Success`, `@Failure`)
    *   Security (`@Security`)

    **Example (for a recipe handler in `backend/internal/handlers/recipes.go`):**

    ```go
    // @Summary Get all recipes
    // @Description Get a list of all recipes
    // @Tags recipes
    // @Accept json
    // @Produce json
    // @Success 200 {array} models.Recipe "Successfully retrieved recipes"
    // @Failure 500 {object} map[string]string "Internal Server Error"
    // @Router /recipes [get]
    func GetRecipes(c *gin.Context) { // Note: Changed to gin.Context
        // ... existing handler logic
    }
    ```

3.  **Generate Documentation:**
    Run the `swag init` command from the root of the `backend` directory. This will generate the `docs` directory.
    ```bash
    cd backend
    swag init
    ```

#### **Phase 3: Serving the Swagger UI**

1.  **Import Generated Docs:**
    In `backend/cmd/server/main.go`, import the generated `docs` package:
    ```go
    import _ "gorecipes/backend/docs" // Adjust import path based on your module name
    ```

2.  **Integrate Swagger UI into Router:**
    Modify `backend/internal/router/router.go` (or `main.go` if the router is defined there) to serve the Swagger UI using `gin-swagger`.

    **Example (assuming Gin router):**

    ```go
    package router

    import (
        "github.com/gin-gonic/gin"
        swaggerFiles "github.com/swaggo/files"
        ginSwagger "github.com/swaggo/gin-swagger"
        _ "gorecipes/backend/docs" // Import generated docs
    )

    func SetupRouter() *gin.Engine { // Note: Changed return type to *gin.Engine
        r := gin.Default()

        // ... existing routes

        // Swagger UI route
        r.GET("/swagger/*any", ginSwagger.WrapHandler(ginSwagger.URL("http://localhost:8080/swagger/doc.json"), swaggerFiles.Handler))

        return r
    }
    ```
    **Note:** The `ginSwagger.WrapHandler` automatically handles serving `swagger.json` and `swagger.yaml` from the `docs` directory.

### **3. High-Level Implementation Plan**

Here's a high-level plan for the implementation:

1.  **Modify `backend/cmd/server/main.go`:**
    *   Add the global API information annotations at the package level.
    *   Import the generated `docs` package: `_ "gorecipes/backend/docs"`.

2.  **Modify `backend/internal/handlers/*.go` files:**
    *   Add `swag` annotations to each API handler function to describe the endpoint, parameters, and responses. This will be an iterative process, adding annotations as needed for each endpoint.
    *   Ensure handler function signatures are compatible with Gin (e.g., `func GetRecipes(c *gin.Context)`).

3.  **Modify `backend/internal/router/router.go`:**
    *   Change the router initialization and return type to `*gin.Engine`.
    *   Add the `gin-swagger` handler to serve the Swagger UI at `/swagger/*any`.

4.  **Update `go.mod` and `go.sum`:**
    *   Run `go mod tidy` after installing `swag` and `gin-swagger` to update dependencies.

5.  **Add `swag init` to build process (e.g., `Dockerfile` or `Makefile`):**
    *   Ensure `swag init` is run before building the application to generate the latest documentation. This could be added to the `Dockerfile` or a build script.

### **Considerations for API Versioning and Authentication**

*   **API Versioning:**
    *   **Path-based versioning:** If using path-based versioning (e.g., `/api/v1/recipes`), ensure the `@BasePath` annotation reflects this.
    *   **Header-based versioning:** If using header-based versioning, document the required header using `@Param` annotations.
*   **Authentication:**
    *   **API Key:** The example above includes `@securityDefinitions.apikey ApiKeyAuth` and `@security ApiKeyAuth` annotations. This defines an API key security scheme and applies it to an endpoint.
    *   **OAuth2/JWT:** For more complex authentication schemes like OAuth2 or JWT, `swag` supports `@securityDefinitions.oauth2` and `@securityDefinitions.bearer` annotations, which would be defined globally and then applied to specific endpoints.

### **Mermaid Diagram: High-Level Flow**

```mermaid
graph TD
    A[Go Backend Application] --> B(API Endpoints in Handlers)
    B -- Annotate with swag --> C{swag CLI}
    C -- Generates --> D[docs/ Directory]
    D -- Contains --> D1[docs.go]
    D -- Contains --> D2[swagger.json]
    D -- Contains --> D3[swagger.yaml]
    A -- Imports docs.go --> E[main.go]
    E -- Uses gin-swagger --> F[Router Configuration (Gin)]
    F -- Serves --> G[Swagger UI]
    G -- Fetches API Spec from --> F