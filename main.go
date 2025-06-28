package main

import (
	"fmt"

	"github.com/flimzy/anki"
)

const (
	version = "1.0.0"
)

func main() {
	fmt.Println("ANKI Ollama Translate v" + version)

	akpg, err := anki.ReadFile("test.apkg")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer akpg.Close()

	collection, err := akpg.Collection()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(collection)
}
