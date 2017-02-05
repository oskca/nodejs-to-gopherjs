package main

import (
	"encoding/json"
	"log"
	"os"
	"fmt"
)

func main() {
	fpath := os.Args[1]
	log.Println(fpath)
	r, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	// parse
	var a ApiFile
	dec := json.NewDecoder(r)
	err = dec.Decode(&a)
	if err != nil {
		log.Fatal(err)
	}
	// spew.Dump(a)
	println("source:")
	fmt.Println(a.decl())
}
