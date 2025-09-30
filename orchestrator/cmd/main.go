package main

import (
	"context"
	_ "encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable not set")
	}

	// 1. --- DEFINE THE TOOL FOR THE LLM ---
	// Define the function declaration for the model
	getWeatherFunc := &genai.FunctionDeclaration{
		Name:        "getWeather",
		Description: "Get the current weather for a specific location.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"location": {
					Type:        genai.TypeString,
					Description: "Any city in the United States, i.e. New_York_City",
				},
			},
			Required: []string{"location"},
		},
	}

	// Configure the client and tools
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}
	tools := &genai.Tool{FunctionDeclarations: []*genai.FunctionDeclaration{getWeatherFunc}}
	config := &genai.GenerateContentConfig{Tools: []*genai.Tool{tools}}

	// Define user prompt
	contents := []*genai.Content{{Role: "user", Parts: []*genai.Part{{Text: "Tell me weather in New York City."}}}}

	// Send request with function declarations
	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-pro",
		contents,
		config,
	)
	if err != nil {
		log.Fatal(err)
	}

	//  Check for a function call
	if resp.Candidates[0].Content.Parts[0].FunctionCall != nil {
		fCall := resp.Candidates[0].Content.Parts[0].FunctionCall
		fmt.Printf("Function call: %s\n", fCall.Name)
		fmt.Printf("Arguments: %v\n", fCall.Args)
		result, err := callGetWeatherAPI(fCall.Args["location"].(string))
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("Result: %s\n", string(result))

		// Create a function response part
		part := genai.NewPartFromFunctionResponse(fCall.Name, map[string]any{"result": result})

		// Append function call and result of the function execution to contents
		contents = append(contents, resp.Candidates[0].Content)
		contents = append(contents, &genai.Content{Role: "user", Parts: []*genai.Part{part}})

		finalResp, err := client.Models.GenerateContent(
			ctx,
			"gemini-2.5-flash",
			contents,
			config,
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(finalResp.Text())
	} else {
		fmt.Println("No function call found in the response.")
		fmt.Println(resp.Text())
	}
}

// callGetWeatherAPI is our function to call the actual tool server
func callGetWeatherAPI(location string) ([]byte, error) {
	log.Println("calling weather API")
	apiURL := fmt.Sprintf("http://localhost:8080/getWeather?location=%s", location)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to tool server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tool server returned non-200 status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
