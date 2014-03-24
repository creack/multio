package multio

import (
	"errors"
	"io"
	"sync"
)

var (
	// ErrClosedIO is caused by using the writer while closed.
	ErrClosedIO = errors.New("this I/O has been closed")
)

// MultiWriter is a io.Writer that writes to N underlying writers.
// New underlying writers can be add via MultiWriter.Add()
type MultiWriter struct {
	sync.RWMutex
	outputs []io.WriteCloser
	closed  bool
}

func (mw *MultiWriter) init() error {
	mw.RLock()
	if mw.closed {
		mw.RUnlock()
		return ErrClosedIO
	}
	init := false
	if mw.outputs == nil {
		init = true
	}
	mw.RUnlock()
	if init {
		mw.Lock()
		// Make sure nothing happend between the Runlock and the lock
		if mw.outputs == nil {
			mw.outputs = []io.WriteCloser{}
		}
		mw.Unlock()
	}
	return nil
}

// Write send buf to all underlying writers.
func (mw *MultiWriter) Write(buf []byte) (int, error) {
	if err := mw.init(); err != nil {
		return -1, err
	}

	mw.Lock()
	defer mw.Unlock()

	var (
		err    error
		bufLen = len(buf)
	)

	for _, out := range mw.outputs {
		// TODO: Evict Writer upon error
		n, e1 := out.Write(buf)
		if n != bufLen {
			err = io.ErrShortWrite
		}
		if e1 != nil {
			err = e1
		}

	}
	return len(buf), err
}

// Add adds a new writer in the MultiWriter.
// Any Write after this will send data to the Added writer.
func (mw *MultiWriter) Add(out io.WriteCloser) error {
	if err := mw.init(); err != nil {
		return err
	}

	mw.Lock()
	mw.outputs = append(mw.outputs, out)
	mw.Unlock()
	return nil
}

// Close terminate the writer and close all underlying writers.
func (mw *MultiWriter) Close() error {
	mw.Lock()
	defer mw.Unlock()

	var err error

	for _, out := range mw.outputs {
		if e1 := out.Close(); e1 != nil {
			err = e1
		}
	}
	mw.outputs = nil
	mw.closed = true
	return err
}
