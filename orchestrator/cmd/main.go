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
					Description: "The city and state, e.g., San Francisco, CA",
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

	// Send request with function declarations
	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-pro",
		genai.Text("Tell me weather in London"),
		&genai.GenerateContentConfig{Tools: []*genai.Tool{{FunctionDeclarations: []*genai.FunctionDeclaration{getWeatherFunc}}}},
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
	} else {
		fmt.Println("No function call found in the response.")
		fmt.Println(resp.Text())
	}

	//// Add the function declaration to the model's toolset
	//client.Tools = []*genai.Tool{
	//	{FunctionDeclarations: []*genai.FunctionDeclaration{getWeatherFunc}},
	//}

	// 2. --- START THE CONVERSATION ---
	//prompt := "What is the current weather like in Tokyo?"
	//log.Printf("User Prompt: %s\n", prompt)
	//
	//// Send the prompt to the model
	//resp, err := client.GenerateContent(ctx, genai.Text(prompt))
	//if err != nil {
	//	log.Fatalf("Error generating content: %v", err)
	//}

	// 3. --- CHECK IF THE MODEL WANTS TO CALL A TOOL ---
	//part := resp.Candidates[0].Content.Parts[0]
	//if fc, ok := part.(genai.FunctionCall); ok {
	//	log.Printf("Model wants to call a function: %s\n", fc.Name)
	//	log.Printf("Arguments: %v\n", fc.Args)
	//
	//	// We only have one tool, but a switch is good for scalability
	//	switch fc.Name {
	//	case "getWeather":
	//		location := fc.Args["location"].(string)
	//
	//		// 4. --- EXECUTE THE TOOL (Call our other server) ---
	//		weatherData, err := callGetWeatherAPI(location)
	//		if err != nil {
	//			log.Fatalf("Error calling weather API: %v", err)
	//		}
	//		log.Printf("Tool Result: %s\n", string(weatherData))
	//
	//		// 5. --- SEND THE TOOL'S RESPONSE BACK TO THE MODEL ---
	//		// Create a FunctionResponse part with the tool's output
	//		fr := &genai.FunctionResponse{
	//			Name:    "getWeather",
	//			Content: string(weatherData),
	//		}
	//
	//		// Send the response back to the model
	//		resp, err = client.GenerateContent(ctx, fr)
	//		if err != nil {
	//			log.Fatalf("Error generating content after function call: %v", err)
	//		}
	//	default:
	//		log.Fatalf("Unknown function call: %s", fc.Name)
	//	}
	//}
	//
	//// 6. --- PRINT THE FINAL, NATURAL-LANGUAGE RESPONSE ---
	//finalResponse := resp.Candidates[0].Content.Parts[0]
	//if text, ok := finalResponse.(genai.Text); ok {
	//	fmt.Println("---")
	//	fmt.Println("Final Answer from Gemini:")
	//	fmt.Println(text)
	//	fmt.Println("---")
	//} else {
	//	log.Printf("Unexpected final response type: %T\n", finalResponse)
	//}
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
