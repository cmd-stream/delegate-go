package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	bmock "github.com/cmd-stream/base-go/testdata/mock"
	"github.com/cmd-stream/delegate-go"
	"github.com/cmd-stream/delegate-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

const Delta = 100 * time.Millisecond

func TestDelegate(t *testing.T) {

	var (
		conf = Conf{
			SysDataSendDuration: time.Second,
		}
		serverInfo     = delegate.ServerInfo([]byte("server info"))
		serverSettings = delegate.ServerSettings{MaxCmdSize: 500}
	)

	t.Run("If ServerInfo len is zero, New should panic",
		func(t *testing.T) {
			var wantErr = ErrEmptyInfo
			defer func() {
				if r := recover(); r != nil {
					err := r.(error)
					if err != wantErr {
						t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
					}
				}
			}()
			New[any](conf, nil, serverSettings, nil, nil)
		})

	t.Run("If send ServerInfo fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("send ServerInfo error")
				conn      = bmock.NewConn()
				transport = mock.NewServerTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSendServerInfo(
					func(info delegate.ServerInfo) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				transportFactory = MakeServerTransportFactory(conn, transport, t)
				handler          = New[any](conf, serverInfo, serverSettings,
					transportFactory,
					nil)
				mocks = []*mok.Mock{conn.Mock, transport.Mock, transportFactory.Mock}
			)
			testDelegate(context.Background(), conn, handler, wantErr, mocks, t)
		})

	t.Run("If send ServerSettings fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("send ServerSettings error")
				conn      = bmock.NewConn()
				transport = mock.NewServerTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSendServerInfo(
					func(info delegate.ServerInfo) (err error) { return nil },
				).RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSendServerSettings(
					func(settings delegate.ServerSettings) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				transportFactory = MakeServerTransportFactory(conn, transport,
					t)
				handler = New[any](conf, serverInfo, serverSettings,
					transportFactory,
					nil)
				mocks = []*mok.Mock{conn.Mock, transport.Mock, transportFactory.Mock}
			)
			testDelegate(context.Background(), conn, handler, wantErr, mocks, t)
		})

	t.Run("If Transport.Handle fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("done")
				conn      = bmock.NewConn()
				transport = MakeServerTransport(time.Now(), serverInfo,
					200*time.Millisecond,
					serverSettings,
					0,
					conf.SysDataSendDuration)
				transportFactory = MakeServerTransportFactory(conn,
					transport, t)
				transportHandler = mock.NewTransportHandler().RegisterHandle(
					func(ctx context.Context, transport delegate.ServerTransport[any]) error {
						return wantErr
					},
				)
				handler = New[any](conf, serverInfo,
					serverSettings,
					transportFactory,
					transportHandler)
				mocks = []*mok.Mock{conn.Mock, transport.Mock, transportFactory.Mock,
					transportHandler.Mock}
			)
			testDelegate(context.Background(), conn, handler, wantErr, mocks, t)
		})

	t.Run("If Transport.SetSendDeadline fails with an error on ServerInfo send, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("SendServerInfo error")
				conn      = bmock.NewConn()
				transport = mock.NewServerTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				factory = MakeServerTransportFactory(conn, transport, t)
				handler = Delegate[any]{conf: conf, factory: factory}
				err     = handler.Handle(context.Background(), conn)
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
		})

	t.Run("If Transport.SetSendDeadline fails with an error on ServerSettings send, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("SendServerSettings error")
				conn      = bmock.NewConn()
				transport = mock.NewServerTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSendServerInfo(
					func(info delegate.ServerInfo) (err error) { return nil },
				).RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				factory = MakeServerTransportFactory(conn, transport, t)
				handler = Delegate[any]{conf: conf, factory: factory}
				err     = handler.Handle(context.Background(), conn)
			)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
		})

}

func SameTime(t1, t2 time.Time) bool {
	return !(t1.Before(t2.Truncate(Delta)) || t1.After(t2.Add(Delta)))
}

func MakeServerTransportFactory(conn net.Conn,
	transport delegate.ServerTransport[any],
	t *testing.T,
) mock.ServerTransportFactory {
	return mock.NewServerTransportFactory().RegisterNew(
		func(c net.Conn) delegate.ServerTransport[any] {
			if !reflect.DeepEqual(conn, conn) {
				t.Errorf("unepxected conn, want '%v' actual '%v'", conn, c)
			}
			return transport
		},
	)
}

func MakeServerTransport(startTime time.Time, info delegate.ServerInfo,
	infoDelay time.Duration,
	settings delegate.ServerSettings,
	settingsDelay time.Duration,
	SysDataSendDuration time.Duration,
) mock.ServerTransport {
	return mock.NewServerTransport().RegisterSetSendDeadline(
		func(deadline time.Time) (err error) {
			wantDeadline := startTime.Add(SysDataSendDuration)
			if !SameTime(deadline, wantDeadline) {
				return fmt.Errorf("ServerTransport.SendServerInfo(), unepxected deadline, want '%v' actual '%v'",
					wantDeadline,
					deadline)
			}
			return nil
		},
	).RegisterSendServerInfo(
		func(i delegate.ServerInfo) (error error) {
			if !bytes.Equal(i, info) {
				return fmt.Errorf("ServerTransport.SendServerInfo(), unexpected info, want '%v' actual '%v'",
					info, i)
			}
			return nil
		},
	).RegisterSetSendDeadline(
		func(deadline time.Time) (err error) {
			wantDeadline := startTime.Add(SysDataSendDuration)
			if !SameTime(wantDeadline, deadline) {
				return fmt.Errorf("ServerTransport.SendServerSettings(), unepxected deadline, want '%v' actual '%v'",
					wantDeadline,
					deadline)
			}
			return nil
		},
	).RegisterSendServerSettings(
		func(s delegate.ServerSettings) (error error) {
			if !reflect.DeepEqual(s, settings) {
				return fmt.Errorf("ServerTransport.SendServerSettings(), unexpected settings, want '%v' actual '%v'",
					settings, s)
			}
			return nil
		},
	)
}

func testDelegate[T any](ctx context.Context, conn bmock.Conn,
	Delegate Delegate[T],
	wantErr error,
	mocks []*mok.Mock,
	t *testing.T,
) {
	err := Delegate.Handle(ctx, conn)
	if err != wantErr {
		t.Fatalf("unexpected error, want '%v' actual '%v'", wantErr, err)
	}
	if info := mok.CheckCalls(mocks); len(info) > 0 {
		t.Error(info)
	}
}
