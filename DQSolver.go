package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
)

func processDQ(url string) (string, error) {
	cleanScreenshotErr := os.RemoveAll("screenshots")
	if cleanScreenshotErr != nil && !os.IsNotExist(cleanScreenshotErr) {
		fmt.Println(cleanScreenshotErr)
	}

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

	fmt.Println("def chrome")

	run1Err := chromedp.Run(ctx, chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second), takeScreenshot("0 init"))
	if run1Err != nil {
		return "", fmt.Errorf("chromedp run 1 err: %w", run1Err)
	}
	jsResult := false

	for i := 0; i < 40; i++ {
		label := strconv.Itoa(i + 1)
		run2Err := chromedp.Run(ctx,
			chromedp.WaitReady("#NextButton"),
			takeScreenshot(label),
			chromedp.Click("#NextButton"),
			chromedp.WaitReady("#footerframe"),
			chromedp.Sleep(time.Second),

			chromedp.Evaluate(`
   				(function() {
				   var elem = document.evaluate(
     				    "//p[contains(text(), 'Please write the following validation code')]",
    				     document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null
   				   ).singleNodeValue;
					return elem !== null;
   				})()`, &jsResult,
				chromedp.EvalAsValue,
			),
			takeScreenshot(label),
		)

		if run2Err != nil {
			fmt.Println(run2Err)
			return "", fmt.Errorf("chromedp run 2 err, the survey may have expired: %w", run2Err)
		}

		if jsResult {
			screenshotRunErr := chromedp.Run(ctx, takeScreenshot("99 confirmation"))
			if screenshotRunErr != nil {
				return "", fmt.Errorf("chromedp screenshot run err: %w", screenshotRunErr)
			}
			break
		}

		if i == 39 {
			return "", fmt.Errorf("DQ bruteforce timed out")
		}
	}
	var res string
	run3Err := chromedp.Run(ctx, chromedp.Text(`.ValCode`, &res, chromedp.ByQuery))
	if run3Err != nil {
		return "", fmt.Errorf("chromedp run 3 err: %w", run3Err)
	}
	if len(res) < 17 {
		return "", fmt.Errorf("verification code string unexpected length: %s", res)
	}
	return res[17:], nil
}
