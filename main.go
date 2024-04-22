package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	social "avtxtr/social"
	"avtxtr/utils"

	"github.com/lmittmann/tint"
)

var (
	CLEAR_LIST_EVERY_TIME_UNIT = "CLEAR_LIST_EVERY_TIME_UNIT"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC1123Z,
		}),
	))

	// rate limiter
	maxRequestPerTimeUnit := 50
	if envValue := os.Getenv("MAX_REQUEST_PER_TIME_UNIT"); envValue != "" {
		if parsedValue, err := strconv.Atoi(envValue); err != nil {
			slog.Warn("MAX_REQUEST_PER_TIME_UNIT is not a valid number, using default value", "value", maxRequestPerTimeUnit)
		} else {
			maxRequestPerTimeUnit = parsedValue
			slog.Info("MAX_REQUEST_PER_TIME_UNIT is set", "value", maxRequestPerTimeUnit)
		}
	} else {
		slog.Warn("MAX_REQUEST_PER_TIME_UNIT is not set, using default value", "value", maxRequestPerTimeUnit)
	}
	ipList := make(map[string]int)
	go utils.RateLimitClearer(&ipList)

	requestTimeout := 12 * time.Second
	if envValue := os.Getenv("REQUEST_TIMEOUT"); envValue != "" {
		if parsedValue, err := time.ParseDuration(envValue); err != nil {
			slog.Warn("REQUEST_TIMEOUT is not a valid duration, using default value", "value", requestTimeout)
		} else {
			requestTimeout = parsedValue
			slog.Info("REQUEST_TIMEOUT is set", "value", requestTimeout)
		}
	} else {
		slog.Warn("REQUEST_TIMEOUT is not set, using default value", "value", requestTimeout)
	}

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/Delnegend/avtxtr", http.StatusSeeOther)
	})

	http.HandleFunc("GET /{social}/{username}", func(w http.ResponseWriter, r *http.Request) {
		// rate limit
		ip := r.RemoteAddr
		if ipList[ip] > maxRequestPerTimeUnit {
			slog.Warn("rate limited", "ip", ip)
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()

		// extract values
		socialcode := r.PathValue("social")
		username := r.PathValue("username")
		var fallbackAvatar string

		// parse fallback url from localhost:8080/{social}/{username}?fallback=url
		if fallback := r.URL.Query().Get("fallback"); fallback != "" {
			if _, err := url.ParseRequestURI(fallback); err != nil {
				http.Error(w, "invalid fallback URL", http.StatusBadRequest)
				return
			}
			fallbackAvatar = fallback
		}

		slog.Debug("new request", "social", socialcode, "username", username)

		// match the social, get the avatarUrl
		var avatarUrl string
		var err error
		switch socialcode {
		case "deviantart":
			avatarUrl, err = social.DeviantArt(ctx, username)
		case "fa":
			avatarUrl, err = social.FA(ctx, username)
		case "gravatar":
			avatarUrl, err = social.Gravatar(username)
		case "ig", "instagram":
			avatarUrl, err = social.Instagram(ctx, username)
		case "threads":
			avatarUrl, err = social.Threads(ctx, username)
		case "reddit":
			avatarUrl, err = social.Reddit(ctx, username)
		case "telegram":
			avatarUrl, err = social.Telegram(ctx, username)
		case "twitter", "x":
			avatarUrl, err = social.Twitter(ctx, username)
		case "youtube":
			avatarUrl, err = social.YouTube(ctx, username)
		default:
			http.Error(w, "invalid social", http.StatusBadRequest)
			return
		}

		// if can't parse
		if err != nil {
			switch fallbackAvatar != "" {
			case false:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			case true:
				if err := utils.WriteResponseImage(ctx, fallbackAvatar, &w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// validate the url
		if _, err = url.ParseRequestURI(avatarUrl); err != nil {
			switch fallbackAvatar != "" {
			case false:
				http.Error(w, "invalid avatar URL from 3rd party server", http.StatusInternalServerError)
				return
			case true:
				if err := utils.WriteResponseImage(ctx, fallbackAvatar, &w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// write the image
		if err := utils.WriteResponseImage(ctx, avatarUrl, &w); err != nil {
			switch fallbackAvatar != "" {
			case false:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			case true:
				if err := utils.WriteResponseImage(ctx, fallbackAvatar, &w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		http.Error(w, "I'm a teapot", http.StatusTeapot)
	})

	slog.Info("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
