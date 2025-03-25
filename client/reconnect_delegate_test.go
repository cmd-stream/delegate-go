package dcln

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	bcln "github.com/cmd-stream/base-go/client"
	bmock "github.com/cmd-stream/base-go/testdata/mock"
	"github.com/cmd-stream/delegate-go"
	dmock "github.com/cmd-stream/delegate-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

func TestReconnectDelegate(t *testing.T) {

	var (
		ops = []SetOption{
			WithServerInfoReceiveDuration(0),
		}
		serverInfo = delegate.ServerInfo([]byte("server info"))
	)

	t.Run("NewReconnect should check ServerInfo",
		func(t *testing.T) {
			var (
				wantErr   error = nil
				transport       = MakeClientTransport(time.Now(), serverInfo,
					200*time.Millisecond,
					300*time.Millisecond,
					time.Second,
					t)
				factory = dmock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock, factory.Mock}
			)
			testReconnectDelegateCreation(serverInfo, factory, ops, wantErr, mocks,
				t)
		})

	t.Run("If ServerInfo check fails with an error, NewReconnect should return it",
		func(t *testing.T) {
			var (
				wantErr   error = errors.New("SetReceiveDeadline error")
				transport       = dmock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					})
				factory = dmock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock, factory.Mock}
			)
			testReconnectDelegateCreation(serverInfo, factory, ops, wantErr, mocks,
				t)
		})

	t.Run("If ClientTransportFactory.New fails with an error, NewReconnect should return it",
		func(t *testing.T) {
			var (
				wantErr error = errors.New("transport creation error")
				factory       = dmock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return nil, wantErr
					},
				)
				mocks = []*mok.Mock{factory.Mock}
			)
			testReconnectDelegateCreation(serverInfo, factory, ops, wantErr, mocks,
				t)
		})

	t.Run("Reconnect should work correctly", func(t *testing.T) {
		var (
			transport1 = MakeClientTransport(time.Now(), serverInfo,
				200*time.Millisecond,
				300*time.Millisecond,
				time.Second,
				t)
			wantTransport = MakeClientTransport(time.Now(), serverInfo,
				200*time.Millisecond,
				300*time.Millisecond,
				time.Second,
				t)
			factory = dmock.NewClienTransportFactory().RegisterNew(
				func() (delegate.ClienTransport[any], error) {
					return transport1, nil
				},
			).RegisterNew(
				func() (delegate.ClienTransport[any], error) {
					return nil, errors.New("transport creation error")
				},
			).RegisterNew(
				func() (delegate.ClienTransport[any], error) {
					return nil, errors.New("transport creation error")
				},
			).RegisterNew(
				func() (delegate.ClienTransport[any], error) {
					return wantTransport, nil
				},
			)
			mocks = []*mok.Mock{transport1.Mock, wantTransport.Mock, factory.Mock}
		)
		delegate, _ := NewReconnect[any](serverInfo, factory, ops...)
		delegate.Reconnect()
		transport := delegate.getTransport()
		if transport != wantTransport {
			t.Errorf("unexpected transport, want '%v' actual '%v'", wantTransport,
				transport)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("Reconnect should return ErrClosed, if the delegate is closed",
		func(t *testing.T) {
			var (
				wantErr    = bcln.ErrClosed
				transport1 = MakeClientTransport(time.Now(), serverInfo,
					200*time.Millisecond,
					300*time.Millisecond,
					time.Second,
					t).RegisterClose(
					func() (err error) { return nil },
				)
				factory = dmock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport1, nil
					},
				).RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						time.Sleep(100 * time.Millisecond)
						return nil, errors.New("transport creation error")
					},
				)
				mocks = []*mok.Mock{transport1.Mock, factory.Mock}
			)
			delegate, _ := NewReconnect[any](serverInfo, factory, ops...)
			go func() {
				time.Sleep(50 * time.Millisecond)
				if err := delegate.Close(); err != nil {
					panic(err)
				}
			}()
			err := delegate.Reconnect()
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If Transport.Send fails with an error, Send should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("Delegate.Send error")
				ctn     = dmock.NewClienTransport().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						return wantErr
					},
				)
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				mocks    = []*mok.Mock{ctn.Mock}
				delegate = ReconnectDelegate[any]{transport: tn}
			)
			err := delegate.Send(1, bmock.NewCmd())
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Transport.Send should send same seq and cmd as Send",
		func(t *testing.T) {
			var (
				wantSeq base.Seq = 1
				wantCmd          = bmock.NewCmd()
				ctn              = dmock.NewClienTransport().RegisterSend(
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
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				mocks    = []*mok.Mock{ctn.Mock}
				delegate = ReconnectDelegate[any]{transport: tn}
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
				ctn                 = dmock.NewClienTransport().RegisterReceive(
					func() (seq base.Seq, r base.Result, err error) {
						return wantSeq, wantResult, wantErr
					},
				)
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				mocks    = []*mok.Mock{ctn.Mock}
				delegate = ReconnectDelegate[any]{transport: tn}
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
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Conf should return the options that was obtained during creation",
		func(t *testing.T) {
			var (
				wantOptions = Options{}
				delegate    = ReconnectDelegate[any]{options: wantOptions}
			)
			options := delegate.Options()
			if options != wantOptions {
				t.Errorf("unexpected options, want '%v' actual '%v'", wantOptions,
					options)
			}
		})

	t.Run("LocalAddr should return Transport.LocalAddr", func(t *testing.T) {
		var (
			wantAddr = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			ctn      = dmock.NewClienTransport().RegisterLocalAddr(
				func() (a net.Addr) {
					return wantAddr
				},
			)
			tn = &atomic.Value{}
		)
		tn.Store(ctn)
		var (
			mocks    = []*mok.Mock{ctn.Mock}
			delegate = ReconnectDelegate[any]{transport: tn}
		)
		addr := delegate.LocalAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("RemoteAddr should return Transport.RemoteAddr", func(t *testing.T) {
		var (
			wantAddr = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			ctn      = dmock.NewClienTransport().RegisterRemoteAddr(
				func() (addr net.Addr) {
					return wantAddr
				},
			)
			tn = &atomic.Value{}
		)
		tn.Store(ctn)
		var (
			mocks    = []*mok.Mock{ctn.Mock}
			delegate = ReconnectDelegate[any]{transport: tn}
		)
		addr := delegate.RemoteAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("If Tranposrt.Close fails with an error, Close should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("Close error")
				ctn     = dmock.NewClienTransport().RegisterClose(
					func() (err error) {
						return wantErr
					},
				)
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				mocks    = []*mok.Mock{ctn.Mock}
				delegate = ReconnectDelegate[any]{transport: tn}
			)

			err := delegate.Close()
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("SetSendDeadline should call corresponding Transport.SetSendDeadline",
		func(t *testing.T) {
			var (
				wantErr     = errors.New("SetSendDeadline error")
				wantDeadine = time.Now()
				ctn         = dmock.NewClienTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						if deadline != wantDeadine {
							return fmt.Errorf("unexpected deadline %v, want %v", deadline,
								wantDeadine)
						}
						return wantErr
					},
				)
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				mocks    = []*mok.Mock{ctn.Mock}
				delegate = ReconnectDelegate[any]{transport: tn}
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
			ctn     = dmock.NewClienTransport().RegisterFlush(
				func() (err error) { return wantErr },
			)
			tn = &atomic.Value{}
		)
		tn.Store(ctn)
		var (
			mocks    = []*mok.Mock{ctn.Mock}
			delegate = ReconnectDelegate[any]{transport: tn}
		)
		err := delegate.Flush()
		if err != wantErr {
			t.Errorf("unexpected error, want '%v' actual '%v' ", wantErr, err)
		}
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("SetReceiveDeadline should call corresponding Transport.SetReceiveDeadline",
		func(t *testing.T) {
			var (
				wantErr     = errors.New("SetReceiveDeadline error")
				wantDeadine = time.Now()
				ctn         = dmock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						if deadline != wantDeadine {
							return fmt.Errorf("unexpected deadline %v, want %v", deadline,
								wantDeadine)
						}
						return wantErr
					},
				)
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				mocks    = []*mok.Mock{ctn.Mock}
				delegate = ReconnectDelegate[any]{transport: tn}
			)
			err := delegate.SetReceiveDeadline(wantDeadine)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v' ", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If ServerInfo check fails with the ErrServerInfoMismatch, Reconnect should return it",
		func(t *testing.T) {
			var (
				wantErr = ErrServerInfoMismatch
				ctn     = dmock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterReceiveServerInfo(
					func() (info delegate.ServerInfo, err error) {
						return []byte("different info"), nil
					},
				)
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				factory = dmock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return ctn, nil
					},
				)
				closeFlag uint32
				mocks     = []*mok.Mock{ctn.Mock}
				delegate  = ReconnectDelegate[any]{
					transport:  tn,
					factory:    factory,
					closedFlag: &closeFlag,
				}
			)
			err := delegate.Reconnect()
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v' ", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If ServerInfo check fails with an error, Reconnect should try again",
		func(t *testing.T) {
			var (
				wantErr   = bcln.ErrClosed
				closeFlag uint32
				ctn       = dmock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						closeFlag = 1
						return errors.New("SetReceiveDeadline error")
					},
				)
				tn = &atomic.Value{}
			)
			tn.Store(ctn)
			var (
				factory = dmock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return ctn, nil
					},
				)
				mocks    = []*mok.Mock{ctn.Mock}
				delegate = ReconnectDelegate[any]{
					transport:  tn,
					factory:    factory,
					closedFlag: &closeFlag,
				}
			)
			err := delegate.Reconnect()
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v' ", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

}

func testReconnectDelegateCreation(serverInfo delegate.ServerInfo,
	factory delegate.ClienTransportFactory[any],
	ops []SetOption,
	wantErr error,
	mocks []*mok.Mock,
	t *testing.T,
) {
	_, err := NewReconnect(serverInfo, factory, ops...)
	if err != wantErr {
		t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
	}
	if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
		t.Error(infomap)
	}
}
