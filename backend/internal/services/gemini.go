package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiService provides a client for interacting with the Google Gemini API.
type GeminiService struct {
	client *genai.GenerativeModel
}

// NewGeminiService creates a new GeminiService.
func NewGeminiService(ctx context.Context) (*GeminiService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash-latest")

	return &GeminiService{
		client: model,
	}, nil
}

// ProcessRecipeImage sends an image to the Gemini API and returns the extracted recipe information.
func (s *GeminiService) ProcessRecipeImage(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	prompt := genai.Text(
		"Extract the recipe details from the provided image. Return a JSON object with the following structure:\n" +
			"{\n" +
			"  \"name\": \"Recipe name\",\n" +
			"  \"ingredients\": [\"ingredient 1\", \"ingredient 2\", ...],\n" +
			"  \"method\": \"Step-by-step cooking instructions\"\n" +
			"}\n" +
			"\n" +
			"Only include the JSON object in your response, nothing else.",
	)

	resp, err := s.client.GenerateContent(ctx, genai.ImageData("jpeg", fileBytes), prompt)
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if txt, ok := part.(genai.Text); ok {
		return string(txt), nil
	}

	return "", fmt.Errorf("unexpected response format")
}
