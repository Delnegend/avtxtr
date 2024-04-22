package social

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"avtxtr/utils"

	"github.com/google/uuid"
)

func scrapeWithCookie(ctx context.Context, username string) (string, error) {
	// get the cookies from env
	cookie_a, cookie_b := os.Getenv("FA_COOKIE_A"), os.Getenv("FA_COOKIE_B")

	// validate cookies
	if _, err := uuid.Parse(cookie_a); err != nil {
		return "", fmt.Errorf("FA_COOKIE_A is not a valid UUID string")
	}
	if _, err := uuid.Parse(cookie_b); err != nil {
		return "", fmt.Errorf("FA_COOKIE_B is not a valid UUID string")
	}

	// new req
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.furaffinity.net/user/"+username, nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(&http.Cookie{Name: "a", Value: cookie_a, Path: "/"})
	req.AddCookie(&http.Cookie{Name: "b", Value: cookie_b, Path: "/"})

	// do the req
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read the response
	contentByte, err := utils.ReadAllWithContext(ctx, resp.Body)
	if err != nil {
		return "", err
	}

	return string(contentByte), nil
}

func scrapeNoCookie(ctx context.Context, username string) (string, error) {
	// init new req
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.furaffinity.net/user/"+username, nil)
	if err != nil {
		return "", err
	}

	// do the req w/ default client & read the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	contentByte, err := utils.ReadAllWithContext(ctx, resp.Body)
	if err != nil {
		return "", err
	}
	content := string(contentByte)

	if strings.Contains(content, "user cannot be found") {
		return "", fmt.Errorf("user cannot be found")
	}

	if strings.Contains(content, "registered users only") {
		return "", fmt.Errorf("FA cookies required")
	}

	return content, nil
}

func FA(ctx context.Context, username string) (string, error) {
	// validate username
	username = strings.ToLower(strings.TrimSpace(username))
	if len(username) > 50 {
		return "", fmt.Errorf("username is too long")
	}
	if match, _ := regexp.MatchString(`[^a-z0-9_\-~.]`, username); match {
		return "", fmt.Errorf("invalid characters in username")
	}

	// do the scraping
	var content string
	var err error
	if os.Getenv("FA_COOKIE_A") != "" && os.Getenv("FA_COOKIE_B") != "" {
		content, err = scrapeWithCookie(ctx, username)
	} else {
		content, err = scrapeNoCookie(ctx, username)
	}
	if err != nil {
		return "", err
	}

	// strip unnecessary data
	start := "<userpage-nav-avatar>"
	startIndex := strings.Index(content, start)
	if startIndex == -1 {
		return "", fmt.Errorf("userpage-nav-avatar element not found: %s", content)
	}
	endIndex := strings.Index(content[startIndex:], "</userpage-nav-avatar>")
	if endIndex == -1 {
		return "", fmt.Errorf("userpage-nav-avatar end not found")
	}
	content = content[startIndex : startIndex+endIndex]

	// use regex to find match
	pattern, _ := regexp.Compile(`<img.*?src="(.*?)"\/>`)
	match := pattern.FindStringSubmatch(content)
	if len(match) < 2 {
		return "", fmt.Errorf("no avatar found for %s", username)
	}
	result := match[1]

	// more post-processing
	if result == "" {
		return "", fmt.Errorf("no avatar found for %s", username)
	}
	if strings.HasPrefix(result, "//") {
		result = "https:" + result
	}
	return result, nil
}
