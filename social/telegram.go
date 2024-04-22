package social

import (
	"avtxtr/utils"
	"context"
	"fmt"
	"net/http"
	"regexp"
)

func Telegram(ctx context.Context, username string) (string, error) {
	// validate username
	if len(username) < 5 || len(username) > 32 {
		return "", fmt.Errorf("username must be at least 5 characters long")
	}
	if match, _ := regexp.MatchString(`[^a-z0-9_]`, username); match {
		return "", fmt.Errorf("invalid characters in username")
	}

	// create req
	req, err := http.NewRequestWithContext(ctx, "GET", "https://t.me/"+username, nil)
	if err != nil {
		return "", err
	}

	// do req
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
	content := string(contentByte)

	pattern, err := regexp.Compile(`<img class="tgme_page_photo_image" src="(.*?)">`)
	if err != nil {
		return "", err
	}
	match := pattern.FindStringSubmatch(content)
	if len(match) < 2 {
		return "", fmt.Errorf("the user may not have a profile picture, or it's private")
	}
	result := match[1]

	if result == "" {
		return "", fmt.Errorf("no avatar found for %s", username)
	}

	return result, nil
}
