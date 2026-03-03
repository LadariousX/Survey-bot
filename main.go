/*
IDEAS:
  - use chromium or go container (lighten)
*/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
)

var modes = [4]string{"Help", "Cane", "What", "DQ"} // code options for initial mode filter
var (
	DefaultEmail string
	BotToken     string
	CapSolverKey string
	OpenAIKey    string
)
var reconnectFailures = 0

// findChromePath detects Chrome/Chromium path for both local and container environments
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

// runSolverSafe runs a solver function in a goroutine with panic recovery
func runSolverSafe(solverName string, solverFunc func() error) error {
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in %s solver: %v", solverName, r)
				errChan <- fmt.Errorf("%s solver panicked: %v", solverName, r)
			}
		}()
		err := solverFunc()
		errChan <- err
	}()
	return <-errChan
}

func main() {
	err := godotenv.Load() // get constants and defaults
	if err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}
	BotToken = os.Getenv("BotToken")
	DefaultEmail = os.Getenv("DefaultEmail")

	bot, newSessionErr := discordgo.New("Bot " + BotToken)
	if newSessionErr != nil {
		reconnectFailures++
		if reconnectFailures > 2 {
			log.Fatalf(" reconnection failure, attempt surpassed, exiting")
		}
		log.Printf("Error creating Discord session: %v", newSessionErr)
	}

	bot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		var processErr []error
		// Ignore messages from the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}
		var sendConfImg = true

		//=======================
		// qr reader and local ocr
		//=======================
		if len(m.Attachments) > 0 {
			fmt.Println("image received: ", m.Attachments[0].URL)
			_, msgErr5 := s.ChannelMessageSend(m.ChannelID, "Received image...")
			if msgErr5 != nil {
				panic(fmt.Sprintf("Error (msgErr5) sending Discord message: %v", msgErr5))
			}

			imgPath, imageErr := fetchImage(m.Attachments[0].URL)
			if imageErr != nil { // break if the image doesn't load
				processErr = append(processErr, imageErr)
			}

			QRErr, url := ExtractQR(imgPath) // will not support guest
			if QRErr != nil {
				processErr = append(processErr, fmt.Errorf("%w", QRErr))
			}

			if len(url) > 0 { // perhaps unnecessary check
				if strings.Contains(url, "mydqexperience") {
					DQErr := runSolverSafe("DQ", func() error {
						return processDQ(url)
					})
					if DQErr != nil {
						processErr = append(processErr, DQErr)
					}

				} else if strings.Contains(url, "whataburger") {
					whataburgerErr := runSolverSafe("Whataburger", func() error {
						return processWhataburger(url, DefaultEmail)
					})
					if whataburgerErr != nil {
						processErr = append(processErr, whataburgerErr)
					}

				} else {
					processErr = append(processErr, fmt.Errorf("error sorting the QR URL to solvers"))
				}
			}

			//=======================
			// if regular text message
			//=======================
		} else {

			codeValidityErr, code := isValidCode(m.Content)
			if codeValidityErr != nil { // invalid code
				sendConfImg = false
				_, msgErr0 := s.ChannelMessageSend(m.ChannelID, "Code format error, use the \"Help\" command for further instruction.")
				if msgErr0 != nil {
					panic(fmt.Sprintf("Error (msgErr0) sending Discord message: %v", msgErr0))
				}

			} else { // case of valid code
				if code[0] != "Help" { // if help cmd then skip RX code message
					_, msgErr1 := s.ChannelMessageSend(m.ChannelID, "Received code, processing...")
					if msgErr1 != nil {
						panic(fmt.Sprintf("Error (msgErr1) sending Discord message: %v", msgErr1))
					}
				}

				// choose and handoff to survey solver
				switch code[0] {
				case "Help":
					sendConfImg = false
					processErr = nil
					out := ""
					for i := range modes {
						out = out + modes[i] + ", "
					}
					_, msgErr2 := s.ChannelMessageSend(m.ChannelID, "send images of QR codes or use the CLI messenger. "+
						"Command format:\n\"<Mode> <Code/URL> <Email>\""+
						"\nAvailable modes: "+out+"\nGuest Email function is not supported by all modes or by QR.")
					if msgErr2 != nil {
						panic(fmt.Sprintf("Error (msgErr2) sending Discord message: %v", msgErr2))
					}

				case "Cane":
					CaneErr := runSolverSafe("Canes", func() error {
						return processCanes(code[1], code[2]) // code: [(mode) code email]
					})
					if CaneErr != nil {
						processErr = append(processErr, CaneErr)
					}
				case "What":
					WhatErr := runSolverSafe("Whataburger", func() error {
						return processWhataburger(code[1], code[2])
					})
					if WhatErr != nil {
						processErr = append(processErr, WhatErr)
					} else {

					}

				case "DQ":
					DQErr := runSolverSafe("DQ", func() error {
						return processDQ(code[1])
					})
					if DQErr != nil {
						processErr = append(processErr, DQErr)
					}
				}
			}
		}

		// handel errors then clear

		if len(processErr) > 0 {
			log.Printf("Error processing survey: %v", processErr)
			_, msgErr3 := s.ChannelMessageSend(m.ChannelID, "There was an error completing the survey:\n--> "+processErr[len(processErr)-1].Error())
			if msgErr3 != nil {
				panic(fmt.Sprintf("Error (msgErr3) sending Discord message: %v", msgErr3))
			}
			sendConfImg = false

		} else {
			if sendConfImg { //respond with result, picture if no conf img send failure msg
				confImgErr := confImg(s, m.ChannelID)
				if confImgErr != nil {
					_, msgErr4 := s.ChannelMessageSend(m.ChannelID, "No image available, there was likely an undetected automation error")
					if msgErr4 != nil {
						panic(fmt.Sprintf("Error (msgErr4) sending Discord message: %v", msgErr4))
					}
				}
			}
		}
	})

	// Open a websocket to Discord
	newSessionErr = bot.Open()
	if newSessionErr != nil {
		log.Fatalf("Error opening connection to Discord: %v", newSessionErr)
	}

	fmt.Println("Bot has started")
	// Wait for a termination signal to exit
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	botCloseErr := bot.Close()
	if botCloseErr != nil {
		log.Fatalf("Error closing Discord session, terminating program: %v", botCloseErr)
	}
}

