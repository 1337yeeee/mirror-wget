package main

import "mirror-wget/internal/engine"

func main() {
	err := engine.Handle()
	if err != nil {
		panic(err)
	}
}
