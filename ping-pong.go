package delegate

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
)

type PingCmd[T any] struct{}

func (c PingCmd[T]) Exec(ctx context.Context, seq base.Seq, at time.Time,
	receiver T, proxy base.Proxy) (err error) {
	_, err = proxy.SendWithDeadline(seq, PongResult{}, time.Time{})
	return
}

type PongResult struct{}

func (r PongResult) LastOne() bool {
	return true
}
