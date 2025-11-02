package server_test

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/cmd-stream/delegate-go"
	dsrv "github.com/cmd-stream/delegate-go/server"
	cmocks "github.com/cmd-stream/testkit-go/mocks/core"
	mocks "github.com/cmd-stream/testkit-go/mocks/delegate/server"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestDelegate(t *testing.T) {
	var (
		delta                      = 100 * time.Millisecond
		wantServerInfoSendDuration = time.Second
		ops                        = []dsrv.SetOption{
			dsrv.WithServerInfoSendDuration(wantServerInfoSendDuration),
		}
		serverInfo = delegate.ServerInfo([]byte("server info"))
	)

	t.Run("If ServerInfo len is zero, New should panic",
		func(t *testing.T) {
			wantErr := dsrv.ErrEmptyInfo
			defer func() {
				if r := recover(); r != nil {
					err := r.(error)
					if err != wantErr {
						t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
					}
				}
			}()
			dsrv.New[any](nil, nil, nil, ops...)
		})

	t.Run("If send ServerInfo fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("send ServerInfo error")
				conn      = cmocks.NewConn()
				transport = mocks.NewTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return nil },
				).RegisterSendServerInfo(
					func(info delegate.ServerInfo) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				factory  = makeTransportFactory(conn, transport, t)
				delegate = dsrv.New(serverInfo, factory, nil, ops...)
				mocks    = []*mok.Mock{conn.Mock, transport.Mock, factory.Mock}
			)
			err := delegate.Handle(context.Background(), conn)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Transport.Handle fails with an error, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("done")
				conn      = cmocks.NewConn()
				transport = makeTransport(time.Now(), serverInfo,
					wantServerInfoSendDuration, delta, t)
				factory = makeTransportFactory(conn, transport, t)
				handler = mocks.NewTransportHandler().RegisterHandle(
					func(ctx context.Context, transport dsrv.Transport[any]) error {
						return wantErr
					},
				)
				delegate = dsrv.New(serverInfo, factory, handler, ops...)
				mocks    = []*mok.Mock{
					conn.Mock, transport.Mock, factory.Mock,
					handler.Mock,
				}
			)
			err := delegate.Handle(context.Background(), conn)
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Transport.SetSendDeadline fails with an error on ServerInfo send, Handle should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("SendServerInfo error")
				conn      = cmocks.NewConn()
				transport = mocks.NewTransport().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				factory  = makeTransportFactory(conn, transport, t)
				delegate = dsrv.New(serverInfo, factory, nil, ops...)
				err      = delegate.Handle(context.Background(), conn)
			)
			asserterror.EqualError(err, wantErr, t)
		})
}

func makeTransportFactory(conn net.Conn,
	transport dsrv.Transport[any],
	t *testing.T,
) mocks.TransportFactory {
	return mocks.NewTransportFactory().RegisterNew(
		func(c net.Conn) dsrv.Transport[any] {
			if !reflect.DeepEqual(conn, conn) {
				t.Errorf("unepxected conn, want '%v' actual '%v'", conn, c)
			}
			return transport
		},
	)
}

func makeTransport(startTime time.Time, info delegate.ServerInfo,
	serverInfoSendDuration time.Duration,
	delta time.Duration,
	t *testing.T,
) mocks.Transport {
	return mocks.NewTransport().RegisterSetSendDeadline(
		func(deadline time.Time) (err error) {
			wantDeadline := startTime.Add(serverInfoSendDuration)
			asserterror.SameTime(deadline, wantDeadline, delta, t)
			return nil
		},
	).RegisterSendServerInfo(
		func(i delegate.ServerInfo) (error error) {
			asserterror.EqualDeep(i, info, t)
			return nil
		},
	)
}
