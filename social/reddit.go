package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type userInfo struct {
	Data struct {
		Subreddit struct {
			IconImg string `json:"icon_img"`
		} `json:"subreddit"`
	} `json:"data"`
}

func Reddit(ctx context.Context, username string) (string, error) {
	// validate username
	if len(username) < 3 || len(username) > 20 {
		return "", fmt.Errorf("username length must be between 3 and 20 characters")
	}
	if match, _ := regexp.MatchString(`[^a-zA-Z0-9_-]`, username); match {
		return "", fmt.Errorf("invalid characters in username")
	}

	url := fmt.Sprintf("https://www.reddit.com/user/%s/about.json", username)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var user userInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	result := user.Data.Subreddit.IconImg
	if result == "" {
		return "", fmt.Errorf("no avatar found for %s", username)
	}

	slice := strings.Split(result, "?")
	if len(slice) > 1 {
		return slice[0], nil
	}
	return result, nil
}
