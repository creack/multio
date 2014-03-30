Multio
======

[![GoDoc](https://godoc.org/github.com/creack/multio?status.png)](https://godoc.org/github.com/creack/multio)
[![Build Status](https://travis-ci.org/creack/multio.png)](https://travis-ci.org/creack/multio)
[![Coverage Status](https://coveralls.io/repos/creack/multio/badge.png?branch=master)](https://coveralls.io/r/creack/multio?branch=master)

[![status](https://sourcegraph.com/api/repos/github.com/creack/multio/badges/status.png)](https://sourcegraph.com/github.com/creack/multio)
[![docs examples](https://sourcegraph.com/api/repos/github.com/creack/multio/badges/docs-examples.png)](https://sourcegraph.com/github.com/creack/multio)
[![xrefs](https://sourcegraph.com/api/repos/github.com/creack/multio/badges/xrefs.png)](https://sourcegraph.com/github.com/creack/multio)
[![funcs](https://sourcegraph.com/api/repos/github.com/creack/multio/badges/funcs.png)](https://sourcegraph.com/github.com/creack/multio)
[![top func](https://sourcegraph.com/api/repos/github.com/creack/multio/badges/top-func.png)](https://sourcegraph.com/github.com/creack/multio)
[![library users](https://sourcegraph.com/api/repos/github.com/creack/multio/badges/library-users.png)](https://sourcegraph.com/github.com/creack/multio)

Multiple I/O operation for Golang
=================================

Multio is a simple multiplexing / MultiRead / MultiWriter libarary.

It allows you to use N streams on to of a read/write pair. It can be pipes, socket, or anything that can read and write.

Note that you do need read AND write on both side, you can't have read on one side and write on the other has the protocol (to be documented) sends back acknoledgment messages.
