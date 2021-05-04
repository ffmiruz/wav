package main

import (
	"fmt"
	"github.com/ffmiyo/wav"
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
	fmt.Println(chunk.Data[0][0:100])

}
