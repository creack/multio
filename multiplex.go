package multio

import (
	"errors"
	"io"
	"log"
)

const (
	PageSize = 32 * 1024

	MPVersion = 1
)

var (
	ErrWrongReqSize      = errors.New("Error reading the request: wrong size")
	ErrUnkownRequestType = errors.New("Unkown request type or invalid request")
	ErrWrongType         = errors.New("Multiplexer need to have a Writer and a Reader as argument")
	ErrInvalidMessage    = errors.New("The message is invalid and can't be decoded")
	ErrInvalidVersion    = errors.New("The version from the message does not match the version of the multiplexe")
	ErrInvalidLength     = errors.New("The length from the message is not the length of the buffer")
)

type Multiplexer struct {
	r         io.Reader
	w         io.Writer
	c         io.Closer // TODO: implement Close()
	writeChan chan *Message
	readChans map[int]chan *Message
	ackChans  map[int]chan *Message
}

// decode cannot fail. In case of error, it populate the field err from Message.
func (m *Multiplexer) decodeMsg(src []byte, err error) (*Message, error) {
	if err != nil {
		return nil, err
	}
	msg := &Message{}
	msg.decode(src, nil)
	if msg.err != nil {
		return nil, msg.err
	}
	return msg, nil
}

func (m *Multiplexer) encodeMsg(src []byte) []byte {
	msg := &Message{
		data: src,
	}
	return msg.encode()
}

func (m *Multiplexer) StartRead() error {
	buf := make([]byte, PageSize+HeaderLen)
	for {
		n, err := m.r.Read(buf)
		msg, err := m.decodeMsg(buf[:n], err)
		if err != nil {
			// An error will cause deadlock panic if not properly handled
			log.Print(err)
			continue
		}
		switch msg.kind {
		case Frame:
			// Send the message. Use goroutine to queue the messages.
			// We do not use buffered chan because they have a fixed size.
			go func() {
				m.readChans[int(msg.id)] <- msg
			}()
		case Ack:
			m.ackChans[int(msg.id)] <- msg
		case Close:
			fallthrough
		default:
			panic("unimplemented")
		}
	}
	return nil
}

func (m *Multiplexer) StartWrite() error {
	for msg := range m.writeChan {
		encoded := msg.encode()
		m.w.Write(encoded)
	}
	return nil
}

func NewMultiplexer(rwc ...interface{}) (*Multiplexer, error) {
	m := &Multiplexer{}
	for _, rwc := range rwc {
		if r, ok := rwc.(io.Reader); ok && m.r == nil {
			m.r = r
		}
		if w, ok := rwc.(io.Writer); ok && m.w == nil {
			m.w = w
		}
		if c, ok := rwc.(io.Closer); ok && m.c == nil {
			m.c = c
		}
	}
	if m.r == nil || m.w == nil {
		return nil, ErrWrongType
	}
	m.writeChan = make(chan *Message)
	m.readChans = map[int]chan *Message{}
	m.ackChans = map[int]chan *Message{}
	go m.StartRead()
	go m.StartWrite()
	return m, nil
}

func (m *Multiplexer) NewWriter(id int) io.Writer {
	if _, exists := m.ackChans[id]; exists {
		return nil
	}

	m.ackChans[id] = make(chan *Message)

	return &Writer{
		id:        id,
		writeChan: m.writeChan,
		ackChan:   m.ackChans[id],
	}
}

func (m *Multiplexer) NewReader(id int) io.Reader {
	if _, exists := m.readChans[id]; exists {
		return nil
	}
	m.readChans[id] = make(chan *Message)
	return &Reader{
		id:        id,
		writeChan: m.writeChan,
		readChan:  m.readChans[id],
	}
}

type Writer struct {
	id        int
	writeChan chan *Message
	ackChan   chan *Message
	errChan   chan error
}

func (w *Writer) Write(buf []byte) (n int, err error) {
	w.writeChan <- NewMessage(Frame, w.id, buf)
	msg := <-w.ackChan
	return msg.n, msg.err
}

type Reader struct {
	id        int
	readChan  chan *Message
	writeChan chan *Message
}

func (r *Reader) Read(buf []byte) (int, error) {

	// Wait for a message
	msg := <-r.readChan
	copy(buf, msg.data)

	// Send ACK
	r.writeChan <- NewMessage(Ack, r.id, nil)
	return msg.n, msg.err
}
