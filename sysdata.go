package delegate

import (
	muss "github.com/mus-format/mus-stream-go"
	"github.com/mus-format/mus-stream-go/ord"
	"github.com/mus-format/mus-stream-go/raw"
	"github.com/mus-format/mus-stream-go/varint"
)

// ServerInfo helps the client identify a compatible server.
type ServerInfo []byte

// MarshalServerInfoMUS marshals a ServerInfo to the MUS format.
func MarshalServerInfoMUS(info ServerInfo, w muss.Writer) (n int, err error) {
	return ord.MarshalSlice[byte](info, nil,
		muss.MarshallerFn[byte](raw.MarshalByte),
		w)
}

// UnmarshalServerInfoMUS unmarshals a ServerInfo from the MUS format.
func UnmarshalServerInfoMUS(r muss.Reader) (info ServerInfo, n int, err error) {
	return ord.UnmarshalSlice[byte](nil,
		muss.UnmarshallerFn[byte](raw.UnmarshalByte),
		r)
}

// SizeServerInfoMUS returns the size of ServerInfo in the MUS format.
func SizeServerInfoMUS(info ServerInfo) (size int) {
	return ord.SizeSlice[byte](info, nil, muss.SizerFn[byte](raw.SizeByte))
}

// ServerSettings are sent by the server to the client.
//
// MaxCmdSize specifies the maximum allowed Сommand size. If it's <= 0, the size
// is unlimited.
type ServerSettings struct {
	MaxCmdSize int
}

// MarshalServerSettingsMUS marshals a ServerSettings to the MUS format.
func MarshalServerSettingsMUS(settings ServerSettings, w muss.Writer) (n int,
	err error) {
	return varint.MarshalInt(settings.MaxCmdSize, w)
}

// UnmarshalServerSettingsMUS unmarhsals a ServerSettings from the MUS format.
func UnmarshalServerSettingsMUS(r muss.Reader) (settings ServerSettings, n int,
	err error) {
	settings.MaxCmdSize, n, err = varint.UnmarshalInt(r)
	return
}

// SizeServerSettingsMUS returns the size of ServerSettings in the MUS
// format.
func SizeServerSettingsMUS(settings ServerSettings) (size int) {
	return varint.SizeInt(settings.MaxCmdSize)
}
