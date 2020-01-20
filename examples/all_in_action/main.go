package main

import "log"

func main() {
	if err := startServer(); err != nil {
		log.Fatal(err.Error())
	}
}
