package utils

import (
	"context"
	"fmt"
	"io"
)

type readResult struct {
	data []byte
	err  error
}

func ReadAllWithContext(ctx context.Context, r io.Reader) ([]byte, error) {
	resultCh := make(chan readResult, 1)
	go func() {
		if r == nil {
			resultCh <- readResult{data: nil, err: fmt.Errorf("reader is nil")}
			return
		}
		data, err := io.ReadAll(r)
		resultCh <- readResult{data: data, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		return result.data, result.err
	}
}
