package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Dreamacro/clash/constant"
)

const (
	// uses HTTPS incase some services cheat with HTTP packets inspection and early return without proxying
	GOOGLE_RTT_PROBE_URL = "https://www.gstatic.com/generate_204"
)

func ProbeLinkMinRtt(N int, p constant.Proxy) (int, error) {
	targetUrlString := GOOGLE_RTT_PROBE_URL
	Logger.Debug().
		Str("proxy", p.Addr()).
		Str("desc", p.Name()).
		Str("url", targetUrlString).
		Msg("probe_rtt")

	conn, err := DialProxyConn(context.Background(), p, targetUrlString)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := createHTTPClient(conn)
	defer client.CloseIdleConnections()

	fn := func() (int, error) {
		req, err := http.NewRequest(http.MethodHead, targetUrlString, nil)
		if err != nil {
			return 0, err
		}
		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			err = fmt.Errorf("http status code: %d", resp.StatusCode)
			return 0, err
		}
		return int(time.Since(start).Milliseconds()), nil
	}

	minRtt := 0
	for i := 0; i < N; i++ {
		if rtt, err := fn(); err != nil {
			return 0, err
		} else {
			if minRtt == 0 || rtt < minRtt {
				minRtt = rtt
			}
		}
	}
	return minRtt, nil
}
