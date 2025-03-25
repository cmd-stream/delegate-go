package dcln

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/base-go/testdata/mock"
	"github.com/cmd-stream/delegate-go"
	"github.com/ymz-ncnk/mok"
)

func TestKeepaliveDelegate(t *testing.T) {

	t.Run("Receive should not return PongResult", func(t *testing.T) {
		var (
			wantErr = errors.New("receive error")
			d       = mock.NewClientDelegate().RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
					return 0, delegate.PongResult{}, nil
				},
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
					err = wantErr
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			dlgt = NewKeepalive[any](d,
				WithKeepaliveTime(5*time.Second),
				WithKeepaliveIntvl(time.Second),
			)
			mocks            = []*mok.Mock{d.Mock}
			seq, result, err = dlgt.Receive()
		)
		if seq != 0 {
			t.Errorf("unexpected seq, want '%v' actual '%v'", 0, seq)
		}
		if result != nil {
			t.Errorf("unexpected result, want '%v' actual '%v'", nil, result)
		}
		if err != wantErr {
			t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
		}
		if err := dlgt.Close(); err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("KeepaliveDelegate should send Ping Commands if no Commands was send",
		func(t *testing.T) {
			var (
				done               = make(chan struct{})
				wantSeq            = 0
				wantCmd            = delegate.PingCmd[any]{}
				start              = time.Now()
				wantKeepaliveTime  = 2 * 200 * time.Millisecond
				wantKeepaliveIntvl = 200 * time.Millisecond
				d                  = mock.NewClientDelegate().RegisterNSetSendDeadline(2,
					func(deadline time.Time) (err error) {
						wantTime := time.Time{}
						if !SameTime(wantTime, deadline) {
							t.Errorf("unexpected dealine, want '%v' actual '%v'", wantTime,
								deadline)
						}
						return
					},
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						var (
							wantTime   = start.Add(wantKeepaliveTime)
							actualTime = time.Now()
						)
						if !SameTime(wantTime, actualTime) {
							t.Errorf("unexpected time, want '%v' actual '%v'", wantTime,
								actualTime)
						}
						if seq != 0 {
							t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq, seq)
						}
						if !reflect.DeepEqual(wantCmd, cmd) {
							t.Errorf("unexpected cmd, want '%v' actual '%v'",
								wantCmd,
								cmd)
						}
						return nil
					},
				).RegisterFlush(
					func() (err error) { return nil },
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						var (
							wantTime   = start.Add(wantKeepaliveTime).Add(wantKeepaliveIntvl)
							actualTime = time.Now()
						)
						if !SameTime(wantTime, actualTime) {
							t.Errorf("unexpected time, want '%v' actual '%v'", wantTime,
								actualTime)
						}
						if seq != 0 {
							t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq, seq)
						}
						if !reflect.DeepEqual(wantCmd, cmd) {
							t.Errorf("unexpected cmd, want '%v' actual '%v'",
								wantCmd,
								cmd)
						}
						return nil
					},
				).RegisterFlush(
					func() (err error) { defer close(done); return nil },
				).RegisterClose(
					func() (err error) { return nil },
				)

				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive[any](d,
					WithKeepaliveTime(wantKeepaliveTime),
					WithKeepaliveIntvl(wantKeepaliveIntvl))
			)
			dlgt.Keepalive(&sync.Mutex{})
			select {
			case <-done:
			case <-time.NewTimer(wantKeepaliveTime + wantKeepaliveIntvl + time.Second).C:
				t.Fatal("test lasts too long")
			}

			if err := dlgt.Close(); err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Command flushing should delay a ping", func(t *testing.T) {
		var (
			done               = make(chan struct{})
			start              = time.Now()
			flushDelay         = 200 * time.Millisecond
			wantKeepaliveTime  = 2 * 200 * time.Millisecond
			wantKeepaliveIntvl = 200 * time.Millisecond
			d                  = mock.NewClientDelegate().RegisterFlush(
				func() (err error) { return nil },
			).RegisterSetSendDeadline(
				func(deadline time.Time) (err error) {
					wantTime := time.Time{}
					if !SameTime(wantTime, deadline) {
						t.Errorf("unexpected dealine, want '%v' actual '%v'", wantTime,
							deadline)
					}
					return
				},
			).RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					var (
						wantTime   = start.Add(flushDelay).Add(wantKeepaliveTime)
						actualTime = time.Now()
					)
					if !SameTime(wantTime, actualTime) {
						t.Errorf("unexpected time, want '%v' actual '%v'", wantTime,
							actualTime)
					}
					return
				},
			).RegisterFlush(
				func() (err error) { defer close(done); return nil },
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks = []*mok.Mock{d.Mock}
			dlgt  = NewKeepalive[any](d,
				WithKeepaliveTime(wantKeepaliveTime),
				WithKeepaliveIntvl(wantKeepaliveIntvl))
		)
		dlgt.Keepalive(&sync.Mutex{})
		time.Sleep(flushDelay)
		if err := dlgt.Flush(); err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
		select {
		case <-done:
		case <-time.NewTimer(wantKeepaliveTime + flushDelay + time.Second).C:
			t.Fatal("test lasts too long")
		}
		if err := dlgt.Close(); err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("Close should cancel ping sending", func(t *testing.T) {
		var (
			d = mock.NewClientDelegate().RegisterClose(
				func() (err error) { return nil },
			)
			mocks = []*mok.Mock{d.Mock}
			dlgt  = NewKeepalive[any](d,
				WithKeepaliveTime(200*time.Millisecond),
				WithKeepaliveIntvl(200*time.Millisecond),
			)
		)
		if err := dlgt.Close(); err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
		time.Sleep(400 * time.Millisecond)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("If ClientDelegate.Close fails with an error, Close should return it and ping shold not be canceled",
		func(t *testing.T) {
			var (
				done    = make(chan struct{})
				wantErr = errors.New("close error")
				d       = mock.NewClientDelegate().RegisterClose(
					func() (err error) { return wantErr },
				).RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) { return nil },
				).RegisterFlush(
					func() (err error) { defer close(done); return nil },
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive[any](d,
					WithKeepaliveTime(200*time.Millisecond),
					WithKeepaliveIntvl(200*time.Millisecond),
				)
			)
			dlgt.Keepalive(&sync.Mutex{})
			if err := dlgt.Close(); err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
			select {
			case <-done:
			case <-time.NewTimer(time.Second).C:
				t.Fatal("test lsasts too long")
			}
			if err := dlgt.Close(); err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If ping sending fails with an error, nothing should happen",
		func(t *testing.T) {
			var (
				done = make(chan struct{})
				d    = mock.NewClientDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						defer close(done)
						return errors.New("send error")
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive[any](d,
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
			if err := dlgt.Close(); err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If ClientDelegate.Flush fails with an error, Flush should return it and ping sending should not be delayed",
		func(t *testing.T) {
			var (
				done    = make(chan struct{})
				wantErr = errors.New("flush error")
				d       = mock.NewClientDelegate().RegisterFlush(
					func() (err error) { return wantErr },
				).RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						defer close(done)
						return
					},
				).RegisterFlush(
					func() (err error) { return nil },
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{d.Mock}
				dlgt  = NewKeepalive[any](d,
					WithKeepaliveTime(200*time.Millisecond),
					WithKeepaliveIntvl(200*time.Millisecond),
				)
			)
			dlgt.Keepalive(&sync.Mutex{})
			if err := dlgt.Flush(); err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
			select {
			case <-done:
			case <-time.NewTimer(time.Second).C:
				t.Fatal("test lsasts too long")
			}
			if err := dlgt.Close(); err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

}
