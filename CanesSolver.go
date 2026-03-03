package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"math/rand"
	"os"
	"strings"
	"time"
)

func processCanes(rawSurveyCode, email string) error {
	SurveyCode := strings.Split(rawSurveyCode, "-")

	var OpenResponse = [7]string{
		"Canes has the best batter i think for all the restaurants. its really light and not super dry and crunchy. I like to get it buttered on both sides toast and it tastes much better!",
		"Canes is nice especially if i ever get the chance to go inside because everybody always super friendly and i live the place. Always think abt getting merch but never do, I should though.",
		"Its always packed so that's bound to mean something.",
		"At the drive through they were playing music I like, which is normal really I need to find the playlist.",
		"Not certin but I ve only like canes for a few years. I was introduced to it by this girl i liked, I may avoid her at all cost now but my love for canes remains",
		"Back in highschool I would keep canes napkins on the dash of my truck. kinda like some people do with wataburger signs.",
		"I love the lemonaid but somtimes idk what to get lemonaid or coke bc it can be kinda tart somtimes but it makes me feel healthy. idk if thats the case tho."}
	randResponse := OpenResponse[rand.Intn(len(OpenResponse))]

	randNum := make([]string, 10)
	for i := 0; i < len(randNum); i++ {
		randNum[i] = fmt.Sprintf("%d", rand.Intn(2)+4) // Convert randNum int to string
	}

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

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://raisingcane.survey.marketforce.com/?languageId=1"),
		chromedp.WaitVisible("/html/body/div[1]/div/div/div[2]/form/div/input[15]"),

		chromedp.SendKeys("//*[@id=\"EntryCode1\"]", SurveyCode[0]), // enter code in 4 parts
		chromedp.SendKeys("//*[@id=\"EntryCode2\"]", SurveyCode[1]),
		chromedp.SendKeys("//*[@id=\"EntryCode3\"]", SurveyCode[2]),
		chromedp.SendKeys("//*[@id=\"EntryCode4\"]", SurveyCode[3]),
		takeScreenshot("1 survey code"),
		chromedp.Click("/html/body/div[1]/div/div/div[2]/form/div/input[15]"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("//*[@id=\"Question_034235147237135146234008059057115075206176156241_1\"]"), //select food&drink
		takeScreenshot("2 food and drink"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("//*[@id=\"Question_172123088239018139033186249169214168132113219077_2\"]"), //select dive through
		takeScreenshot("3 drive through"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("//*[@id=\"Question_060089005111074016140194198086015224174167035074_1\"]"), //check fingers
		chromedp.Click("//*[@id=\"Question_060089005111074016140194198086015224174167035074_3\"]"), //check fries
		chromedp.Click("//*[@id=\"Question_060089005111074016140194198086015224174167035074_4\"]"), //check toast
		chromedp.Click("//*[@id=\"Question_060089005111074016140194198086015224174167035074_6\"]"), //check sauce
		chromedp.Click("//*[@id=\"Question_060089005111074016140194198086015224174167035074_7\"]"), //check beverage
		takeScreenshot("4 select products"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("//*[@id=\"Question_002217024192199059192156167054135233137111129251_5\"]"), //check 5 service
		takeScreenshot("5 service"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_190110018235012241181070084023253011159027082036_"+randNum[0]), //check rand food
		takeScreenshot("6 rand food"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_082222013163035246181028169002071235047104241224_"+randNum[1]), //rate chicken
		chromedp.Click("#Question_185007230251238102196142156242082209093175084086_2"),           //rate fries
		chromedp.Click("#Question_177128234184229204201147103064153140164156022210_"+randNum[2]), //rate toast
		chromedp.Click("#Question_187210049013078003052233053230046143189221184222_"+randNum[3]), //rate sauce
		chromedp.Click("#Question_229102146110152151116011243123039013154240066036_"+randNum[4]), //rate drink
		takeScreenshot("7 rate products"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_165191176113180194048053244103030123159061089177_2"), // cold fry
		chromedp.Click("#Question_165191176113180194048053244103030123159061089177_3"), // soggy fry
		takeScreenshot("8 soggy fries"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_223073255134019224218141232077055145147092193011_1"), // correct order
		takeScreenshot("9 correct order"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_100086062153039021039089086164000077075014253152_"+randNum[5]), //rate atmosphere
		chromedp.Click("#Question_046135018179179234083162134107198001065147124115_"+randNum[6]), //rate service
		chromedp.Click("#Question_159087132092018244219249063090070068157074241184_"+randNum[7]), //rate time
		chromedp.Click("#Question_164077072230226086104076184246139073208221180092_"+randNum[8]), //rate time2
		chromedp.Click("#Question_051205056145110239018139062045126193110089112156_"+randNum[9]), //rate clean
		takeScreenshot("10 rate location"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_069031138043091199136231018154194180153169004066_4"), // rate price
		takeScreenshot("11 rate price"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_025140229039115035077028114014117055002107017108_5"), // likely to return
		chromedp.Click("#Question_167098115105091073124147168242161171080075013066_5"), // recommend
		takeScreenshot("12 tell friends"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_170022163004160052062245139223026106174248084253_5"), // i <3 canes
		takeScreenshot("13 I <3 canes"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_124167117038057047045059058181039040200147214103_2"), // fav part
		takeScreenshot("14 fav part"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.Click("#Question_129200253244023221041177171072200245119151058165_2"), // no prob
		takeScreenshot("15 no problems"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.SendKeys("//*[@id=\"Question_222078254060033117128218030121204242116210144200\"]", randResponse), //open-ended response
		takeScreenshot("16 open ended response"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // next
		chromedp.WaitReady("/html/body/div[1]/section/div/form/nav/div/div[2]/input"),
		chromedp.Sleep(time.Second),
		chromedp.SendKeys("//*[@id=\"Questions[0].CustomFields[0].Response\"]", "Layden Blackwell"),
		chromedp.SendKeys("//*[@id=\"Questions[0].CustomFields[1].Response\"]", "1841 HWY 111 South"),
		chromedp.SendKeys("//*[@id=\"Questions[0].CustomFields[2].Response\"]", "Edna"),
		chromedp.SendKeys("//*[@id=\"Questions[0].CustomFields[3].Response\"]", "TX"),
		chromedp.SendKeys("//*[@id=\"Questions[0].CustomFields[4].Response\"]", "77957"),
		chromedp.SendKeys("//*[@id=\"Questions[0].CustomFields[5].Response\"]", email),
		chromedp.SendKeys("//*[@id=\"Questions[0].CustomFields[6].Response\"]", "3612357978"),
		takeScreenshot("17 contact info"),
		chromedp.Click("/html/body/div[1]/section/div/form/nav/div/div[2]/input"), // submit
		chromedp.Sleep(10*time.Second),
		takeScreenshot("99 confirmation"),
	)

	if err != nil {
		return fmt.Errorf("chromedp err: %w", err)
	}

	return nil
}
