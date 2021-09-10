package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func calcIC(text string) float64 {
	text = strings.ToUpper(text)
	text = strings.ReplaceAll(text, " ", "")
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	n := float64(len(text))
	var sum = 0.0
	for _, letter := range alphabet {
		var f = strings.Count(text, string(letter))
		sum += float64(f * (f - 1))
		//fmt.Printf("%s: %d\n", string(letter), f)
	}
	return sum / (n * (n - 1))
}

func readFile() {
	content, err := ioutil.ReadFile("test.txt")
	if err != nil {
		fmt.Println("test")
	}
	fmt.Println(string(content))
	fmt.Println(calcIC(string(content)))
}

func main() {
	readFile()
}
