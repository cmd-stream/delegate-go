package client_test

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cmd-stream/core-go"
	ccln "github.com/cmd-stream/core-go/client"
	cmock "github.com/cmd-stream/core-go/test/mock"
	"github.com/cmd-stream/delegate-go"
	dcln "github.com/cmd-stream/delegate-go/client"
	clnmock "github.com/cmd-stream/delegate-go/test/mock/client"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestReconnectDelegate(t *testing.T) {
	var (
		ops        = []dcln.SetOption{dcln.WithServerInfoReceiveDuration(0)}
		serverInfo = delegate.ServerInfo([]byte("server info"))
	)

	t.Run("NewReconnect should check ServerInfo",
		func(t *testing.T) {
			var (
				wantErr   error = nil
				transport       = makeClientTransport(serverInfo)
				factory         = clnmock.NewTransportFactory().RegisterNew(
					func() (dcln.Transport[any], error) {
						return transport, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock, factory.Mock}
			)
			_, err := dcln.NewReconnect(serverInfo, factory, ops...)
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("If ServerInfo check fails with an error, NewReconnect should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("SetReceiveDeadline error")
				transport = clnmock.NewTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					})
				factory = clnmock.NewTransportFactory().RegisterNew(
					func() (dcln.Transport[any], error) {
						return transport, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock, factory.Mock}
			)
			_, err := dcln.NewReconnect(serverInfo, factory, ops...)
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("If ClientTransportFactory.New fails with an error, NewReconnect should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("transport creation error")
				factory = clnmock.NewTransportFactory().RegisterNew(
					func() (dcln.Transport[any], error) {
						return nil, wantErr
					},
				)
				mocks = []*mok.Mock{factory.Mock}
			)
			_, err := dcln.NewReconnect(serverInfo, factory, ops...)
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("Reconnect should work correctly", func(t *testing.T) {
		var (
			transport1    = makeClientTransport(serverInfo)
			wantTransport = makeClientTransport(serverInfo)
			factory       = clnmock.NewTransportFactory().RegisterNew(
				func() (dcln.Transport[any], error) {
					return transport1, nil
				},
			).RegisterNew(
				func() (dcln.Transport[any], error) {
					return nil, errors.New("transport creation error")
				},
			).RegisterNew(
				func() (dcln.Transport[any], error) {
					return nil, errors.New("transport creation error")
				},
			).RegisterNew(
				func() (dcln.Transport[any], error) {
					return wantTransport, nil
				},
			)
			mocks = []*mok.Mock{transport1.Mock, wantTransport.Mock, factory.Mock}
		)
		delegate, _ := dcln.NewReconnect(serverInfo, factory, ops...)
		delegate.Reconnect()
		transport := delegate.Transport()
		if transport != wantTransport {
			t.Errorf("unexpected transport, want '%v' actual '%v'", wantTransport,
				transport)
		}
		asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
	})

	t.Run("Reconnect should return ErrClosed, if the delegate is closed",
		func(t *testing.T) {
			var (
				wantErr    = ccln.ErrClosed
				transport1 = makeClientTransport(serverInfo).RegisterClose(
					func() (err error) { return nil },
				)
				factory = clnmock.NewTransportFactory().RegisterNew(
					func() (dcln.Transport[any], error) {
						return transport1, nil
					},
				).RegisterNew(
					func() (dcln.Transport[any], error) {
						time.Sleep(100 * time.Millisecond)
						return nil, errors.New("transport creation error")
					},
				)
				mocks = []*mok.Mock{transport1.Mock, factory.Mock}
			)
			delegate, _ := dcln.NewReconnect(serverInfo, factory, ops...)
			go func() {
				time.Sleep(50 * time.Millisecond)
				if err := delegate.Close(); err != nil {
					panic(err)
				}
			}()
			err := delegate.Reconnect()
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("If Transport.Send fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantN   = 1
				wantErr = errors.New("Delegate.Send error")
				clnTran = clnmock.NewTransport().RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						return wantN, wantErr
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				mocks    = []*mok.Mock{clnTran.Mock}
				delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
					dcln.Options{})
			)
			n, err := delegate.Send(1, cmock.NewCmd())
			asserterror.Equal(t, n, wantN)
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("Transport.Send should send same seq and cmd as Send",
		func(t *testing.T) {
			var (
				wantSeq core.Seq = 1
				wantCmd          = cmock.NewCmd()
				wantN            = 2
				wantErr error    = nil
				clnTran          = clnmock.NewTransport().RegisterSend(
					func(seq core.Seq, cmd core.Cmd[any]) (n int, err error) {
						asserterror.Equal(t, seq, wantSeq)
						asserterror.EqualDeep(t, cmd, wantCmd)
						return wantN, wantErr
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				mocks    = []*mok.Mock{clnTran.Mock}
				delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
					dcln.Options{})
			)
			n, err := delegate.Send(wantSeq, wantCmd)
			asserterror.Equal(t, n, wantN)
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("Receive should return same seq and cmd as Tranposrt.Receive",
		func(t *testing.T) {
			var (
				wantSeq    core.Seq = 1
				wantResult          = cmock.NewResult()
				wantN               = 3
				wantErr             = errors.New("receive failed")
				clnTran             = clnmock.NewTransport().RegisterReceive(
					func() (seq core.Seq, r core.Result, n int, err error) {
						return wantSeq, wantResult, wantN, wantErr
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				mocks    = []*mok.Mock{clnTran.Mock}
				delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
					dcln.Options{})
			)
			seq, result, n, err := delegate.Receive()
			asserterror.Equal(t, seq, wantSeq)
			asserterror.EqualDeep(t, result, wantResult)
			asserterror.Equal(t, n, wantN)
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("Conf should return the options that was obtained during creation",
		func(t *testing.T) {
			var (
				wantOptions = dcln.Options{}
				delegate    = dcln.NewReconnectWithoutInfo[any](nil, nil, nil,
					wantOptions)
			)
			options := delegate.Options()
			asserterror.EqualDeep(t, options, wantOptions)
		})

	t.Run("LocalAddr should return Transport.LocalAddr", func(t *testing.T) {
		var (
			wantAddr = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			clnTran  = clnmock.NewTransport().RegisterLocalAddr(
				func() (a net.Addr) { return wantAddr },
			)
			tran = &atomic.Value{}
		)
		tran.Store(clnTran)
		var (
			mocks    = []*mok.Mock{clnTran.Mock}
			delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
				dcln.Options{})
		)
		addr := delegate.LocalAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
		asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
	})

	t.Run("RemoteAddr should return Transport.RemoteAddr", func(t *testing.T) {
		var (
			wantAddr = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			clnTran  = clnmock.NewTransport().RegisterRemoteAddr(
				func() (addr net.Addr) { return wantAddr },
			)
			tran = &atomic.Value{}
		)
		tran.Store(clnTran)
		var (
			mocks    = []*mok.Mock{clnTran.Mock}
			delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
				dcln.Options{})
		)
		addr := delegate.RemoteAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
		asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
	})

	t.Run("If Tranposrt.Close fails with an error, Close should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("Close error")
				clnTran = clnmock.NewTransport().RegisterClose(
					func() (err error) {
						return wantErr
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				mocks    = []*mok.Mock{clnTran.Mock}
				delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
					dcln.Options{})
			)
			err := delegate.Close()
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("SetSendDeadline should call corresponding Transport.SetSendDeadline",
		func(t *testing.T) {
			var (
				wantErr     = errors.New("SetSendDeadline error")
				wantDeadine = time.Now()
				clnTran     = clnmock.NewTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						asserterror.Equal(t, deadline, wantDeadine)
						return wantErr
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				mocks    = []*mok.Mock{clnTran.Mock}
				delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
					dcln.Options{})
			)
			err := delegate.SetSendDeadline(wantDeadine)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v' ", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Flush should call corresponding Transport.Flush", func(t *testing.T) {
		var (
			wantErr = errors.New("Flush error")
			clnTran = clnmock.NewTransport().RegisterFlush(
				func() (err error) { return wantErr },
			)
			tran = &atomic.Value{}
		)
		tran.Store(clnTran)
		var (
			mocks    = []*mok.Mock{clnTran.Mock}
			delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran, dcln.Options{})
		)
		err := delegate.Flush()
		asserterror.EqualError(t, err, wantErr)
		asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
	})

	t.Run("SetReceiveDeadline should call corresponding Transport.SetReceiveDeadline",
		func(t *testing.T) {
			var (
				wantErr     = errors.New("SetReceiveDeadline error")
				wantDeadine = time.Now()
				clnTran     = clnmock.NewTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						if deadline != wantDeadine {
							return fmt.Errorf("unexpected deadline %v, want %v", deadline,
								wantDeadine)
						}
						return wantErr
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				mocks    = []*mok.Mock{clnTran.Mock}
				delegate = dcln.NewReconnectWithoutInfo[any](nil, nil, tran,
					dcln.Options{})
			)
			err := delegate.SetReceiveDeadline(wantDeadine)
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("If ServerInfo check fails with the ErrServerInfoMismatch, Reconnect should return it",
		func(t *testing.T) {
			var (
				wantErr = dcln.ErrServerInfoMismatch
				clnTran = clnmock.NewTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterReceiveServerInfo(
					func() (info delegate.ServerInfo, err error) {
						return []byte("different info"), nil
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				factory = clnmock.NewTransportFactory().RegisterNew(
					func() (dcln.Transport[any], error) {
						return clnTran, nil
					},
				)
				closeFlag uint32
				mocks     = []*mok.Mock{clnTran.Mock}
				delegate  = dcln.NewReconnectWithoutInfo(factory, &closeFlag, tran,
					dcln.Options{})
			)
			err := delegate.Reconnect()
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})

	t.Run("If ServerInfo check fails with an error, Reconnect should try again",
		func(t *testing.T) {
			var (
				wantErr   = ccln.ErrClosed
				closeFlag uint32
				clnTran   = clnmock.NewTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						closeFlag = 1
						return errors.New("SetReceiveDeadline error")
					},
				)
				tran = &atomic.Value{}
			)
			tran.Store(clnTran)
			var (
				factory = clnmock.NewTransportFactory().RegisterNew(
					func() (dcln.Transport[any], error) {
						return clnTran, nil
					},
				)
				mocks    = []*mok.Mock{clnTran.Mock}
				delegate = dcln.NewReconnectWithoutInfo(factory, &closeFlag, tran,
					dcln.Options{})
			)
			err := delegate.Reconnect()
			asserterror.EqualError(t, err, wantErr)
			asserterror.EqualDeep(t, mok.CheckCalls(mocks), mok.EmptyInfomap)
		})
}
