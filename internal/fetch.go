package internal

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Dreamacro/clash/constant"
)

type JsonBody map[string]any

func dialProxyConn(ctx context.Context, p constant.Proxy, targetUrlString string) (net.Conn, error) {
	addr, err := urlToMetadata(targetUrlString)
	if err != nil {
		return nil, err
	}
	return p.DialContext(ctx, &addr)
}

func createHTTPClient(conn net.Conn) *http.Client {
	transport := &http.Transport{
		Dial: func(string, string) (net.Conn, error) {
			return conn, nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       20 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &client
}

func FetchWithProxy(ctx context.Context, p constant.Proxy, targetUrlString string) ([]byte, error) {
	Logger.Debug().
		Str("proxy", p.Addr()).
		Str("desc", p.Name()).
		Str("url", targetUrlString).
		Msg("fetch")
	conn, err := dialProxyConn(ctx, p, targetUrlString)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, targetUrlString, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip")
	// req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")
	req = req.WithContext(ctx)

	client := createHTTPClient(conn)
	defer client.CloseIdleConnections()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http status code: %d", resp.StatusCode)
		return nil, err
	}

	var body io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		body, _ = gzip.NewReader(resp.Body)
		defer body.Close()
	default:
		body = resp.Body
	}
	buf, err := io.ReadAll(body)
	Logger.Debug().Str("resp_body", string(buf)).Msg(targetUrlString)
	return buf, err
}

func urlToMetadata(rawURL string) (addr constant.Metadata, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			err = fmt.Errorf("%s scheme not Support", rawURL)
			return
		}
	}

	addr = constant.Metadata{
		Host:    u.Hostname(),
		DstIP:   nil,
		DstPort: port,
	}
	return
}
