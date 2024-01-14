package exporters

import (
	"context"
	"net/http"
	"sync"
)

func partitionSlice[T any](partitionSize int, slice []T) [][]T {
	var resSlice [][]T
	for i := 0; i*partitionSize < len(slice); i++ {
		begIndex := i * partitionSize
		endIndex := begIndex + partitionSize
		if endIndex >= len(slice) {
			endIndex = len(slice)
		}
		resSlice = append(resSlice, slice[begIndex:endIndex:endIndex])
	}
	return resSlice
}

// MultiHttpClient is a multi-threaded http client
type MultiHttpClient struct {
	Parallelism  int
	Client       http.Client
	IgnoreErrors bool
	requests     []*http.Request
	wg           sync.WaitGroup
	callback     func(res *http.Response)
}

func (m *MultiHttpClient) QueueRequest(req *http.Request) {
	m.requests = append(m.requests, req)
}

func (m *MultiHttpClient) OnResponse(cb func(res *http.Response)) {
	m.callback = cb
}

func (m *MultiHttpClient) Execute(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	ptRequests := partitionSlice(m.Parallelism, m.requests)
outerLoop:
	for _, v := range ptRequests {
		select {
		case <-ctx.Done():
			break outerLoop
		default:
			for _, request := range v {
				m.wg.Add(1)
				go func(req *http.Request) {
					res, err := m.Client.Do(req)
					if err != nil && !m.IgnoreErrors {
						cancel(err)
						m.wg.Done()
						return
					}
					m.callback(res)
					m.wg.Done()
				}(request)
			}
			m.wg.Wait()
		}
	}
	err := context.Cause(ctx)
	if err != nil {
		return err
	}
	return nil
}
