package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
)

type ScanResult struct {
	URL          string
	StatusCode   int
	Body         string
	Header       string
	RenderTime   int64
	ScreenShot   []byte
	ErrorMessage string
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum32())
}

func Scan(url string, screenShot bool, browserlessURL string) (ScanResult, error) {
	allocatorCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), browserlessURL)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	var result ScanResult
	var body string
	var screenshot []byte
	start := time.Now()

	var statusCode int
	var header string

	resp, err := http.Get(url)
	if err == nil {
		statusCode = resp.StatusCode
		headerBytes, _ := json.Marshal(resp.Header)
		header = string(headerBytes)
		resp.Body.Close()
	} else {
		result.ErrorMessage = err.Error()
		return result, err
	}

	viewportWidth := 1920
	viewportHeight := 1080

	tasks := []chromedp.Action{
		chromedp.EmulateViewport(int64(viewportWidth), int64(viewportHeight)),
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &body),
	}
	if screenShot {
		tasks = append(tasks, chromedp.FullScreenshot(&screenshot, 90))
	}
	err = chromedp.Run(ctx, tasks...)

	result = ScanResult{
		URL:          url,
		StatusCode:   statusCode,
		Body:         body,
		Header:       header,
		RenderTime:   time.Since(start).Milliseconds(),
		ScreenShot:   screenshot,
		ErrorMessage: "",
	}
	if err != nil {
		result.ErrorMessage = err.Error()
	}

	return result, err
}

func SaveResult(url string, result ScanResult, screenShot bool) error {
	jsonFileName := fmt.Sprintf("output/%s.json", hash(url))
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Error marshaling result to JSON: %v", err)
	}

	err = ioutil.WriteFile(jsonFileName, resultJSON, 0644)
	if err != nil {
		return fmt.Errorf("Error writing result to file: %v", err)
	}

	if screenShot {
		pngFileName := fmt.Sprintf("output/%s.png", hash(url))
		err = ioutil.WriteFile(pngFileName, result.ScreenShot, 0644)
		if err != nil {
			return fmt.Errorf("Error writing screenshot to file: %v", err)
		}
	}

	return nil
}
