package social

import (
	"avtxtr/utils"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func YouTube(ctx context.Context, username string) (string, error) {
	// validate username
	username = strings.TrimPrefix(username, "@")
	if len(username) < 3 || len(username) > 30 {
		return "", fmt.Errorf("username length must be between 3 and 30")
	}
	if match, _ := regexp.MatchString(`[^a-zA-Z0-9_\-\.]`, username); match {
		return "", fmt.Errorf("invalid characters in username")
	}

	// create req
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.youtube.com/@"+username, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Googlebot-Image")

	// do req
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read response
	contentByte, err := utils.ReadAllWithContext(ctx, resp.Body)
	if err != nil {
		return "", err
	}
	content := string(contentByte)

	// find the image url
	rgx, err := regexp.Compile(`<meta property="og:image" content="(.*?)">`)
	if err != nil {
		return "", err
	}
	match := rgx.FindStringSubmatch(content)
	if len(match) < 2 {
		return "", fmt.Errorf("no avatar found for %s", username)
	}
	result := match[1]

	if result == "" {
		return "", fmt.Errorf("no avatar found for %s", username)
	}

	return result, nil
}
