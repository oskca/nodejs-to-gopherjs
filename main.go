package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
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
	// create file
	opath := strings.Replace(fpath, ".json", ".go", 1)
	w, err := os.OpenFile(opath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.WriteString(w, a.decl())
	if err != nil {
		log.Fatal(err)
	}
}
