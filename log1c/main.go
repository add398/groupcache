/**
 * @Author: Log1c
 * @Description:
 * @File:  main
 * @Version: 1.0.0
 * @Date: 2022/2/23 21:55
 */

package main

import (
	"context"
	"errors"
	"flag"
	"github.com/golang/groupcache"
	"log"
	"net/http"
	"strings"
)

// Simple groupcache example: https://github.com/golang/groupcache
// Running 3 instances:
// go run groupcache.go -addr=:8080 -pool=http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082
// go run groupcache.go -addr=:8081 -pool=http://127.0.0.1:8081,http://127.0.0.1:8080,http://127.0.0.1:8082
// go run groupcache.go -addr=:8082 -pool=http://127.0.0.1:8082,http://127.0.0.1:8080,http://127.0.0.1:8081
// Testing:
// curl localhost:8080/color?name=red


var Store = map[string][]byte{
	"red":   []byte("#FF0000"),
	"green": []byte("#00FF00"),
	"blue":  []byte("#0000FF"),
}

var Group = groupcache.NewGroup("foobar", 64<<20, groupcache.GetterFunc(
	func(ctx context.Context, key string, dest groupcache.Sink) error {
		log.Println("looking up", key)
		v, ok := Store[key]
		if !ok {
			return errors.New("color not found")
		}
		dest.SetBytes(v)
		return nil
	},
))

func main() {
	addr := flag.String("addr", ":8080", "server address")
	peers := flag.String("pool", "http://localhost:8080", "server pool list")
	flag.Parse()
	http.HandleFunc("/color", func(w http.ResponseWriter, r *http.Request) {
		color := r.FormValue("name")
		var b []byte
		err := Group.Get(nil, color, groupcache.AllocatingByteSliceSink(&b))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Write(b)
		w.Write([]byte{'\n'})
	})
	p := strings.Split(*peers, ",")
	pool := groupcache.NewHTTPPool(p[0])
	pool.Set(p...)
	http.ListenAndServe(*addr, nil)
}
