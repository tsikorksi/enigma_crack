package main

import (
	"fmt"
	"github.com/emedvedev/enigma"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

//const englishIOC = 0.06577359255736807

// IOC of english text, calculated using one time script from the corpus 'The Count of Monte Cristo'
// 0.06577359255736807

//IC on rotors and config -> IC on plugboard -> trigram on plugboard

// Swaps two letters in ciphertext and checks the IoC for the new version
func singleSwap(content string, swap string) float64 {
	var replacer = strings.NewReplacer(swap[0:1], swap[1:2], swap[1:2], swap[0:1])
	content = replacer.Replace(content)
	var final = calcIC(content)
	return final
}

// Finds all legal swaps and their
func findSwaps(content string) map[string]float64 {
	var scoreMap = make(map[string]float64)
	for _, letter := range alphabet {
		for _, letter2 := range alphabet {
			s := strings.Split(string(letter)+string(letter2), "")
			sort.Strings(s)
			var joined = strings.Join(s, "")
			_, k := scoreMap[joined]
			if letter != letter2 || k {
				var score = singleSwap(content, joined)
				scoreMap[joined] = score
			}
		}

	}
	return scoreMap
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

// Brute force a most likely rotor config
func rotorBrute(content string) []enigma.RotorConfig {
	var rotors = [8]string{"I", "II", "V", "VI", "VII", "VIII", "Beta", "Gamma"}
	var max = calcIC(content)
	var rotorA enigma.RotorConfig
	var rotorB = enigma.RotorConfig{
		ID:    "IV",
		Start: "B"[0],
		Ring:  1,
	}
	var rotorAFinal enigma.RotorConfig
	var rotorBFinal enigma.RotorConfig
	for _, rotor := range rotors {
		for _, letter := range alphabet {
			for i := 0; i < 27; i++ {
				rotorA = enigma.RotorConfig{
					ID:    rotor,
					Start: string(letter)[0],
					Ring:  i,
				}
				var local = calcIC(enigmaSimulate(rotorA, rotorB, content))
				if local > max {
					max = local
					rotorAFinal = rotorA
					//rotorBFinal = rotorB
				}
			}
		}
	}
	rotorA = rotorAFinal
	for _, rotor2 := range rotors {
		for _, letter2 := range alphabet {
			for j := 0; j < 27; j++ {

				rotorB = enigma.RotorConfig{
					ID:    rotor2,
					Start: string(letter2)[0],
					Ring:  j,
				}
				var local = calcIC(enigmaSimulate(rotorA, rotorB, content))
				if local > max {
					max = local
					rotorAFinal = rotorA
					rotorBFinal = rotorB
				}
			}
		}

	}
	config := make([]enigma.RotorConfig, 4)
	config[0] = rotorAFinal
	config[1] = rotorBFinal
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
	return config
}

func enigmaSimulate(rotorA enigma.RotorConfig, rotorB enigma.RotorConfig, content string) string {
	config := make([]enigma.RotorConfig, 4)
	config[0] = rotorA
	config[1] = rotorB
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

	plugs := make([]string, 0)

	var trial = enigma.NewEnigma(config, "C-thin", plugs)
	return trial.EncodeString(content)
}

func main() {

	content := readFile("ct.txt")
	tris := readFile("english_tri.txt")
	var score = trigramScore(processTris(tris), content)
	fmt.Println(content)
	fmt.Println(calcIC(content))
	fmt.Println(score)
	fmt.Println(findSwaps(content))
	fmt.Println(rotorBrute(content))
}