func isValidCode(message string) (error, []string) {
	out := make([]string, 3)
	cutMsg := strings.Fields(message) // [mode code email]
	fmt.Println(cutMsg)

	// Handle raw links first
	if len(cutMsg) > 0 {
		if strings.Contains(cutMsg[0], "mydqexperience") {
			out = []string{"DQ", cutMsg[0], ""}
			return nil, out
		}
		if strings.Contains(cutMsg[0], "whataburger") {
			email := DefaultEmail
			if len(cutMsg) >= 2 {
				email = cutMsg[1]
			}
			out = []string{"What", cutMsg[0], email}
			return nil, out
		}
	}

	// Continue with mode formatting
	var modeValid = false

	for i := range modes { // check that mode exist
		if cutMsg[0] == modes[i] {
			modeValid = true
		}
	}

	if !modeValid { // Error for invalid modes
		out = []string{"", "", ""}
		return fmt.Errorf("mode does not exist"), out
	}

	if cutMsg[0] == "Help" {
		out = []string{"Help", "", ""} // Help w/o code
		return nil, out
	}

	if len(message) > 6 { // EX <Cane 123> 4char+ space+ code
		if len(cutMsg) == 2 {
			out = []string{cutMsg[0], cutMsg[1], DefaultEmail}
			return nil, out
		}
		if len(cutMsg) == 3 {
			out = []string{cutMsg[0], cutMsg[1], cutMsg[2]}
			return nil, out
		}
	}
	out = []string{"", "", ""}
	return fmt.Errorf("unknown code / validity error"), out

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
		filename := fmt.Sprintf("screenshots/%s.png", label)
		if err := os.WriteFile(filename, screenshot, 0644); err != nil {
			return fmt.Errorf("failed to save screenshot to screenshots/%s: %w", filename, err)
		}

		fmt.Print(label + ", ")
		return nil
	}
}
