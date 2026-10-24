package server

import (
	"errors"
	"fmt"
	"io"
	"net/url"
)

func readMonitorResponse(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, (4<<20)+1))
	if err == nil && len(data) > 4<<20 {
		return nil, errors.New("monitor response too large")
	}
	return data, err
}

func transportError(err error) error {
	var u *url.Error
	if errors.As(err, &u) {
		return fmt.Errorf("monitor transport %s: %w", u.Op, u.Err)
	}
	return err
}
