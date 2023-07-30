package delegate

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
)

type PingCmd[T any] struct{}

func (c PingCmd[T]) Exec(ctx context.Context, at time.Time, seq base.Seq,
	receiver T, proxy base.Proxy) error {
	return proxy.Send(seq, PongResult{})
}

type PongResult struct{}

func (r PongResult) LastOne() bool {
	return true
}
