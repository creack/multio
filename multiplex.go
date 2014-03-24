package multio

import (
	"errors"
	"github.com/creack/multio/logger"
	"io"
	"sync"
)

const (
	PageSize = 32 * 1024

	MPVersion = 1
)

var log = logger.New(nil, "multiplex", 2)

var (
	ErrWrongReqSize      = errors.New("Error reading the request: wrong size")
	ErrUnkownRequestType = errors.New("Unkown request type or invalid request")
	ErrWrongType         = errors.New("Multiplexer need to have a Writer and a Reader as argument")
	ErrInvalidMessage    = errors.New("The message is invalid and can't be decoded")
	ErrInvalidVersion    = errors.New("The version from the message does not match the version of the multiplexe")
	ErrInvalidLength     = errors.New("The length from the message is not the length of the buffer")
)

type chanMap struct {
	sync.RWMutex
	msgs map[int]chan *Message
}

func (cm *chanMap) Get(key int) chan *Message {
	cm.RLock()
	defer cm.RUnlock()

	if cm.msgs == nil {
		cm.msgs = make(map[int]chan *Message)
	}
	if _, exists := cm.msgs[key]; !exists {
		cm.msgs[key] = make(chan *Message)
	}
	return cm.msgs[key]
}

func (cm *chanMap) Delete(key int) {
	cm.Lock()
	defer cm.Unlock()

	if _, exists := cm.msgs[key]; exists {
		delete(cm.msgs, key)
	}
}

func (cm *chanMap) SetChanIfNotExist(key int) {
	cm.Lock()
	defer cm.Unlock()

	if cm.msgs == nil {
		cm.msgs = make(map[int]chan *Message)
	}
	if _, exists := cm.msgs[key]; exists {
		return
	} else {
		cm.msgs[key] = make(chan *Message)
	}
}

type Multiplexer struct {
	r         io.Reader
	w         io.Writer
	c         []io.Closer // TODO: implement Close()
	writeChan chan *Message
	readChans chanMap
	ackChans  chanMap
	closed    bool
}

// decode cannot fail. In case of error, it populate the err field from Message.
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

func (m *Multiplexer) closeChan(id int) {
	if c := m.ackChans.Get(id); c != nil {
		close(c)
		m.ackChans.Delete(id)
	}
	if c := m.readChans.Get(id); c != nil {
		close(c)
		m.readChans.Delete(id)
	}
}

func (m *Multiplexer) StartRead() error {
	defer log.Debug("StartRead finished")
	buf := make([]byte, PageSize+HeaderLen)
	for !m.closed {
		n, err := m.r.Read(buf)
		log.Lprintf(5, "read: %v, closed: %v\n", err, m.closed)
		msg, err := m.decodeMsg(buf[:n], err)
		if err != nil {
			if err == io.EOF {
				m.Close()
				continue
			}
			// An error will cause deadlock panic if not properly handled
			log.Error(err)
			continue
		}
		switch msg.kind {
		case Frame:
			// Send the message. Use goroutine to queue the messages.
			// We do not use buffered chan because they have a fixed size.
			go func() { m.readChans.Get(int(msg.id)) <- msg }()
		case Ack:
			m.ackChans.Get(int(msg.id)) <- msg
		case Close:
			m.closeChan(int(msg.id))
		default:
			panic("unimplemented")
		}
	}
	return nil
}

func (m *Multiplexer) StartWrite() error {
	defer log.Debug("---> StartWrite finished")
	for msg := range m.writeChan {
		encoded := msg.encode()
		m.w.Write(encoded)
		if msg.kind == Close {
			m.closeChan(int(msg.id))
		}
	}
	return nil
}

func NewMultiplexer(rwc ...interface{}) (*Multiplexer, error) {
	m := &Multiplexer{
		c: []io.Closer{},
	}
	for _, rwc := range rwc {
		if r, ok := rwc.(io.Reader); ok && m.r == nil {
			m.r = r
		}
		if w, ok := rwc.(io.Writer); ok && m.w == nil {
			m.w = w
		}
		if c, ok := rwc.(io.Closer); ok {
			log.Debug("Adding a closer")
			m.c = append(m.c, c)
		} else {
			log.Debug("Adding a non-closer")
		}
	}
	if m.r == nil || m.w == nil {
		return nil, ErrWrongType
	}
	m.writeChan = make(chan *Message)
	m.readChans = chanMap{}
	m.ackChans = chanMap{}

	go m.StartRead()
	go m.StartWrite()
	return m, nil
}

func (m *Multiplexer) NewWriter(id int) *WriteCloser {
	m.ackChans.SetChanIfNotExist(id)

	return &WriteCloser{
		id:        id,
		writeChan: m.writeChan,
		ackChan:   m.ackChans.Get(id),
	}
}

func (m *Multiplexer) NewReader(id int) *ReadCloser {
	m.readChans.SetChanIfNotExist(id)

	return &ReadCloser{
		id:        id,
		writeChan: m.writeChan,
		readChan:  m.readChans.Get(id),
	}
}

func (m *Multiplexer) NewReadWriter(id int) *ReadWriteCloser {
	return &ReadWriteCloser{
		ReadCloser:  m.NewReader(id),
		WriteCloser: m.NewWriter(id),
	}
}

func (m *Multiplexer) Close() error {
	m.closed = true
	for id := range m.readChans.msgs {
		m.closeChan(id)
	}
	for id := range m.ackChans.msgs {
		m.closeChan(id)
	}
	log.Debugf("Closing writeChan\n")
	if m.writeChan != nil {
		close(m.writeChan)
		m.writeChan = nil
	}
	log.Debugf("Closing %d i/o\n", len(m.c))
	for _, c := range m.c {
		c.Close()
	}
	return nil
}

type WriteCloser struct {
	id        int
	writeChan chan *Message
	ackChan   chan *Message
}

func (w *WriteCloser) Write(buf []byte) (n int, err error) {
	// Send the buffer to the other side
	w.writeChan <- NewMessage(Frame, w.id, buf)
	// Wait for ACK
	msg := <-w.ackChan
	if msg == nil {
		return 0, io.EOF
	}
	return msg.n, msg.err
}

func (w *WriteCloser) Close() error {
	w.writeChan <- NewMessage(Close, w.id, nil)
	return nil
}

type ReadCloser struct {
	id        int
	readChan  chan *Message
	writeChan chan *Message
}

func (r *ReadCloser) Read(buf []byte) (int, error) {
	// Wait for a message
	msg := <-r.readChan
	if msg == nil {
		return -1, io.EOF
	}
	copy(buf, msg.data)

	// Send ACK
	r.writeChan <- NewMessage(Ack, r.id, nil)
	return msg.n, msg.err
}

func (r *ReadCloser) Close() error {
	r.writeChan <- NewMessage(Close, r.id, nil)
	return nil
}

type ReadWriteCloser struct {
	*ReadCloser
	*WriteCloser
}

func (rw *ReadWriteCloser) Close() error {
	e1 := rw.ReadCloser.Close()
	e2 := rw.WriteCloser.Close()
	if e1 != nil {
		return e1
	}
	return e2
}
