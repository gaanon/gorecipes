# Recipe Photo Upload - Test Plan

## Test Environment Setup
1. **Backend Server**
   - Ensure the backend server is running on `http://localhost:8080`
   - Set the `OPENAI_API_KEY` environment variable
   - Verify database connection is working

2. **Frontend**
   - Ensure frontend dev server is running (typically `http://localhost:5173`)
   - Clear browser cache to ensure latest frontend code is loaded

## Test Cases

### 1. Photo Upload UI
- [ ] Verify "Upload Photo" button is present on the recipes page
- [ ] Clicking the button opens the upload modal
- [ ] Modal allows file selection via click or drag-and-drop
- [ ] Only image files are accepted (JPG, PNG, etc.)
- [ ] File size limit is enforced (5MB)
- [ ] Preview of selected image is shown
- [ ] Loading state is shown during processing
- [ ] Error messages are displayed for invalid files

### 2. Backend Processing
- [ ] Upload a clear photo of a recipe
  - Verify successful response (200 OK)
  - Check that the response includes:
    - Recipe name
    - List of ingredients
    - Cooking method
- [ ] Upload a non-image file
  - Verify error response (400 Bad Request)
- [ ] Upload an image larger than 5MB
  - Verify file size limit error
- [ ] Upload an image with no text
  - Verify appropriate error handling
- [ ] Test with poor quality/blurry images
  - Verify graceful degradation

### 3. Integration with Recipe Creation
- [ ] After successful processing:
  - Verify redirection to new recipe form
  - Check that form is pre-filled with extracted data
  - Verify that all fields are editable
  - Test form submission with the pre-filled data

### 4. Error Handling
- [ ] Test with invalid API key
  - Verify appropriate error message is shown to the user
- [ ] Test with no internet connection
  - Verify offline handling
- [ ] Test with rate limiting
  - Verify proper rate limit handling

## Performance Testing
- [ ] Measure processing time for different image sizes
- [ ] Test with multiple concurrent uploads

## Security Testing
- [ ] Verify that the OpenAI API key is not exposed to the client
- [ ] Test with malicious file uploads
- [ ] Verify proper CORS configuration

## Browser Compatibility
- [ ] Test in Chrome
- [ ] Test in Firefox
- [ ] Test in Safari
- [ ] Test on mobile devices

## Test Data
- Sample recipe images are available in `test_data/recipes/`
  - `simple_recipe.jpg` - Clear photo of a simple recipe
  - `complex_recipe.jpg` - Photo with complex formatting
  - `poor_quality.jpg` - Blurry or poorly lit photo
  - `no_text.jpg` - Image with no recipe text

## Running Tests

### Backend Tests
```bash
cd backend
go test -v ./...
```

### Frontend Tests
```bash
cd frontend
npm test
```

### Manual Testing
1. Start the backend server:
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. Start the frontend dev server:
   ```bash
   cd frontend
   npm run dev
   ```

3. Open `http://localhost:5173/recipes` in your browser
4. Follow the test cases above

## Expected Results
- All test cases should pass
- The user experience should be smooth and intuitive
- Error messages should be clear and helpful
- The feature should work consistently across different browsers and devices
