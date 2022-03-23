package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var Ballast = "1AASSSaaa"

func main() {
	fmt.Printf("Demo spyware (%s)\n", Ballast)
	url := "http://wrs21.winshipway.com/"
	fmt.Printf("Get: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Got\n%s\n", html)
}
