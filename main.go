package main

import "github.com/flu-network/client/catalogue"

func main() {
	cat := &catalogue.Cat{}
	verify(cat.Init())
}

func verify(err error) {
	if err != nil {
		panic(err)
	}
}
