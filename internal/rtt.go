package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Dreamacro/clash/constant"
)

const (
	GOOGLE_RTT_PROBE_URL = "http://www.gstatic.com/generate_204"
)

func probeGoogle(p constant.Proxy) (int, error) {
	targetUrlString := GOOGLE_RTT_PROBE_URL
	conn, err := dialProxyConn(context.Background(), p, targetUrlString)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodHead, targetUrlString, nil)
	if err != nil {
		return 0, err
	}
	client := createHTTPClient(conn)
	defer client.CloseIdleConnections()

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
	// fmt.Println(time.Since(start).Milliseconds())
	return int(time.Since(start).Milliseconds()), nil
}

func ProbeLinkMinRtt(N int, p constant.Proxy) (int, error) {
	minRtt := 0
	for i := 0; i < N; i++ {
		if rtt, err := probeGoogle(p); err != nil {
			return 0, err
		} else {
			if minRtt == 0 || rtt < minRtt {
				minRtt = rtt
			}
		}
	}
	return minRtt, nil
}
