package multio_test

import (
	"fmt"
	"github.com/creack/multio"
	"io"
	"log"
	"os"
)

func ExampleMultiplexer_simplePipe() {
	// Create two pipe pair to have both way read/write
	// We could use socketpair, socket or other, but let's
	// stick with pipes for now
	r, w := io.Pipe()
	r2, w2 := io.Pipe()

	// Channel to wait on the goroutine before exit
	ch := make(chan struct{})
	go func() {
		// When the goroutine finished, unlock (close chan)
		defer close(ch)

		// Create a side of the multiplexer
		m, err := multio.NewMultiplexer(r, w2)
		if err != nil {
			log.Fatal(err)
		}

		// Create two writers
		wr := m.NewWriter(0)
		wr2 := m.NewWriter(1)

		// Send data and close.
		// In this example, we use io.Copy which blocks until close
		// If we do not close before write to wr2, the other goroutine
		// will never read so it will block forever.
		fmt.Fprintf(wr, "Hello World!!!\n")
		wr.Close()
		fmt.Fprintf(wr2, "Hello the World!!!\n")
		wr2.Close()
	}()

	// Create the other side of the multiplexer
	m, err := multio.NewMultiplexer(r2, w)
	if err != nil {
		log.Fatal(err)
	}

	// Create two readers corresponding to our writers
	rd := m.NewReader(0)
	rd2 := m.NewReader(1)

	// Copy everything (until close) to stdout
	io.Copy(os.Stdout, rd)
	io.Copy(os.Stdout, rd2)

	// Wait on the goroutine to finish
	<-ch

	// output:
	// Hello World!!!
	// Hello the World!!!
}

func ExampleMultiplexer_readWriter() {
	// Create two pipe pair to have both way read/write
	// We could use socketpair, socket or other, but let's
	// stick with pipes for now
	r, w := io.Pipe()
	r2, w2 := io.Pipe()

	// Channel to wait on the goroutine before exit
	ch := make(chan struct{})
	go func() {
		// When the goroutine finished, unlock (close chan)
		defer close(ch)

		// Create a side of the multiplexer
		m, err := multio.NewMultiplexer(r, w2)
		if err != nil {
			log.Fatal(err)
		}

		// Create a reader and a writer
		rwc := m.NewReadWriter(0)

		// Send and Read data then close.
		fmt.Fprintf(rwc, "Hello World!!!\n")
		io.Copy(os.Stdout, rwc)
		rwc.Close()
	}()

	// Create the other side of the multiplexer
	m, err := multio.NewMultiplexer(r2, w)
	if err != nil {
		log.Fatal(err)
	}

	// Create a reader and a writer
	rwc := m.NewReadWriter(0)

	buf := make([]byte, multio.PageSize)
	n, err := rwc.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	_, err = rwc.Write(buf[:n])
	if err != nil {
		log.Fatal(err)
	}
	rwc.Close()

	// Wait on the goroutine to finish
	<-ch

	// output:
	// Hello World!!!
}
