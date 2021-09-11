package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// IOC of english text, calculated using one time script from the corpus 'The Count of Monte Cristo'
// 0.06577359255736807

func singleSwap(content string, swap string) float64 {
	var initial = calcIC(content)
	content = strings.ReplaceAll(content, swap[0:1], swap[1:2])
	var final = calcIC(content)
	return final - initial
}

func findSwaps(content string) {
	//var stecker = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	fmt.Println(singleSwap(content, "AB"))
	fmt.Println(singleSwap(content, "AC"))
	fmt.Println(singleSwap(content, "AD"))
	fmt.Println(singleSwap(content, "AE"))
}

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

func processTris(tris string) map[string]int64 {
	var trisMap = make(map[string]int64)
	trisArray := strings.Split(tris, "\n")
	for _, line := range trisArray {
		lineData := strings.Split(line, " ")
		trisMap[lineData[0]], _ = strconv.ParseInt(strings.Trim(lineData[1], "\r"), 10, 32)
	}
	return trisMap
}

func trigramScore(trisMap map[string]int64, content string) int64 {
	var score int64 = 0
	for key, value := range trisMap {
		var count = int64(strings.Count(content, key))
		score += count * value
	}
	return score
}

func readFile(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File read error")
	}
	return string(content)
}

func main() {

	content := readFile("test.txt")
	tris := readFile("english_tri.txt")
	var score = trigramScore(processTris(tris), content)
	fmt.Println(content)
	fmt.Println(calcIC(content))
	fmt.Println(score)
	findSwaps(content)
}
