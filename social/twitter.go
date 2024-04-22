package social

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func scrapeWithBrowser(ctx context.Context, username string) (string, error) {
	chromedpAddr := os.Getenv("CHROMEDP_ADDR")
	if chromedpAddr == "" {
		return "", fmt.Errorf("using basic GET request failed, requires CHROMEDP_ADDR to be set")
	}

	// create contexts
	ctx, cancel := chromedp.NewRemoteAllocator(ctx, chromedpAddr)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// enable network
	err := chromedp.Run(ctx, network.Enable())
	if err != nil {
		return "", err
	}

	resultCh := make(chan string, 1)

	// listen for all network events
	chromedp.ListenTarget(ctx, func(event interface{}) {
		switch ev := event.(type) {
		case *network.EventRequestWillBeSent:
			if ev.Type == network.ResourceTypeImage && strings.HasSuffix(ev.Request.URL, "_200x200.jpg") {
				resultCh <- ev.Request.URL
			}
		}
	})

	// navigate, wait for profile image to be visible
	if err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("https://twitter.com/" + username),
		chromedp.WaitVisible(`img[src$="_200x200.jpg"]`),
	}); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case url := <-resultCh:
		if url == "" {
			return "", fmt.Errorf("no avatar found for %s", username)
		}
		return url, nil
	}
}

func scrapeEmbededVersion(ctx context.Context, username string) (string, error) {
	url := "https://twitter.com/" + username

	// create a get request with custom User-Agent
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Bot")

	// send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read the response body
	type readResult struct {
		body []byte
		err  error
	}
	resultCh := make(chan readResult, 1)
	go func() {
		body, err := io.ReadAll(resp.Body)
		resultCh <- readResult{body: body, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-resultCh:
		if result.err != nil {
			return "", result.err
		}

		// find the image url
		rgx, err := regexp.Compile(
			`<meta content="(https:\/\/pbs.twimg.com\/profile_images\/.*?)" property="og:image" \/>`)
		if err != nil {
			return "", err
		}
		match := rgx.FindStringSubmatch(string(result.body))
		if len(match) < 2 {
			return "", fmt.Errorf("no avatar found for %s", username)
		}

		return match[1], nil
	}
}

func Twitter(ctx context.Context, username string) (string, error) {
	// validate username
	if len(username) < 4 || len(username) > 15 {
		return "", fmt.Errorf("username length must be between 4 and 15")
	}
	if match, _ := regexp.MatchString(`[^a-zA-Z0-9_]`, username); match {
		return "", fmt.Errorf("invalid characters in username")
	}

	// scrape the url
	result, err := scrapeEmbededVersion(ctx, username)
	if err != nil {
		slog.Warn("failed to scrape embed", "username", username, "error", err)
		result, err = scrapeWithBrowser(ctx, username)
	}
	if err != nil {
		return "", err
	}
	return strings.Replace(result, "_200x200", "", 1), nil
}
