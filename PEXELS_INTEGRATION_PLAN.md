# Plan: Pexels API Integration for Automatic Recipe Image Fetching

## Objective
To automatically fetch a relevant image for a new recipe from the Pexels API if the user does not provide one during creation. If the Pexels fetch fails or is not configured, a default placeholder image will be used.

## 1. Prerequisites & Configuration

*   **Pexels API Key**:
    *   Obtain an API key from [https://www.pexels.com/api/](https://www.pexels.com/api/).
    *   The key must be kept confidential.
*   **Environment Variable for API Key**:
    *   The Go backend application will expect the Pexels API key to be available as an environment variable named `PEXELS_API_KEY`.
    *   Set this variable in your local development environment (e.g., in the terminal or via a `.env` file mechanism) and in the production deployment environment.
*   **Default Placeholder Image**:
    *   A default image (e.g., `placeholder.jpg` or `placeholder.png`) must be created and placed in the `backend/uploads/images/` directory. This image will serve as the ultimate fallback.

## 2. Backend Implementation (Go)

### 2.1. New Helper Function

A new Go function will be created, likely within `backend/internal/handlers/recipes.go` or a dedicated utility file.

**Signature:**
```go
func fetchAndSaveImageFromPexels(query string, recipeID string, apiKey string) (savedFilename string, err error)
```

**Functionality:**
1.  **Construct Pexels API Request**:
    *   Target URL: `https://api.pexels.com/v1/search`
    *   Parameters: `query` (e.g., recipe name), `per_page=1` (to get one image).
2.  **Execute HTTP GET Request**:
    *   Use `net/http` package.
    *   Include the `Authorization` header with the `apiKey`.
    *   Handle potential HTTP errors (non-200 status codes).
3.  **Parse JSON Response**:
    *   Use `encoding/json` to unmarshal the Pexels API response.
    *   The response contains a `photos` array. We are interested in the first photo (`photos[0]`).
4.  **Extract Image URL**:
    *   From the photo object, extract a suitable image URL (e.g., `photo.src.large` or `photo.src.original`). Check Pexels API documentation for the best field.
    *   If no photos are found or the URL is missing, return an error.
5.  **Download Image**:
    *   Make another HTTP GET request to the extracted image URL.
    *   Handle potential download errors.
6.  **Determine File Extension**:
    *   Attempt to determine the image type from the `Content-Type` header of the downloaded image response (e.g., `image/jpeg`, `image/png`).
    *   Alternatively, parse the extension from the image URL if reliable.
7.  **Generate Unique Filename**:
    *   Create a filename, e.g., `recipeID + "_pexels" + .jpg` (or appropriate extension).
8.  **Save Image**:
    *   Save the downloaded image data to the `backend/uploads/images/` directory using the generated filename.
    *   Ensure the `uploads/images/` directory exists (create if not).
    *   Use `os` and `io` packages for file operations.
9.  **Return Result**:
    *   If successful, return the `savedFilename` and `nil` error.
    *   If any step fails, return an empty `savedFilename` and a descriptive `error`. Log detailed errors internally.

### 2.2. Modify `CreateRecipe` Handler

The existing `CreateRecipe` handler in `backend/internal/handlers/recipes.go` will be updated:

1.  **Existing User Upload Logic**: The current logic for handling photos uploaded directly by the user (`c.FormFile("photo")`) remains the priority. If a user uploads a file, it's processed as before, and the Pexels fetch is skipped.
2.  **No User Upload Scenario**: If `c.FormFile("photo")` returns `http.ErrMissingFile`:
    a.  **Retrieve `PEXELS_API_KEY`**: Read the API key from the environment variable.
    b.  **Check API Key**: If the key is available and not an empty string:
        i.  **Call `fetchAndSaveImageFromPexels`**:
            *   Use `recipe.Name` as the primary `query`.
            *   Pass the `recipe.ID` and the `apiKey`.
        ii. **Handle Success**: If `fetchAndSaveImageFromPexels` returns a `savedFilename` and no error, assign this filename to `recipe.PhotoFilename`.
        iii. **Handle Failure**: If `fetchAndSaveImageFromPexels` returns an error or an empty filename, log the error (e.g., "Pexels fetch failed for recipe [ID]: [error details]") and proceed to use the default placeholder image (`recipe.PhotoFilename = "placeholder.jpg"`).
    c.  **API Key Not Configured**: If `PEXELS_API_KEY` is not set or is empty, log this (e.g., "Pexels API key not configured, using placeholder.") and set `recipe.PhotoFilename = "placeholder.jpg"`.
3.  **Other File Errors**: If `c.FormFile("photo")` returns an error *other than* `http.ErrMissingFile` (e.g., an issue with processing an uploaded file), this indicates a problem with the user's upload. The handler should return an appropriate error to the client (e.g., `StatusBadRequest`), and the Pexels fetch logic should typically be skipped.

### 2.3. Error Handling and Logging
*   Log informative messages for successful Pexels fetches.
*   Log detailed errors if Pexels API calls, image downloads, or file saving operations fail. These internal errors should generally not fail the entire recipe creation process but should result in the use of the placeholder image.

## 3. Frontend Impact
*   No direct changes are required on the frontend SvelteKit application for this feature to work. The frontend will continue to display images based on the `photo_filename` field returned by the backend API for each recipe.

## 4. Testing Scenarios
*   Recipe creation **with** user-uploaded image.
*   Recipe creation **without** user-uploaded image, **with** `PEXELS_API_KEY` configured:
    *   Pexels successfully finds and downloads an image.
    *   Pexels does not find an image for the query (fallback to placeholder).
    *   Pexels API returns an error (e.g., rate limit, invalid key) (fallback to placeholder).
    *   Image download from Pexels URL fails (fallback to placeholder).
*   Recipe creation **without** user-uploaded image, **without** `PEXELS_API_KEY` configured (fallback to placeholder).
*   Ensure the `placeholder.jpg` is correctly used in fallback scenarios.
*   Verify correct file naming and storage in `uploads/images/`.

## Visual Flow
```mermaid
graph TD
    A[User Submits Create Recipe Form] --> B{Photo Field Provided?};
    B -- Yes --> C[Process User-Uploaded Photo];
    C --> S[Save Recipe with User's PhotoFilename];
    B -- No --> D{Pexels API Key Configured?};
    D -- No --> P1[Log: Pexels Key Missing];
    P1 --> Q[Set PhotoFilename = "placeholder.jpg"];
    D -- Yes --> E[Call fetchAndSaveImageFromPexels(recipeName)];
    E -- Success --> F[Fetched Image Saved];
    F --> G[Set PhotoFilename = FetchedImageFilename];
    E -- Failure (API Error/No Image/Download Error) --> P2[Log: Pexels Fetch Failed];
    P2 --> Q;
    G --> S2[Save Recipe with Fetched PhotoFilename];
    Q --> S3[Save Recipe with Placeholder PhotoFilename];
    S --> Z[Return Success Response];
    S2 --> Z;
    S3 --> Z;