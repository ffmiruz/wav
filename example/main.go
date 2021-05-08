package main

import (
	"fmt"
	"github.com/ffmiyo/wav"
	"io/ioutil"
	"os"
)

func main() {
	file, err := os.Open("test.wav")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	chunk, _ := wav.Unmarshal(file)

	buf := wav.Marshal(chunk)

	err = ioutil.WriteFile("copy.wav", buf, 0644)
	if err != nil {
		panic(err)
	}

}
