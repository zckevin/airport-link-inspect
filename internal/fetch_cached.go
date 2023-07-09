package internal

import (
	"context"
	"sync"

	"github.com/Dreamacro/clash/constant"
	"github.com/zckevin/airport-link-inspect/internal/types"
)

type bytesResult = types.Result[[]byte]

type fetchResult struct {
	body bytesResult
	done chan struct{}
}

func newFetchResult() *fetchResult {
	return &fetchResult{
		done: make(chan struct{}),
	}
}

var (
	fetchResultsHistoryMu sync.Mutex
	fetchResultsHistory   map[string]*fetchResult = make(map[string]*fetchResult)
)

func upsertCachedFetchResult(targetUrlString string) (*fetchResult, bool) {
	fetchResultsHistoryMu.Lock()
	defer fetchResultsHistoryMu.Unlock()

	result, ok := fetchResultsHistory[targetUrlString]
	if !ok {
		result = newFetchResult()
		fetchResultsHistory[targetUrlString] = result
	}
	return result, ok
}

func FetchWithProxyWithHistory(ctx context.Context, p constant.Proxy, targetUrlString string) ([]byte, error) {
	result, ok := upsertCachedFetchResult(targetUrlString)
	if ok {
		<-result.done
	} else {
		result.body = types.AsResult(FetchWithProxy(ctx, p, targetUrlString))
		close(result.done)
	}
	return result.body.Expand()
}
