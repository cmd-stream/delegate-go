package client_test

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/cmd-stream/core-go"
	cmock "github.com/cmd-stream/core-go/testdata/mock"
	"github.com/cmd-stream/delegate-go"
	dcln "github.com/cmd-stream/delegate-go/client"
	mock "github.com/cmd-stream/delegate-go/client/testdata/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestDelegate(t *testing.T) {
	var (
		delta      = 100 * time.Millisecond
		ops        = []dcln.SetOption{dcln.WithServerInfoReceiveDuration(0)}
		serverInfo = delegate.ServerInfo([]byte("server info"))
	)

	t.Run("New should check ServerInfo", func(t *testing.T) {
		var (
			wantErr   error = nil
			conn            = cmock.NewConn()
			transport       = makeClientTransport(serverInfo)
			mocks           = []*mok.Mock{conn.Mock, transport.Mock}
		)
		_, err := dcln.New[any](serverInfo, transport, ops...)
		asserterror.EqualError(err, wantErr, t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("If Transport.SetReceiveDeadline fails with an error before receive ServerInfo, New should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.SetReceiveDeadline")
				transport = mock.NewTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				)
				mocks = []*mok.Mock{transport.Mock}
			)
			_, err := dcln.New(serverInfo, transport, ops...)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Transport.ReceiveServerInfo fails with an error, New should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("Transport.ReceiveServerInfo error")
				transport = mock.NewTransport().RegisterSetReceiveDeadline(
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
			_, err := dcln.New(serverInfo, transport, ops...)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If wrong ServerInfo was received, New should return error",
		func(t *testing.T) {
			var (
				wantErr         = dcln.ErrServerInfoMismatch
				wrongServerInfo = []byte{1}
				transport       = mock.NewTransport().RegisterSetReceiveDeadline(
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
			_, err := dcln.New(serverInfo, transport, ops...)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("New should apply Conf.ServerInfoReceiveDuration", func(t *testing.T) {
		var (
			d                  = time.Second
			wantDeadline       = time.Now().Add(d)
			wantErr      error = nil
			ops                = []dcln.SetOption{dcln.WithServerInfoReceiveDuration(d)}
			transport          = mock.NewTransport().RegisterSetReceiveDeadline(
				func(deadline time.Time) (err error) {
					asserterror.SameTime(deadline, wantDeadline, delta, t)
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
		_, err := dcln.New(serverInfo, transport, ops...)
		asserterror.EqualError(err, wantErr, t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("If Transport.Send fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantN     int = 1
				wantErr       = errors.New("Delegate.Send error")
				transport     = mock.NewTransport().RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						return wantN, wantErr
					},
				)
				delegate = dcln.NewWithoutInfo(transport)
			)
			n, err := delegate.Send(1, cmock.NewCmd())
			asserterror.Equal(n, wantN, t)
			asserterror.EqualError(err, wantErr, t)
		})

	t.Run("Transport.Send should send same seq and cmd as Send",
		func(t *testing.T) {
			var (
				wantSeq   core.Seq = 1
				wantCmd            = cmock.NewCmd()
				wantN     int      = 2
				wantErr   error    = nil
				transport          = mock.NewTransport().RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						asserterror.Equal(seq, wantSeq, t)
						asserterror.EqualDeep(cmd, wantCmd, t)
						return wantN, wantErr
					},
				)
				mocks    = []*mok.Mock{transport.Mock}
				delegate = dcln.NewWithoutInfo(transport)
			)
			n, err := delegate.Send(wantSeq, wantCmd)
			asserterror.Equal(n, wantN, t)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("Receive should return same seq and cmd as Tranposrt.Receive",
		func(t *testing.T) {
			var (
				wantSeq    core.Seq = 1
				wantResult          = cmock.NewResult()
				wantN      int      = 3
				wantErr             = errors.New("receive failed")
				transport           = mock.NewTransport().RegisterReceive(
					func() (seq core.Seq, r core.Result, n int, err error) {
						return wantSeq, wantResult, wantN, wantErr
					},
				)
				mocks    = []*mok.Mock{transport.Mock}
				delegate = dcln.NewWithoutInfo(transport)
			)
			seq, result, n, err := delegate.Receive()
			asserterror.Equal(seq, wantSeq, t)
			asserterror.EqualDeep(result, wantResult, t)
			asserterror.Equal(n, wantN, t)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("Conf should return the conf that was obtained during creation",
		func(t *testing.T) {
			var (
				wantO    = dcln.Options{}
				delegate = dcln.Delegate[any]{}
			)
			o := delegate.Options()
			asserterror.Equal(o, wantO, t)
		})

	t.Run("LocalAddr should return Transport.LocalAddr", func(t *testing.T) {
		var (
			wantAddr  = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			transport = mock.NewTransport().RegisterLocalAddr(
				func() (a net.Addr) {
					return wantAddr
				},
			)
			delegate = dcln.NewWithoutInfo(transport)
		)
		addr := delegate.LocalAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
	})

	t.Run("RemoteAddr should return Transport.RemoteAddr", func(t *testing.T) {
		var (
			wantAddr  = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			transport = mock.NewTransport().RegisterRemoteAddr(
				func() (addr net.Addr) {
					return wantAddr
				},
			)
			delegate = dcln.NewWithoutInfo(transport)
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
				transport = mock.NewTransport().RegisterClose(
					func() (err error) {
						return wantErr
					},
				)
				delegate = dcln.NewWithoutInfo(transport)
			)
			err := delegate.Close()
			asserterror.EqualError(err, wantErr, t)
		})
}

func makeClientTransport(serverInfo delegate.ServerInfo) mock.Transport {
	return mock.NewTransport().RegisterSetReceiveDeadline(
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
