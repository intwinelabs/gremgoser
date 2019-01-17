# gremgoser

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/intwinelabs/gremgoser)

gremgoser is a fast, efficient, and easy-to-use client for the TinkerPop graph database stack. It is a Gremlin language driver which uses WebSockets to interface with Gremlin Server and has a strong emphasis on concurrency and scalability. Gremgoser started as a fork of [gremgo](http://github.com/qasaur/gremgo). The main difference is gremgoser supports serializing and de-serializing interfaces in/out of a graph as well as Vertex and edge creation from Go interfaces. Please keep in mind that gremgoser is still under heavy development and is currently not production ready.

Installation
==========
```
go get github.com/intwinelabs/gremgoser
```

Documentation
==========

* [GoDoc](https://godoc.org/github.com/intwinelabs/gremgoser)

Example
==========
This is a example of a secure connection with authentication.  The example also shows the ability to pass a interface to create a Vertex in the graph as well as the ability to create Edges between interfaces. This also shows de-serialization of interface from the graph.

```go
package main

import (
	"fmt"
	"log"

	"github.com/intwinelabs/gremgoser"
)

func main() {
	errs := make(chan error)
	go func(chan error) {
		err := <-errs
		log.Fatal("Lost connection to the database: " + err.Error())
	}(errs) // Example of connection error handling logic

	host := "wss://gremlin.server:443/"
	user := "username"
	pass := "password"
	conf := gremgoser.SetAuthentication(user, pass)
	dialer := gremgoser.NewDialer(host, conf) // Returns a WebSocket dialer to connect to Gremlin Server
	g, err := gremgoser.Dial(dialer, errs)    // Returns a gremgoser client to interact with
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := g.Execute( // Sends a query to Gremlin Server with bindings
		"g.V()",
		map[string]string{},
		map[string]string{},
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(res)

	// Add a person
	p := Person{
		Id:   uuid.New(),
		PID:  "testID",
		Name: "Ted",
		Age:  48,
	}

	res2, err := g.AddV("person", p)

	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(res2)

	// Add a person
	p2 := Person{
		Id:   uuid.New(),
		PID:  "testID",
		Name: "Donald",
		Age:  72,
	}

	res3, err := g.AddV("person", p2)

	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(res3)

	// Ted Likes Donald
	res4, err := g.AddE("likes", p, p2)
	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(res4)

	// Lets ask the graph for p back as a new instance of a Person
	id := p.Id.String()
	p3 := []Person{}
	q := fmt.Sprintf("g.V('%s')", id)
	//q = "g.V()"
	fmt.Println(q)
	err = g.Get(q, &p3)
	if err != nil {
		fmt.Println(err)
		return
	}
	spew.Dump(p3)
}
```

License
==========

Copyright (c) 2018 Intwine Labs, Inc.

Copyright (c) 2016 Marcus Engvall

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
