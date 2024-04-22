package social

import (
	"avtxtr/utils"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func Threads(ctx context.Context, username string) (string, error) {
	return meta(ctx, username, func(username string) string {
		return fmt.Sprintf("https://www.threads.net/@%s", username)
	})
}
func Instagram(ctx context.Context, username string) (string, error) {
	return meta(ctx, username, func(username string) string {
		return fmt.Sprintf("https://www.instagram.com/%s", username)
	})
}

func meta(ctx context.Context, username string, template func(string) string) (string, error) {
	username = strings.TrimPrefix(username, "@")

	// validate username
	if len(username) < 1 || len(username) > 30 {
		return "", fmt.Errorf("username length must be between 1 and 30")
	}
	if match, _ := regexp.MatchString(`[a-zA-Z0-9_\.]`, username); !match {
		return "", fmt.Errorf("username can only contain letters, numbers, underscore and period")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", template(username), nil)
	if err != nil {
		return "", err
	}

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

	re := regexp.MustCompile(`<meta property="og:image" content="(.*?)" />`)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return "", fmt.Errorf("no avatar found for %s", username)
	}
	result := match[1]
	result = strings.ReplaceAll(result, "&amp;", "&")

	if result == "" {
		return "", fmt.Errorf("no avatar found for %s", username)
	}

	return result, nil
}
