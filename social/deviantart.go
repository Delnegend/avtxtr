package social

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

func DeviantArt(ctx context.Context, username string) (string, error) {
	// validate username
	if len(username) < 3 || len(username) > 20 {
		return "", fmt.Errorf("username length must be between 3 and 20")
	}
	if match, _ := regexp.MatchString(`[^a-zA-Z0-9]`, username); match {
		return "", fmt.Errorf("invalid characters in username")
	}

	url := fmt.Sprintf("https://a.deviantart.net/avatars-big/%s/%s/%s", username[0:1], username[1:2], username)

	// try .jpg first
	res, err := http.Head(url + ".jpg")
	if err != nil {
		return "", err
	}
	if res.StatusCode == http.StatusOK {
		return url + ".jpg", nil
	}

	// try .png
	res, err = http.Head(url + ".png")
	if err != nil {
		return "", err
	}
	if res.StatusCode == http.StatusOK {
		return url + ".png", nil
	}

	return "", fmt.Errorf("no avatar found for %s", username)
}
