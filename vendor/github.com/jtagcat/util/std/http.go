package std

import (
	"context"
	"io"
	"net/http"
)

// http.Post() + ctx
func PostWithContext(ctx context.Context, c *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// http.Get()+ ctx
func GetWithContext(ctx context.Context, c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func CookieByKey(cookies []*http.Cookie, key string) string {
	for _, cookie := range cookies {
		if cookie.Name == key {
			return cookie.Value
		}
	}

	return ""
}
