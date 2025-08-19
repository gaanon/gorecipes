# Docker Setup for Recipe Photo Upload

This document provides instructions for setting up and testing the recipe photo upload feature in a Docker environment.

## Prerequisites

1. Docker and Docker Compose installed
2. OpenAI API key for the photo processing feature
3. (Optional) Test recipe image file

## Environment Setup

1. Copy the example environment file and update it with your API keys:
   ```bash
   cp .env.example .env
   ```
   Edit the `.env` file and set your `OPENAI_API_KEY` and other required environment variables.

2. Build and start the Docker containers:
   ```bash
   docker-compose up --build -d
   ```

3. Verify the containers are running:
   ```bash
   docker-compose ps
   ```

## Testing the Photo Upload Feature

### Using the Web Interface

1. Open your browser and navigate to: http://localhost:5173/recipes
2. Click the "Upload Photo" button
3. Select or drag-and-drop a recipe photo
4. The app will process the photo and pre-fill a new recipe form with the extracted data
5. Review and edit the information, then save the recipe

### Using the Test Script

A test script is provided to test the photo upload API directly:

1. Make the script executable:
   ```bash
   chmod +x scripts/test_photo_upload_docker.sh
   ```

2. Run the test script with a recipe image:
   ```bash
   ./scripts/test_photo_upload_docker.sh path/to/your/recipe.jpg
   ```

   If no image path is provided, it will look for a file named `test_recipe.jpg` in the current directory.

## Troubleshooting

### Common Issues

1. **File Upload Size Limit**:
   - If you encounter file size limit errors, ensure the following settings are properly configured:
     - Nginx: `client_max_body_size` in `frontend/nginx.conf`
     - Backend: Check for any file size limits in the Go server configuration

2. **API Connection Issues**:
   - Verify the backend service is running: `docker-compose logs backend`
   - Check the frontend logs: `docker-compose logs frontend`

3. **OpenAI API Errors**:
   - Ensure your OpenAI API key is valid and has sufficient credits
   - Check the backend logs for detailed error messages

### Viewing Logs

- View all container logs:
  ```bash
  docker-compose logs -f
  ```

- View backend logs:
  ```bash
  docker-compose logs -f backend
  ```

- View frontend logs:
  ```bash
  docker-compose logs -f frontend
  ```

## Performance Considerations

1. **File Upload Size**:
   - The default maximum file size is set to 10MB. Adjust this in the following locations if needed:
     - `frontend/nginx.conf`: `client_max_body_size`
     - `frontend/Dockerfile`: `client_max_body_size` in the Nginx configuration

2. **Processing Time**:
   - Photo processing involves calling the OpenAI API, which may take several seconds
   - Timeout settings are configured in both Nginx and the backend service

## Security Considerations

1. **API Keys**:
   - Never commit your `.env` file to version control
   - The `.env` file is included in `.gitignore` by default

2. **File Uploads**:
   - All uploaded files are validated before processing
   - Only image files are accepted
   - Files are processed in memory and not stored permanently

## Cleaning Up

To stop and remove all containers and volumes:

```bash
docker-compose down -v
```

To completely remove all Docker resources (containers, networks, volumes, and images):

```bash
docker-compose down --rmi all -v
```
