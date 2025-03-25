package dcln

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	bmock "github.com/cmd-stream/base-go/testdata/mock"
	"github.com/cmd-stream/delegate-go"
	dmock "github.com/cmd-stream/delegate-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

const Delta = 100 * time.Millisecond

func TestDelegate(t *testing.T) {

	var (
		ops        = []SetOption{WithServerInfoReceiveDuration(0)}
		serverInfo = delegate.ServerInfo([]byte("server info"))
	)

	t.Run("New should check ServerInfo", func(t *testing.T) {
		var (
			wantErr   error = nil
			conn            = bmock.NewConn()
			transport       = MakeClientTransport(time.Now(), serverInfo,
				200*time.Millisecond,
				300*time.Millisecond,
				time.Second,
				t)
			mocks = []*mok.Mock{conn.Mock, transport.Mock}
		)
		testDelegateCreation(serverInfo, transport, ops, wantErr, mocks, t)
	})

	t.Run("If Transport.SetReceiveDeadline fails with an error before receive ServerInfo, New should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.SetReceiveDeadline")
				transport = dmock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				)
				mocks = []*mok.Mock{transport.Mock}
			)
			testDelegateCreation(serverInfo, transport, ops, wantErr, mocks, t)
		})

	t.Run("If Transport.ReceiveServerInfo fails with an error, New should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.ReceiveServerInfo error")
				transport = dmock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterReceiveServerInfo(
					func() (info delegate.ServerInfo, err error) {
						return nil, wantErr
					},
				)
				mocks = []*mok.Mock{transport.Mock}
			)
			testDelegateCreation(serverInfo, transport, ops, wantErr, mocks, t)
		})

	t.Run("If wrong ServerInfo was received, New should return error",
		func(t *testing.T) {
			var (
				wantErr         = ErrServerInfoMismatch
				wrongServerInfo = []byte{1}
				transport       = dmock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterReceiveServerInfo(
					func() (info delegate.ServerInfo, err error) {
						return wrongServerInfo, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock}
			)
			testDelegateCreation(serverInfo, transport, ops, wantErr, mocks, t)
		})

	t.Run("New should apply Conf.ServerInfoReceiveDuration", func(t *testing.T) {
		var (
			d         = time.Second
			ops       = []SetOption{WithServerInfoReceiveDuration(d)}
			startTime = time.Now()
			transport = dmock.NewClienTransport().RegisterSetReceiveDeadline(
				func(deadline time.Time) (err error) {
					wantDeadline := startTime.Add(d)
					if !SameTime(deadline, wantDeadline) {
						err = fmt.Errorf("unexpected deadline, want '%v' actual '%v'",
							wantDeadline,
							deadline)
					}
					return
				},
			).RegisterReceiveServerInfo(
				func() (info delegate.ServerInfo, err error) {
					return serverInfo, nil
				},
			).RegisterSetReceiveDeadline(
				func(deadline time.Time) (err error) {
					return
				},
			)
			mocks = []*mok.Mock{transport.Mock}
		)
		testDelegateCreation(serverInfo, transport, ops, nil, mocks, t)
	})

	t.Run("If Transport.Send fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Delegate.Send error")
				transport = dmock.NewClienTransport().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						return wantErr
					},
				)
				delegate = Delegate[any]{transport: transport}
			)
			err := delegate.Send(1, bmock.NewCmd())
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
		})

	t.Run("Transport.Send should send same seq and cmd as Send",
		func(t *testing.T) {
			var (
				wantSeq   base.Seq = 1
				wantCmd            = bmock.NewCmd()
				transport          = dmock.NewClienTransport().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						if seq != wantSeq {
							return fmt.Errorf("unexppected seq, want '%v' actual '%v'", wantSeq,
								seq)
						}
						if !reflect.DeepEqual(wantCmd, cmd) {
							t.Errorf("unexpected cmd, want '%v' actual '%v'", wantCmd, cmd)
						}
						return nil
					},
				)
				mocks    = []*mok.Mock{transport.Mock}
				delegate = Delegate[any]{transport: transport}
			)
			err := delegate.Send(wantSeq, wantCmd)
			if err != nil {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Receive should return same seq and cmd as Tranposrt.Receive",
		func(t *testing.T) {
			var (
				wantErr             = errors.New("receive failed")
				wantSeq    base.Seq = 1
				wantResult          = bmock.NewResult()
				transport           = dmock.NewClienTransport().RegisterReceive(
					func() (seq base.Seq, r base.Result, err error) {
						return wantSeq, wantResult, wantErr
					},
				)
				delegate = Delegate[any]{transport: transport}
			)
			seq, result, err := delegate.Receive()
			if seq != wantSeq {
				t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq, seq)
			}
			if result != wantResult {
				t.Errorf("unexpected result, want '%v' actual '%v'", wantResult, result)
			}
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
		})

	t.Run("Conf should return the conf that was obtained during creation",
		func(t *testing.T) {
			var (
				wantO    = Options{}
				delegate = Delegate[any]{}
			)
			o := delegate.Options()
			if o != wantO {
				t.Errorf("unexpected o, want '%v' actual '%v'", wantO, o)
			}
		})

	t.Run("LocalAddr should return Transport.LocalAddr", func(t *testing.T) {
		var (
			wantAddr  = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			transport = dmock.NewClienTransport().RegisterLocalAddr(
				func() (a net.Addr) {
					return wantAddr
				},
			)
			delegate = Delegate[any]{transport: transport}
		)
		addr := delegate.LocalAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
	})

	t.Run("RemoteAddr should return Transport.RemoteAddr", func(t *testing.T) {
		var (
			wantAddr  = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			transport = dmock.NewClienTransport().RegisterRemoteAddr(
				func() (addr net.Addr) {
					return wantAddr
				},
			)
			delegate = Delegate[any]{transport: transport}
		)
		addr := delegate.RemoteAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
	})

	t.Run("If Tranposrt.Close fails with an error, Close should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Close error")
				transport = dmock.NewClienTransport().RegisterClose(
					func() (err error) {
						return wantErr
					},
				)
				delegate = Delegate[any]{transport: transport}
			)
			err := delegate.Close()
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
		})

}

func SameTime(t1, t2 time.Time) bool {
	return !(t1.Before(t2.Truncate(Delta)) || t1.After(t2.Add(Delta)))
}

func MakeClientTransport(startTime time.Time, serverInfo delegate.ServerInfo,
	infoDelay time.Duration,
	settingsDelay time.Duration,
	ServerInfoReceiveTimeout time.Duration,
	t *testing.T,
) dmock.ClienTransport {
	return dmock.NewClienTransport().RegisterSetReceiveDeadline(
		func(deadline time.Time) (err error) {
			return nil
		},
	).RegisterReceiveServerInfo(
		func() (i delegate.ServerInfo, err error) {
			return serverInfo, nil
		},
	).RegisterSetReceiveDeadline(
		func(deadline time.Time) (err error) {
			return nil
		},
	)
}

func testDelegateCreation(serverInfo delegate.ServerInfo,
	transport delegate.ClienTransport[any],
	ops []SetOption,
	wantErr error,
	mocks []*mok.Mock,
	t *testing.T,
) {
	_, err := New(serverInfo, transport, ops...)
	if err != wantErr {
		t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
	}
	if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
		t.Error(infomap)
	}
}
