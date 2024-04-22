package utils

import (
	"context"
	"net/http"
)

func WriteResponseImage(ctx context.Context, avatarUrl string, w *http.ResponseWriter) error {
	// get the image data
	req, err := http.NewRequestWithContext(ctx, "GET", avatarUrl, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// write to response writer
scoped:
	for {
		select {
		case <-ctx.Done():
			return err
		default:
			data := make([]byte, 1024)
			if n, err := resp.Body.Read(data); err != nil {
				break scoped
			} else {
				(*w).Write(data[:n])
			}
		}
	}

	return nil
}
