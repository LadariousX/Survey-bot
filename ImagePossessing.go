package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
)

//func decodeImage(imageData []byte) (gocv.Mat, error) {
//	// Try OpenCV first (works for JPEG, PNG, etc.)
//	img, err := gocv.IMDecode(imageData, gocv.IMReadColor)
//	if err == nil && !img.Empty() {
//		return img, nil
//	}
//
//	// If that fails, try HEIC
//	heifImg, err := goheif.Decode(bytes.NewReader(imageData))
//	if err != nil {
//		return gocv.NewMat(), fmt.Errorf("failed to decode image as HEIC: %w", err)
//	}
//
//	// Convert to JPEG and decode with OpenCV
//	var buf bytes.Buffer
//	if err := jpeg.Encode(&buf, heifImg, &jpeg.Options{Quality: 95}); err != nil {
//		return gocv.NewMat(), fmt.Errorf("failed to encode to JPEG: %w", err)
//	}
//
//	img, err = gocv.IMDecode(buf.Bytes(), gocv.IMReadColor)
//	if err != nil {
//		return gocv.NewMat(), fmt.Errorf("failed to decode converted image: %w", err)
//	}
//
//	return img, nil
//}

func fetchImage(attachmentURL string) (string, error) {

	// Parse URL to extract filename
	parsedURL, err := url.Parse(attachmentURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract filename from URL path
	filename := path.Base(parsedURL.Path)

	// Remove query parameters if they ended up in the filename
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	// Trim any leading/trailing whitespace
	filename = strings.TrimSpace(filename)

	// Fallback if filename extraction fails
	if filename == "" || filename == "/" || filename == "." {
		filename = "downloaded_image.jpg"
	}
	resp, err := http.Get(attachmentURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image from discord: %w", err)
	}
	defer func(Body io.ReadCloser) {
		HTTPBodyCloseErr := Body.Close()
		if HTTPBodyCloseErr != nil {
			log.Printf("Error closing body: %v", HTTPBodyCloseErr)
		}
	}(resp.Body)

	// Read the entire response body into memory
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read full image: %w", err)
	}
	if len(imageData) == 0 {
		return "", fmt.Errorf("error: image data is empty")
	}

	// Save to filesystem
	err = os.WriteFile(filename, imageData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return filename, nil
}

//func saveImage(img gocv.Mat) error { // debug-only remove at deployment
//	filename := "postProcess.jpg"
//	ok := gocv.IMWrite(filename, img)
//	if !ok {
//		return fmt.Errorf("failed to save image")
//	}
//	return nil
//}

func ExtractQR(path string) (error, string) {
	// Use venv python locally, fallback to system python in container
	pythonPath := ".venv/bin/python3"
	if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
		pythonPath = "python3"
	}
	cmd := exec.Command(pythonPath, "QRReader.py", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("QR extraction error: %v\nOutput: %s", err, string(output))
		return fmt.Errorf("failed to detect QR: %s", err), ""
	}

	// Split output into lines and find the URL
	lines := strings.Split(string(output), "\n")
	var decoded string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines that start with http:// or https://
		if strings.HasPrefix(strings.ToLower(line), "http://") || strings.HasPrefix(strings.ToLower(line), "https://") {
			decoded = line
			break
		}
	}

	if decoded == "" {
		return fmt.Errorf("no URL found in QR code output"), ""
	}

	// Unescape HTML entities (e.g., &amp; -> &)
	decoded = html.UnescapeString(decoded)

	fmt.Println("found code: ", decoded)
	return nil, decoded
}

//func readText(imageURL string) (string, error) {
//	// request body
//	requestBody := map[string]interface{}{
//		"model": "gpt-4o-mini",
//		"messages": []map[string]interface{}{
//			{
//				"role": "user",
//				"content": []map[string]interface{}{
//					{"type": "text", "text": "This a recept, extract and return only the code exactly as printed in format."},
//					{"type": "image_url", "image_url": map[string]string{"url": imageURL}}, // Corrected format
//				},
//			},
//		},
//	}
//	// Convert request body to JSON
//	jsonData, err := json.Marshal(requestBody)
//	if err != nil {
//		return "", fmt.Errorf("failed to marshal request: %w", err)
//	}
//
//	// Create HTTP request
//	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
//	if err != nil {
//		return "", fmt.Errorf("failed to create request: %w", err)
//	}
//	req.Header.Set("Authorization", "Bearer "+OpenAIKey)
//	req.Header.Set("Content-Type", "application/json")
//
//	// Send request
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", fmt.Errorf("failed to send request: %w", err)
//	}
//	defer resp.Body.Close()
//
//	// Read response
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", fmt.Errorf("failed to read response: %v", err)
//	}
//
//	// Parse response
//	var response map[string]interface{}
//	if err := json.Unmarshal(body, &response); err != nil {
//		return "", fmt.Errorf("failed to parse response: %v", err)
//	}
//
//	// Extract text from response
//	choices, ok := response["choices"].([]interface{})
//	if !ok || len(choices) == 0 {
//		return "", fmt.Errorf("no response choices found")
//	}
//
//	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
//	if !ok {
//		return "", fmt.Errorf("invalid response format")
//	}
//
//	content, ok := message["content"].(string)
//	if !ok {
//		return "", fmt.Errorf("no content found: %w", err)
//	}
//
//	// Print extracted text
//	return content, nil
//}
