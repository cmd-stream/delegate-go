package client

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/core-go"
	cclnmock "github.com/cmd-stream/core-go/test/mock/client"
	"github.com/cmd-stream/delegate-go"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestKeepaliveDelegate(t *testing.T) {
	delta := 100 * time.Millisecond

	t.Run("Receive should not return PongResult", func(t *testing.T) {
		var (
			wantSeq      core.Seq    = 0
			wantResult   core.Result = nil
			wantN                    = 1
			wantErr                  = errors.New("receive error")
			wantCloseErr error       = nil
			d                        = cclnmock.NewDelegate().RegisterReceive(
				func() (seq core.Seq, result core.Result, n int, err error) {
					return 0, delegate.PongResult{}, 2, nil
				},
			).RegisterReceive(
				func() (seq core.Seq, result core.Result, n int, err error) {
					n = wantN
					err = wantErr
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			dlgt = NewKeepalive(d,
				WithKeepaliveTime(5*time.Second),
				WithKeepaliveIntvl(time.Second),
			)
			mocks               = []*mok.Mock{d.Mock}
			seq, result, n, err = dlgt.Receive()
		)
		asserterror.Equal(t, err, wantErr)
		asserterror.Equal(t, seq, wantSeq)
		asserterror.Equal(t, result, wantResult)
		asserterror.Equal(t, n, wantN)

		err = dlgt.Close()
		asserterror.Equal(t, err, wantCloseErr)

		asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
	})

	t.Run("KeepaliveDelegate should send Ping Commands if no Commands was send",
		func(t *testing.T) {
			var (
				done               = make(chan struct{})
				wantCmd            = delegate.PingCmd[any]{}
				start              = time.Now()
				wantKeepaliveTime  = 2 * 200 * time.Millisecond
				wantKeepaliveIntvl = 200 * time.Millisecond
				d                  = cclnmock.NewDelegate().RegisterNSetSendDeadline(2,
					func(deadline time.Time) (err error) {
						wantDeadline := time.Time{}
						asserterror.SameTime(t, deadline, wantDeadline, delta)
						return
					},
				).RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						var (
							wantTime   = start.Add(wantKeepaliveTime)
							actualTime = time.Now()
						)
						asserterror.SameTime(t, actualTime, wantTime, delta)
						asserterror.Equal(t, seq, 0)
						asserterror.EqualDeep(t, cmd, wantCmd)
						return 1, nil
					},
				).RegisterFlush(
					func() (err error) { return nil },
				).RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						var (
							wantTime   = start.Add(wantKeepaliveTime).Add(wantKeepaliveIntvl)
							actualTime = time.Now()
						)
						asserterror.SameTime(t, actualTime, wantTime, delta)
						asserterror.Equal(t, seq, 0)
						asserterror.EqualDeep(t, cmd, wantCmd)
						return 1, nil
					},
				).RegisterFlush(
					func() (err error) { defer close(done); return nil },
				).RegisterClose(
					func() (err error) { return nil },
				)

				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive(d,
					WithKeepaliveTime(wantKeepaliveTime),
					WithKeepaliveIntvl(wantKeepaliveIntvl))
			)
			dlgt.Keepalive(&sync.Mutex{})

			select {
			case <-done:
			case <-time.NewTimer(wantKeepaliveTime + wantKeepaliveIntvl + time.Second).C:
				t.Fatal("test lasts too long")
			}

			err := dlgt.Close()
			asserterror.EqualError(t, err, nil)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("Command flushing should delay a ping", func(t *testing.T) {
		var (
			done               = make(chan struct{})
			start              = time.Now()
			flushDelay         = 200 * time.Millisecond
			wantKeepaliveTime  = 2 * 200 * time.Millisecond
			wantKeepaliveIntvl = 200 * time.Millisecond
			d                  = cclnmock.NewDelegate().RegisterFlush(
				func() (err error) { return nil },
			).RegisterSetSendDeadline(
				func(deadline time.Time) (err error) {
					wantDeadline := time.Time{}
					asserterror.SameTime(t, deadline, wantDeadline, delta)
					return
				},
			).RegisterSend(
				func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
					var (
						wantTime   = start.Add(flushDelay).Add(wantKeepaliveTime)
						actualTime = time.Now()
					)
					asserterror.SameTime(t, actualTime, wantTime, delta)
					return
				},
			).RegisterFlush(
				func() (err error) { defer close(done); return nil },
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks = []*mok.Mock{d.Mock}
			dlgt  = NewKeepalive(d,
				WithKeepaliveTime(wantKeepaliveTime),
				WithKeepaliveIntvl(wantKeepaliveIntvl))
		)
		dlgt.Keepalive(&sync.Mutex{})
		time.Sleep(flushDelay)

		err := dlgt.Flush()
		asserterror.EqualError(t, err, nil)

		select {
		case <-done:
		case <-time.NewTimer(wantKeepaliveTime + flushDelay + time.Second).C:
			t.Fatal("test lasts too long")
		}

		err = dlgt.Close()
		asserterror.EqualError(t, err, nil)
		asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
	})

	t.Run("Close should cancel ping sending", func(t *testing.T) {
		var (
			d = cclnmock.NewDelegate().RegisterClose(
				func() (err error) { return nil },
			)
			mocks = []*mok.Mock{d.Mock}
			dlgt  = NewKeepalive(d, WithKeepaliveTime(200*time.Millisecond),
				WithKeepaliveIntvl(200*time.Millisecond),
			)
		)
		err := dlgt.Close()
		asserterror.EqualError(t, err, nil)

		time.Sleep(400 * time.Millisecond)
		asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
	})

	t.Run("If ClientDelegate.Close fails with an error, Close should return it and ping shold not be canceled",
		func(t *testing.T) {
			var (
				done    = make(chan struct{})
				wantErr = errors.New("close error")
				d       = cclnmock.NewDelegate().RegisterClose(
					func() (err error) { return wantErr },
				).RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) { return 1, nil },
				).RegisterFlush(
					func() (err error) { defer close(done); return nil },
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive(d,
					WithKeepaliveTime(200*time.Millisecond),
					WithKeepaliveIntvl(200*time.Millisecond),
				)
			)
			dlgt.Keepalive(&sync.Mutex{})

			err := dlgt.Close()
			asserterror.EqualError(t, err, wantErr)

			select {
			case <-done:
			case <-time.NewTimer(time.Second).C:
				t.Fatal("test lsasts too long")
			}

			err = dlgt.Close()
			asserterror.EqualError(t, err, nil)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("If ping sending fails with an error, nothing should happen",
		func(t *testing.T) {
			var (
				done = make(chan struct{})
				d    = cclnmock.NewDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						defer close(done)
						return 1, errors.New("send error")
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive(d,
					WithKeepaliveTime(200*time.Millisecond),
					WithKeepaliveIntvl(200*time.Millisecond),
				)
			)
			dlgt.Keepalive(&sync.Mutex{})

			select {
			case <-done:
			case <-time.NewTimer(time.Second).C:
				t.Fatal("test lsasts too long")
			}

			err := dlgt.Close()
			asserterror.EqualError(t, err, nil)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("If ClientDelegate.Flush fails with an error, Flush should return it and ping sending should not be delayed",
		func(t *testing.T) {
			var (
				done    = make(chan struct{})
				wantErr = errors.New("flush error")
				d       = cclnmock.NewDelegate().RegisterFlush(
					func() (err error) { return wantErr },
				).RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						defer close(done)
						return
					},
				).RegisterFlush(
					func() (err error) { return nil },
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive(d,
					WithKeepaliveTime(200*time.Millisecond),
					WithKeepaliveIntvl(200*time.Millisecond),
				)
			)
			dlgt.Keepalive(&sync.Mutex{})

			err := dlgt.Flush()
			asserterror.EqualError(t, err, wantErr)

			select {
			case <-done:
			case <-time.NewTimer(time.Second).C:
				t.Fatal("test lsasts too long")
			}

			err = dlgt.Close()
			asserterror.EqualError(t, err, nil)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})
}
