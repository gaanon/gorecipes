# Recipe Commenting Feature Plan

**Goal:** Allow users to add, view, edit, and delete comments on recipes. If a user does not provide an author name, it defaults to "Gopher".

## I. Data Model for Comments

A new data structure for comments:
*   `ID`: Unique identifier for the comment (e.g., UUID).
*   `RecipeID`: ID of the recipe this comment belongs to.
*   `AuthorName`: Name of the person who wrote the comment. Defaults to "Gopher" if submitted empty.
*   `Text`: Content of the comment (required).
*   `CreatedAt`: Timestamp (Unix epoch) of creation.
*   `UpdatedAt`: Timestamp (Unix epoch) of last modification.

## II. Backend API Endpoints

New API endpoints under recipes:
*   **`POST /api/v1/recipes/{recipe_id}/comments`**
    *   **Action:** Creates a new comment.
    *   **Request Body:** `{ "author_name": "string" (optional), "text": "string" (required) }`
    *   **Behavior:** If `author_name` is empty or whitespace, it's set to "Gopher".
    *   **Response:** `201 Created` with the new comment object.
*   **`GET /api/v1/recipes/{recipe_id}/comments`**
    *   **Action:** Retrieves all comments for a recipe, sorted by `CreatedAt`.
    *   **Response:** `200 OK` with an array of comment objects.
*   **`PUT /api/v1/recipes/{recipe_id}/comments/{comment_id}`**
    *   **Action:** Updates an existing comment's text.
    *   **Request Body:** `{ "text": "string" }`
    *   **Response:** `200 OK` with the updated comment object.
*   **`DELETE /api/v1/recipes/{recipe_id}/comments/{comment_id}`**
    *   **Action:** Deletes a comment.
    *   **Response:** `204 No Content` or `200 OK`.

## III. Database Interaction (BadgerDB)

*   **Key Structure:** `comment:<recipe_id>:<comment_id>`
    *   Example: `comment:recipe-uuid-123:comment-uuid-abc`
    *   Allows prefix scan for `comment:<recipe_id>:` to get all comments for a recipe.
*   **Value:** JSON-marshalled comment object.

## IV. Backend Implementation Sketch

1.  **New Go Model (`backend/internal/models/comment.go`):**
    ```go
    package models

    type Comment struct {
        ID         string `json:"id"`
        RecipeID   string `json:"recipe_id"`
        AuthorName string `json:"author_name"`
        Text       string `json:"text"`
        CreatedAt  int64  `json:"created_at"`
        UpdatedAt  int64  `json:"updated_at"`
    }
    ```
2.  **New Database Functions (`backend/internal/database/badgerdb.go`):**
    *   Functions for CRUD operations on comments.
3.  **New API Handlers (e.g., `backend/internal/handlers/comments.go`):**
    *   `CreateCommentForRecipe(c *gin.Context)`: Includes logic to default `AuthorName` to "Gopher".
    *   `GetCommentsForRecipe(c *gin.Context)`
    *   `UpdateComment(c *gin.Context)`
    *   `DeleteComment(c *gin.Context)`
4.  **Router Updates (`backend/internal/router/router.go`):**
    *   Map new routes to handlers.

## V. Frontend UI/UX (SvelteKit)

1.  **Recipe Detail Page (`frontend/src/routes/recipes/[id]/+page.svelte`):**
    *   Add a "Comments" section below the recipe method.
2.  **Displaying Comments:**
    *   Fetch comments via `GET /api/v1/recipes/{recipe_id}/comments` on page load.
    *   List comments showing `AuthorName`, `Text`, `CreatedAt`.
3.  **Adding a Comment:**
    *   Form with "Your Name" (optional) and "Comment" (textarea, required) fields.
    *   Submit triggers `POST` request. Update list optimistically or on success.
4.  **Editing a Comment:**
    *   "Edit" button per comment.
    *   Transforms comment display into an inline editing form.
    *   Save triggers `PUT` request.
5.  **Deleting a Comment:**
    *   "Delete" button per comment.
    *   Show confirmation prompt before `DELETE` request.
6.  **New Svelte Components:**
    *   `CommentCard.svelte`: Displays individual comment, handles edit/delete.
    *   `RecipeComments.svelte`: Manages comment list, "add comment" form, API interactions.

## VI. Workflow Diagram

```mermaid
graph TD
    subgraph Feature: Recipe Comments
        direction LR

        subgraph User Interaction (Frontend)
            U_ViewRecipe[Views Recipe Page] --> U_SeeCommentsSection[Sees Comments Section]
            U_SeeCommentsSection --> U_ViewComments[Views Existing Comments]
            U_SeeCommentsSection --> U_AddCommentForm[Uses Add Comment Form]
            U_AddCommentForm --> U_SubmitComment[Submits New Comment]
            U_ViewComments --> U_EditComment[Clicks Edit on a Comment]
            U_ViewComments --> U_DeleteComment[Clicks Delete on a Comment]
        end

        subgraph Frontend Logic (SvelteKit)
            PageLoad([`recipes/[id]/+page.svelte`]) --> FetchCommentsApiCall{GET /recipes/{id}/comments}
            AddForm([Add Comment Form]) --> AddCommentApiCall{POST /recipes/{id}/comments}
            EditAction[Edit Comment Action] --> UpdateCommentApiCall{PUT /recipes/{id}/comments/{cid}}
            DeleteAction[Delete Comment Action] --> DeleteCommentApiCall{DELETE /recipes/{id}/comments/{cid}}
            FetchCommentsApiCall --> DisplayComments[Display Comment List]
            DisplayComments --> CommentCard([`CommentCard.svelte`])
        end

        subgraph Backend Logic (Go/Gin)
            Router([`router.go`])
            Router -- routes --> CommentHandlers([`comments.go`])
            CommentHandlers -- uses --> CommentDBFuncs([`badgerdb.go` Comment Funcs])
            CommentDBFuncs -- interacts --> DB[(BadgerDB: comment:{rid}:{cid})]
            CommentHandlers -- uses --> CommentModel([`comment.go` Model])
            ApiRequest[POST /recipes/{id}/comments \n { author_name: "optional", text: "required" }] --> CommentHandlers
            CommentHandlers --> CheckAuthor{AuthorName Empty?}
            CheckAuthor -- Yes --> SetDefaultAuthor[Set AuthorName = "Gopher"]
            SetDefaultAuthor --> SaveComment[Save Comment to DB]
            CheckAuthor -- No --> SaveComment
            SaveComment --> ApiResponse[201 Created \n { id, recipe_id, author_name, ... }]

        end

        AddCommentApiCall --> Router
        FetchCommentsApiCall --> Router
        UpdateCommentApiCall --> Router
        DeleteCommentApiCall --> Router
    end
```

## VII. Simplifications for Initial Version
*   No user authentication/authorization for comments.
*   Basic error handling.