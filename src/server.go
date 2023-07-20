package main

import (
	"fmt"
	"log"
)

func ListenAndServe(host string, port uint16) {
	log.Printf("Listening on %s:%d\n", host, port)

	if err := app.Listen(fmt.Sprintf("%s:%d", host, port)); err != nil {
		panic(err)
	}
}
