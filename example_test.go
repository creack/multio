package multio_test

import (
	"fmt"
	"github.com/creack/multio"
	"io"
	"log"
)

func ExampleSimple() {
	r, w := io.Pipe()
	r2, w2 := io.Pipe()
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		m, err := multio.NewMultiplexer(r, w2)
		if err != nil {
			log.Fatal(err)
		}
		wr := m.NewWriter(0)
		fmt.Fprintf(wr, "Hello World!!!")

		wr2 := m.NewWriter(1)
		fmt.Fprintf(wr2, "Hello the World!!!")
	}()
	m, _ := multio.NewMultiplexer(r2, w)

	var (
		buf = make([]byte, multio.PageSize)
		n   int
		err error
	)

	rd := m.NewReader(0)
	n, err = rd.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])

	rd2 := m.NewReader(1)
	n, err = rd2.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
	<-ch

	// output:
	// Hello World!!!
	// Hello the World!!!
}
