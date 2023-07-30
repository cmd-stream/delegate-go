package client

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	base_client "github.com/cmd-stream/base-go/client"
	base_mock "github.com/cmd-stream/base-go/testdata/mock"
	"github.com/cmd-stream/delegate-go"
	"github.com/cmd-stream/delegate-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

func TestReconnectDelegate(t *testing.T) {

	var (
		conf = Conf{
			SysDataReceiveTimeout: 0,
		}
		serverInfo     = delegate.ServerInfo([]byte("server info"))
		serverSettings = delegate.ServerSettings{MaxCmdSize: 500}
	)

	t.Run("NewReconnect should check ServerInfo and ServerSettings",
		func(t *testing.T) {
			var (
				wantErr   error = nil
				transport       = MakeClientTransport(time.Now(), serverInfo,
					200*time.Millisecond,
					serverSettings,
					300*time.Millisecond,
					time.Second,
					t)
				factory = mock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock, factory.Mock}
			)
			testReconnectDelegateCreation(conf, serverInfo, factory, wantErr, mocks,
				t)
		})

	t.Run("If ServerInfo check fails with an error, NewReconnect should return it",
		func(t *testing.T) {
			var (
				wantErr   error = errors.New("SetReceiveDeadline error")
				transport       = mock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					})
				factory = mock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock, factory.Mock}
			)
			testReconnectDelegateCreation(conf, serverInfo, factory, wantErr, mocks,
				t)
		})

	t.Run("If ServerSettings apply fails with an error, NewReconnect should return it",
		func(t *testing.T) {
			var (
				wantErr   error = errors.New("SetReceiveDeadline error")
				transport       = mock.NewClienTransport().RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterReceiveServerInfo(
					func() (i delegate.ServerInfo, err error) {
						return serverInfo, nil
					},
				).RegisterSetReceiveDeadline(
					func(deadline time.Time) (err error) { return wantErr },
				)
				factory = mock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				mocks = []*mok.Mock{transport.Mock, factory.Mock}
			)
			testReconnectDelegateCreation(conf, serverInfo, factory, wantErr, mocks,
				t)
		})

	t.Run("If ClientTransportFactory.New fails with an error, NewReconnect should return it",
		func(t *testing.T) {
			var (
				wantErr error = errors.New("transport creation error")
				factory       = mock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return nil, wantErr
					},
				)
				mocks = []*mok.Mock{factory.Mock}
			)
			testReconnectDelegateCreation(conf, serverInfo, factory, wantErr, mocks,
				t)
		})

	t.Run("Reconnect should work correctly", func(t *testing.T) {
		var (
			transport1 = MakeClientTransport(time.Now(), serverInfo,
				200*time.Millisecond,
				serverSettings,
				300*time.Millisecond,
				time.Second,
				t)
			wantTransport = MakeClientTransport(time.Now(), serverInfo,
				200*time.Millisecond,
				serverSettings,
				300*time.Millisecond,
				time.Second,
				t)
			factory = mock.NewClienTransportFactory().RegisterNew(
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
		delegate, _ := NewReconnect[any](conf, serverInfo, factory)
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
				wantErr    = base_client.ErrClosed
				transport1 = MakeClientTransport(time.Now(), serverInfo,
					200*time.Millisecond,
					serverSettings,
					300*time.Millisecond,
					time.Second,
					t).RegisterClose(
					func() (err error) { return nil },
				)
				factory = mock.NewClienTransportFactory().RegisterNew(
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
			delegate, _ := NewReconnect[any](conf, serverInfo, factory)
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
				wantErr   = errors.New("Delegate.Send error")
				transport = func() (transport atomic.Value) {
					transport.Store(mock.NewClienTransport().RegisterSend(
						func(seq base.Seq, cmd base.Cmd[any]) (err error) {
							return wantErr
						},
					))
					return
				}()
				mocks    = []*mok.Mock{transport.Load().(mock.ClienTransport).Mock}
				delegate = ReconnectDelegate[any]{transport: transport}
			)
			err := delegate.Send(1, base_mock.NewCmd())
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
				wantSeq   base.Seq = 1
				wantCmd            = base_mock.NewCmd()
				transport          = func() (transport atomic.Value) {
					transport.Store(mock.NewClienTransport().RegisterSend(
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
					))
					return
				}()
				mocks    = []*mok.Mock{transport.Load().(mock.ClienTransport).Mock}
				delegate = ReconnectDelegate[any]{transport: transport}
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
				wantResult          = base_mock.NewResult()
				transport           = func() (transport atomic.Value) {
					transport.Store(mock.NewClienTransport().RegisterReceive(
						func() (seq base.Seq, r base.Result, err error) {
							return wantSeq, wantResult, wantErr
						},
					))
					return
				}()
				mocks    = []*mok.Mock{transport.Load().(mock.ClienTransport).Mock}
				delegate = ReconnectDelegate[any]{transport: transport}
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

	t.Run("Conf should return the conf that was obtained during creation",
		func(t *testing.T) {
			var (
				wantConf = conf
				delegate = ReconnectDelegate[any]{conf: wantConf}
			)
			conf := delegate.Conf()
			if conf != wantConf {
				t.Errorf("unexpected conf, want '%v' actual '%v'", wantConf, conf)
			}
		})

	t.Run("LocalAddr should return Transport.LocalAddr", func(t *testing.T) {
		var (
			wantAddr  = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			transport = func() (transport atomic.Value) {
				transport.Store(mock.NewClienTransport().RegisterLocalAddr(
					func() (a net.Addr) {
						return wantAddr
					},
				))
				return
			}()
			delegate = ReconnectDelegate[any]{transport: transport}
		)
		addr := delegate.LocalAddr()
		if addr != wantAddr {
			t.Errorf("unexpected addr, want '%v' actual '%v'", wantAddr, addr)
		}
	})

	t.Run("RemoteAddr should return Transport.RemoteAddr", func(t *testing.T) {
		var (
			wantAddr  = &net.IPAddr{IP: net.ParseIP("127.0.0.1:9000")}
			transport = func() (transport atomic.Value) {
				transport.Store(mock.NewClienTransport().RegisterRemoteAddr(
					func() (addr net.Addr) {
						return wantAddr
					},
				))
				return
			}()
			delegate = ReconnectDelegate[any]{transport: transport}
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
				transport = func() (transport atomic.Value) {
					transport.Store(mock.NewClienTransport().RegisterClose(
						func() (err error) {
							return wantErr
						},
					))
					return
				}()
				delegate = ReconnectDelegate[any]{transport: transport}
			)
			err := delegate.Close()
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
		})

	t.Run("SetSendDeadline should call corresponding Transport.SetSendDeadline",
		func(t *testing.T) {
			var (
				wantErr        = errors.New("SetSendDeadline error")
				wantDeadine    = time.Now()
				val, transport = func() (val atomic.Value, transport mock.ClienTransport) {
					transport = mock.NewClienTransport().RegisterSetSendDeadline(
						func(deadline time.Time) (err error) {
							if deadline != wantDeadine {
								return fmt.Errorf("unexpected deadline %v, want %v", deadline,
									wantDeadine)
							}
							return wantErr
						},
					)
					val.Store(transport)
					return
				}()
				mocks    = []*mok.Mock{transport.Mock}
				delegate = ReconnectDelegate[any]{transport: val}
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
			wantErr        = errors.New("Flush error")
			val, transport = func() (val atomic.Value, transport mock.ClienTransport) {
				transport = mock.NewClienTransport().RegisterFlush(
					func() (err error) { return wantErr },
				)
				val.Store(transport)
				return
			}()
			mocks    = []*mok.Mock{transport.Mock}
			delegate = ReconnectDelegate[any]{transport: val}
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
				wantErr        = errors.New("SetReceiveDeadline error")
				wantDeadine    = time.Now()
				val, transport = func() (val atomic.Value, transport mock.ClienTransport) {
					transport = mock.NewClienTransport().RegisterSetReceiveDeadline(
						func(deadline time.Time) (err error) {
							if deadline != wantDeadine {
								return fmt.Errorf("unexpected deadline %v, want %v", deadline,
									wantDeadine)
							}
							return wantErr
						},
					)
					val.Store(transport)
					return
				}()
				mocks    = []*mok.Mock{transport.Mock}
				delegate = ReconnectDelegate[any]{transport: val}
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
				wantErr        = ErrServerInfoMismatch
				val, transport = func() (val atomic.Value, transport mock.ClienTransport) {
					transport = mock.NewClienTransport().RegisterSetReceiveDeadline(
						func(deadline time.Time) (err error) {
							return nil
						},
					).RegisterReceiveServerInfo(
						func() (info delegate.ServerInfo, err error) {
							return []byte("different info"), nil
						},
					)
					val.Store(transport)
					return
				}()
				factory = mock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				closeFlag uint32
				mocks     = []*mok.Mock{transport.Mock}
				delegate  = ReconnectDelegate[any]{transport: val, factory: factory, closedFlag: &closeFlag}
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
				wantErr        = base_client.ErrClosed
				closeFlag      uint32
				val, transport = func() (val atomic.Value, transport mock.ClienTransport) {
					transport = mock.NewClienTransport().RegisterSetReceiveDeadline(
						func(deadline time.Time) (err error) {
							closeFlag = 1
							return errors.New("SetReceiveDeadline error")
						},
					)
					val.Store(transport)
					return
				}()
				factory = mock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				mocks    = []*mok.Mock{transport.Mock}
				delegate = ReconnectDelegate[any]{transport: val, factory: factory, closedFlag: &closeFlag}
			)
			err := delegate.Reconnect()
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v' ", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If ServerSettings apply fails with an error, Reconnect should try again",
		func(t *testing.T) {
			var (
				wantErr        = base_client.ErrClosed
				closeFlag      uint32
				val, transport = func() (val atomic.Value, transport mock.ClienTransport) {
					transport = mock.NewClienTransport().RegisterSetReceiveDeadline(
						func(deadline time.Time) (err error) {
							return nil
						},
					).RegisterReceiveServerInfo(
						func() (info delegate.ServerInfo, err error) {
							return serverInfo, nil
						},
					).RegisterSetReceiveDeadline(
						func(deadline time.Time) (err error) {
							closeFlag = 1
							return errors.New("SetReceiveDeadline error")
						},
					)
					val.Store(transport)
					return
				}()
				factory = mock.NewClienTransportFactory().RegisterNew(
					func() (delegate.ClienTransport[any], error) {
						return transport, nil
					},
				)
				mocks    = []*mok.Mock{transport.Mock}
				delegate = ReconnectDelegate[any]{info: serverInfo, transport: val,
					factory: factory, closedFlag: &closeFlag}
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

func testReconnectDelegateCreation(conf Conf, serverInfo delegate.ServerInfo,
	factory delegate.ClienTransportFactory[any],
	wantErr error,
	mocks []*mok.Mock,
	t *testing.T,
) {
	_, err := NewReconnect(conf, serverInfo, factory)
	if err != wantErr {
		t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
	}
	if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
		t.Error(infomap)
	}
}
