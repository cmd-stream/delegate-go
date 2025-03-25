package dcln

import (
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	var (
		o                             = Options{}
		wantServerInfoReceiveDuration = time.Second
	)
	Apply([]SetOption{
		WithServerInfoReceiveDuration(wantServerInfoReceiveDuration),
	}, &o)

	if o.ServerInfoReceiveDuration != wantServerInfoReceiveDuration {
		t.Errorf("unexpected ServerInfoReceiveDuration, want %v actual %v",
			wantServerInfoReceiveDuration, o.ServerInfoReceiveDuration)
	}

}

func TestKeepAliveOptions(t *testing.T) {
	var (
		o                  = KeepaliveOptions{}
		wantKeepaliveTime  = 2 * time.Second
		wantKeepaliveIntvl = 3 * time.Second
	)
	ApplyKeepAlive([]SetKeepaliveOption{
		WithKeepaliveTime(wantKeepaliveTime),
		WithKeepaliveIntvl(wantKeepaliveIntvl),
	}, &o)

	if o.KeepaliveTime != wantKeepaliveTime {
		t.Errorf("unexpected KeepaliveTime, want %v actual %v", wantKeepaliveTime,
			o.KeepaliveTime)
	}

	if o.KeepaliveIntvl != wantKeepaliveIntvl {
		t.Errorf("unexpected KeepaliveIntvl, want %v actual %v", wantKeepaliveIntvl,
			o.KeepaliveIntvl)
	}
}
