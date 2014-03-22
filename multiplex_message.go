package multio

import (
	"encoding/binary"
	"fmt"
)

// Message padding
const (
	HeaderLen = 16 // Size of the header
)
const (
	VersionIndex = iota << 2
	IDIndex      // Index where is the Fd
	SizeIndex    // Index where is the Frame size
	TypeIndex    // Index where is the type of the frame
)

func init() {
	fmt.Printf("init: %d, %d, %d, %d\n\n", VersionIndex, IDIndex, SizeIndex, TypeIndex)
}

// Message types
const (
	Frame = iota
	Ack
	Close
)

type Message struct {
	version uint32
	kind    uint32
	id      uint32
	size    uint32

	data []byte
	n    int
	err  error

	ack chan struct{}
}

// decode can't fail. If populate the err field in case of error
func (m *Message) decode(src []byte, err error) {
	if len(src) < HeaderLen {
		m.err = ErrInvalidMessage
		return
	}
	m.err = err
	println("version index:", VersionIndex)
	m.version = binary.BigEndian.Uint32(src[VersionIndex : VersionIndex+4])
	if m.version != MPVersion {
		m.err = ErrInvalidVersion
		return
	}
	println("version buf:", src[VersionIndex:VersionIndex+4])
	m.id = binary.BigEndian.Uint32(src[IDIndex : IDIndex+4])
	m.n = int(binary.BigEndian.Uint32(src[SizeIndex : SizeIndex+4]))
	if m.n != len(src)-HeaderLen {
		println("rr n:", m.n, src[SizeIndex], src[SizeIndex+1], src[SizeIndex+2], src[SizeIndex+3])
		m.err = ErrInvalidLength
		return
	}
	m.kind = binary.BigEndian.Uint32(src[TypeIndex : TypeIndex+4])
	fmt.Printf("-------> KIND (decode): %d, %v, %d\n", m.kind, src[TypeIndex:TypeIndex+4], TypeIndex)
	if m.n > 0 {
		m.data = make([]byte, m.n)
		copy(m.data, src[HeaderLen:])
	}
}

// encode cannot fail
func (m *Message) encode() []byte {
	buf := make([]byte, len(m.data)+HeaderLen)

	binary.BigEndian.PutUint32(buf[VersionIndex:VersionIndex+4], m.version)
	binary.BigEndian.PutUint32(buf[IDIndex:IDIndex+4], m.id)
	binary.BigEndian.PutUint32(buf[SizeIndex:SizeIndex+4], uint32(m.n))
	binary.BigEndian.PutUint32(buf[TypeIndex:TypeIndex+4], m.kind)

	fmt.Printf("-------> KIND (encode): %d, %v, %d\n", m.kind, buf[TypeIndex:TypeIndex+4], TypeIndex)
	copy(buf[HeaderLen:], m.data)
	fmt.Printf("-------> KIND (encode): %d, %v, %d\n", m.kind, buf[TypeIndex:TypeIndex+4], TypeIndex)
	return buf
}

func NewMessage(kind, id int, data []byte) *Message {
	println(len(data))
	return &Message{
		version: MPVersion,
		id:      uint32(id),
		kind:    uint32(kind),
		data:    data,
		n:       len(data),
	}
}
