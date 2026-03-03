package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	capsolver "github.com/capsolver/capsolver-go"
	"github.com/chromedp/chromedp"
)

func processWhataburger(url, email string) error {
	capSolver := capsolver.CapSolver{ApiKey: os.Getenv("CapSolverKey")}

	cleanScreenshotErr := os.RemoveAll("screenshots")
	if cleanScreenshotErr != nil && !os.IsNotExist(cleanScreenshotErr) {
		fmt.Println(cleanScreenshotErr)
	}
	task := map[string]any{
		"type":        "ReCaptchaV2EnterpriseTaskProxyLess",
		"websiteURL":  url,
		"websiteKey":  "6Ldxd94ZAAAAANgjv1UpUZ1nAj-P35y3etQOwBrC",
		"isInvisible": true,
	}

	fmt.Print("Solving CAPTCHA... ")
	solution, err := capSolver.Solve(task)
	if err != nil {
		return fmt.Errorf("CAPTCHA solving failed: %w", err)
	}
	captchaToken := solution.Solution.GRecaptchaResponse

	// Start Chromedp with browser path detection
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(findChromePath()),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
	defer func() {
		takeScreenshot("98 failure")
		cancel()
	}()

	var captchaResult string
	var prefacePage bool

	err = chromedp.Run(ctx, // run 1
		chromedp.Navigate(url),
		// inject CAPTCHA token
		chromedp.WaitReady(`textarea[name="g-recaptcha-response"]`),
		chromedp.EvaluateAsDevTools(
			`document.querySelectorAll('textarea[name="g-recaptcha-response"]').forEach(el => {
        	el.value = "`+captchaToken+`";
        	el.dispatchEvent(new Event('change', { bubbles: true }));});`, nil),
		chromedp.EvaluateAsDevTools( // Ensure site detects new CAPTCHA response
			`document.querySelector('textarea#g-recaptcha-response-100000').value = "`+captchaToken+`";
    		document.querySelector('textarea#g-recaptcha-response-100000').dispatchEvent(new Event('change', { bubbles: true }));`, nil),
		chromedp.EvaluateAsDevTools(`document.body.innerText.includes("incorrect CAPTCHA") ? "FAILED" : "PASSED";`, &captchaResult),
		chromedp.Evaluate(`
        	Array.from(document.querySelectorAll("button, input[type='button']")).some(el => el.id === "NextButton");`, &prefacePage),
	)
	if err != nil {
		return fmt.Errorf("chromedp run (1) err: %w", err)
	}

	if captchaResult == "FAILED" {
		return fmt.Errorf("CAPTCHA token was not accepted by website")
	}
	fmt.Println("CAPTCHA token accepted")

	if !prefacePage { // skip the preface page with hidden buttons if applicable
		fmt.Println("running preface")
		err = chromedp.Run(ctx,
			chromedp.WaitReady("body"),
			chromedp.Sleep(2*time.Second),
			takeScreenshot("0"),
			chromedp.Click("#NextButton", chromedp.ByID),
			//chromedp.WaitReady("#QID1319640458-6-label"),
			chromedp.Sleep(2*time.Second),
			takeScreenshot("0.5"),
			chromedp.Click("#QID1319640458-6-label", chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
			takeScreenshot("1"),
		)
		if err != nil {
			return fmt.Errorf("chromedp run (2) err: %w", err)
		}
	}
	jsResult := false
	i := 0
	for jsResult == false {
		err = chromedp.Run(ctx,
			chromedp.WaitReady("#NextButton"),
			chromedp.Click("#NextButton", chromedp.ByID),
			chromedp.Sleep(2*time.Second),

			chromedp.Evaluate(`
   				(function () {
					var elem = document.evaluate(
					"//*[@id='QR~QID1319640437~1']",
					document,
					null,
					XPathResult.FIRST_ORDERED_NODE_TYPE,
					null
					).singleNodeValue;
					
					  return elem !== null;
					})()`, &jsResult,
				chromedp.EvalAsValue,
			),
			takeScreenshot(strconv.Itoa(i+2)),
		)

		if err != nil {
			return fmt.Errorf("chromedp run (3) err: %w", err)
		}
		i++
	}
	fmt.Println("JS result true")

	err = chromedp.Run(ctx,
		// enter email & send
		chromedp.SendKeys(`#QR\~QID1319640437\~1`, email, chromedp.ByQuery),
		chromedp.Click("#NextButton", chromedp.ByID),
		chromedp.Sleep(time.Second),
		takeScreenshot("99 done"),
	)
	if err != nil {
		return fmt.Errorf("chromedp run (4) err: %w", err)
	}

	return nil
}
