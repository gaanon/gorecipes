# Recipe Photo Processing Feature

This feature allows users to upload a photo of a recipe (e.g., from a magazine or cookbook) and have the app automatically extract the recipe details using AI.

## How It Works

1.  **Frontend**:
    *   User clicks "Upload Photo" button on the recipes page
    *   User selects or drags & drops an image file
    *   The image is sent to the backend for processing
    *   Once processed, the user is redirected to the new recipe form with the extracted data pre-filled
    *   User can review and edit the extracted information before saving

2.  **Backend**:
    *   Receives the uploaded image
    *   Sends the image to Google's Gemini API for processing
    *   Extracts recipe name, ingredients, and method from the image
    *   Returns the structured data to the frontend

## Setup Instructions

1.  **Environment Variables**:
    Add the following to your `.env` file in the backend directory:
    ```
    GEMINI_API_KEY=your_gemini_api_key_here
    ```

2.  **Dependencies**:
    The backend requires the `github.com/google/generative-ai-go/genai` package, which is already added to `go.mod`.

## API Endpoint

*   **POST** `/api/v1/recipes/process-photo`
    *   Accepts a multipart form with an image file
    *   Returns a JSON object with the extracted recipe data:
        ```json
        {
          "name": "Recipe Name",
          "ingredients": ["ingredient 1", "ingredient 2", ...],
          "method": "Step-by-step cooking instructions"
        }
        ```

## Frontend Components

1.  **RecipePhotoUploader.svelte**:
    *   Handles file upload UI
    *   Shows preview of the selected image
    *   Manages the upload and processing state
    *   Handles errors and user feedback

2.  **Integration with New Recipe Form**:
    *   The new recipe page (`/recipes/new`) accepts URL parameters to pre-fill the form
    *   After processing a photo, the user is redirected with the extracted data
    *   The form shows a helpful message when pre-filled from a photo

## Error Handling

*   File size limit: 10MB
*   Supported formats: Common image formats (JPEG, PNG, etc.)
*   Clear error messages are shown for invalid files or processing failures

## Security Considerations

*   The image is only used for processing and is not stored permanently
*   The Gemini API key must be kept secure and not exposed to the client
*   File uploads are validated on both client and server side

## Future Improvements

1.  Support for multiple recipe cards in a single photo
2.  Better handling of complex recipe formats
3.  Allow users to correct OCR errors directly in the upload preview
4.  Add support for more image formats and sources (e.g., URL, camera)
