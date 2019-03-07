package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	file, _ := ioutil.ReadFile("pi.json")

	var data map[string]interface{}

	json.Unmarshal(file, &data)
	fmt.Println(data)
}
