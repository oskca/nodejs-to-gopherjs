package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
)

var enableComment bool

type decler interface {
	decl() string
	comment() string
}

func declSlice(ar interface{}) string {
	v := reflect.ValueOf(ar)
	if v.IsNil() {
		return ""
	}
	var t []decler
	ret := reflect.MakeSlice(reflect.TypeOf(t), 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		ret = reflect.Append(ret, v.Index(i))
	}
	// declers to string
	ds := ret.Interface().([]decler)
	ss := []string{}
	for i := 0; i < len(ds); i++ {
		comment := ds[i].comment()
		if enableComment && comment != "" {
			ss = append(ss, comment)
		}
		ss = append(ss, ds[i].decl())
	}
	return strings.Join(ss, "\n\t")
}

func main() {
	fpath := flag.Arg(0)
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
	w, err := os.OpenFile(opath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.WriteString(w, a.decl())
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	flag.BoolVar(&enableComment, "c", false, "generate comment")
	flag.Parse()
}
