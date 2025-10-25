package server

import (
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	var (
		o                          = Options{}
		wantServerInfoSendDuration = time.Second
	)
	Apply([]SetOption{
		WithServerInfoSendDuration(wantServerInfoSendDuration),
	}, &o)

	if o.ServerInfoSendDuration != wantServerInfoSendDuration {
		t.Errorf("unexpected ServerInfoSendDuration, want %v actual %v",
			wantServerInfoSendDuration, o.ServerInfoSendDuration)
	}
}
