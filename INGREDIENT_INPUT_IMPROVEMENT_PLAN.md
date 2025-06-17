# Plan for "One Ingredient Per Line" Feature

**Goal:** Modify the recipe creation process to allow users to input ingredients one per line, instead of comma-separated.

**Impact Assessment:**
*   **Frontend (`frontend/src/routes/recipes/new/+page.svelte`):**
    *   The input field for ingredients needs to change from a single-line text input to a multi-line textarea.
    *   The `form-hint` text needs to be updated to reflect the new input method.
    *   The logic for preparing the `ingredientsStr` before sending it to the backend needs to be updated to send the raw textarea content.
*   **Backend (`backend/internal/handlers/recipes.go`):**
    *   The `CreateRecipe` handler needs to change how it parses the incoming `ingredientsStr`. Instead of splitting by commas, it should split by newlines.
    *   The `UpdateRecipe` handler also processes ingredients in a similar way, so it will need the same modification.
    *   The Swagger documentation comments should be updated to reflect the new input format.

**Detailed Steps:**

1.  **Frontend Modifications (`frontend/src/routes/recipes/new/+page.svelte`)**
    *   Change the `<input type="text" ...>` for ingredients to `<textarea ...>`.
    *   Remove the `placeholder` attribute and `form-hint` related to "comma-separated".
    *   The `on:submit` handler will send the raw `ingredientsStr` (multi-line text) directly.

2.  **Backend Modifications (`backend/internal/handlers/recipes.go`)**
    *   In the `CreateRecipe` function:
        *   Change `strings.Split(ingredientsStr, ",")` to `strings.Split(ingredientsStr, "\n")`.
        *   The existing `strings.TrimSpace` and `uniqueIngredients` logic will handle empty lines and whitespace.
    *   In the `UpdateRecipe` function:
        *   Apply the same change: `strings.Split(ingredientsStr, ",")` to `strings.Split(ingredientsStr, "\n")`.
    *   Update Swagger comments:
        *   Modify `@Param ingredients formData string false "Comma-separated list of ingredients"` to `"Newline-separated list of ingredients"` in both `CreateRecipe` and `UpdateRecipe` handler comments.

**Mermaid Diagram:**

```mermaid
graph TD
    A[User enters new recipe] --> B{Frontend: new/+page.svelte};
    B -- Ingredients input as multi-line text --> C[Frontend: Form Submission];
    C -- formData.append('ingredients', ingredientsStr) --> D[Backend: CreateRecipe/UpdateRecipe handler];
    D -- c.PostForm("ingredients") --> E[Backend: Parse ingredientsStr];
    E -- Split by newline character --> F[Backend: Process individual ingredients];
    F -- Save to Database --> G[Recipe created/updated];