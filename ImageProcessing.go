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
