package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
)

// TODO: add see more button after interal server errror in html

func main() {
	godotenv.Load()
	addr := "0.0.0.0:3000"

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/survey-bot", handleSurvey)

	log.Printf("Survey-bot listening on %q", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// findChromePath detects Chrome / Chromium path for both local and container environments
func findChromePath() string {
	// Container path (Chromium)
	if _, err := os.Stat("/usr/bin/chromium"); err == nil {
		return "/usr/bin/chromium"
	}
	// macOS paths (Chrome)
	if _, err := os.Stat("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"); err == nil {
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}
	// Linux Chrome
	if _, err := os.Stat("/usr/bin/google-chrome"); err == nil {
		return "/usr/bin/google-chrome"
	}
	// Fallback to default (chromedp will search standard locations)
	return ""
}

func runSolver(solverName string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in %s solver: %v", solverName, r)
			err = fmt.Errorf("%s solver panicked: %v", solverName, r)
		}
	}()
	return fn()
}

type surveyRequest struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}

type surveyResponse struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func handleSurvey(w http.ResponseWriter, r *http.Request) {
	var req surveyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	println(req.URL, req.Email)
	if req.URL == "" || req.Email == "" {
		http.Error(w, "url and email are required", http.StatusBadRequest)
		return
	}

	var solveErr []error
	resp := surveyResponse{Message: "Survey completed successfully."}

	switch {
	case strings.Contains(req.URL, "feedback.whataburger.com"):
		WhatErr := runSolver("Whataburger", func() error {
			return processWhataburger(req.URL, req.Email)
		})
		if WhatErr != nil {
			solveErr = append(solveErr, WhatErr)
		}
		resp.Message = "Survey completed successfully. Check promotions in your email for coupon."

	case strings.Contains(req.URL, "mydqexperience"):
		DQErr := runSolver("DQ", func() error {
			var err error
			resp.Code, err = processDQ(req.URL)
			return err
		})
		if DQErr != nil {
			solveErr = append(solveErr, DQErr)
		}

	default:
		http.Error(w, "unrecognized survey URL", http.StatusBadRequest)
		return
	}

	if solveErr != nil {
		var msgs []string
		for _, e := range solveErr {
			msgs = append(msgs, e.Error())
		}
		http.Error(w, strings.Join(msgs, "\n"), http.StatusInternalServerError)
		return
	}

	println(resp.Message)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	println("response sent")
}

func confImg(session *discordgo.Session, channelID string) error {

	files, findFileErr := os.ReadDir("screenshots")
	if findFileErr != nil {
		return findFileErr
	}
	sort.Slice(files, func(i, j int) bool { // find file by the highest number
		return files[i].Name() < files[j].Name()
	})
	latestImgPath := files[len(files)-1]
	file, imgOpenErr := os.Open("screenshots/" + latestImgPath.Name())
	if imgOpenErr != nil {
		return fmt.Errorf("failed to open file: %w", imgOpenErr)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}(file)

	_, sendImgErr := session.ChannelFileSend(channelID, "confirmation.png", file)
	if sendImgErr != nil {
		return fmt.Errorf("failed to send file: %w", sendImgErr)
	}
	return nil
}

func takeScreenshot(label string) chromedp.ActionFunc {
	time.Sleep(20 * time.Millisecond)
	return func(ctx context.Context) error {
		var screenshot []byte
		// Capture the screenshot
		if err := chromedp.FullScreenshot(&screenshot, 100).Do(ctx); err != nil {
			return fmt.Errorf("failed to capture screenshot: %w", err)
		}

		// Save the screenshot to a file
		if err := os.MkdirAll("screenshots", 0755); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", "folder", err)
		}
		filename := fmt.Sprintf("screenshots/%d_%s.png", os.Getpid(), label)
		if err := os.WriteFile(filename, screenshot, 0644); err != nil {
			return fmt.Errorf("failed to save screenshot to screenshots/%s: %w", filename, err)
		}

		fmt.Print(label + ", ")
		return nil
	}
}
