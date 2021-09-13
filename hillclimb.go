package main

import (
	"fmt"
	"github.com/emedvedev/enigma"
	"io/ioutil"
	"strconv"
	"strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const englishIOC = 0.06577359255736807

// IOC of english text, calculated using one time script from the corpus 'The Count of Monte Cristo'
// 0.06577359255736807

//IC on rotors and config -> IC on plugboard -> trigram on plugboard

func singleSwap(content string, swap string) float64 {
	content = strings.ReplaceAll(content, swap[0:1], swap[1:2])
	var final = calcIC(content)
	return final
}

func findSwaps(content string) {
	var scoreMap = make(map[string]float64)
	for _, letter := range alphabet {
		for _, letter2 := range alphabet {
			var score = singleSwap(content, string(letter)+string(letter2))
			scoreMap[string(letter)+string(letter2)] = score
		}

	}
	fmt.Println(scoreMap)
}

func calcIC(text string) float64 {

	text = strings.ToUpper(text)
	text = strings.ReplaceAll(text, " ", "")
	n := float64(len(text))
	var sum = 0.0
	for _, letter := range alphabet {
		var f = strings.Count(text, string(letter))
		sum += float64(f * (f - 1))
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
func enigmaSimulate() {
	config := make([]enigma.RotorConfig, 4)
	config[0] = enigma.RotorConfig{
		ID:    "",
		Start: 0,
		Ring:  1,
	}
	config[1] = enigma.RotorConfig{
		ID:    "",
		Start: 0,
		Ring:  1,
	}
	config[2] = enigma.RotorConfig{
		ID:    "IV",
		Start: "B"[0],
		Ring:  1,
	}
	config[3] = enigma.RotorConfig{
		ID:    "III",
		Start: "Q"[0],
		Ring:  16,
	}

	plugs := make([]string, 1)

	enigma.NewEnigma(config, "C-thin", plugs)

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
