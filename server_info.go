package delegate

import (
	muss "github.com/mus-format/mus-stream-go"
	"github.com/mus-format/mus-stream-go/ord"
	"github.com/mus-format/mus-stream-go/raw"
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
