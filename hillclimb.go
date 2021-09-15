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
func singleSwap(content string, swap string, config []enigma.RotorConfig) float64 {
	var swapList = []string{swap}
	return calcIC(enigmaSimulate(config[0], config[1], content, swapList))
}

// Finds all legal swaps and their IoC after such a swap is made
func findSwaps(content string, config []enigma.RotorConfig) map[string]float64 {
	var scoreMap = make(map[string]float64)
	for _, letter := range alphabet {
		for _, letter2 := range alphabet {
			// Because AB = BA ...
			s := strings.Split(string(letter)+string(letter2), "")
			sort.Strings(s)
			var joined = strings.Join(s, "")
			_, k := scoreMap[joined]
			// Because AA etc. is not possible...
			if letter != letter2 || k {
				var score = singleSwap(content, joined, config)
				scoreMap[joined] = score
			}
		}

	}

	return scoreMap
}

// removes worse performing half of possible swaps
func extractBetter(swaps map[string]float64) map[string]float64 {
	var total float64 = 0
	var vals []float64
	for _, val := range swaps {
		total += val
		vals = append(vals, val)
	}
	var avg = total / float64(len(swaps))
	for key, value := range swaps {
		if value < avg {
			delete(swaps, key)
		}
	}
	return swaps
}

// Calculates IoC
func calcIC(text string) float64 {
	n := float64(len(text))
	var sum = 0.0
	for _, letter := range alphabet {
		var f = strings.Count(text, string(letter))
		sum += float64(f * (f - 1))
	}
	return sum / (n * (n - 1))
}

// Generates Trigram score lookup table
// Trigrams sourced from:
// https://github.com/torognes/enigma/blob/master/english_trigrams.txt
func processTris(tris string) map[string]int64 {
	var trisMap = make(map[string]int64)
	trisArray := strings.Split(tris, "\n")
	for _, line := range trisArray {
		lineData := strings.Split(line, " ")
		trisMap[lineData[0]], _ = strconv.ParseInt(strings.Trim(lineData[1], "\r"), 10, 32)
	}
	return trisMap
}

// Generates Trigram score for content from pre-generated trigram map
func trigramScore(trisMap map[string]int64, content string) int64 {
	var score int64 = 0
	for key, value := range trisMap {
		var count = int64(strings.Count(content, key))
		score += count * value
	}
	return score
}

// Read file into string
func readFile(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File read error")
	}
	return string(content)
}

// Brute force a most likely rotor config
func rotorBrute(content string) []enigma.RotorConfig {
	plugs := make([]string, 0)
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
				var local = calcIC(enigmaSimulate(rotorA, rotorB, content, plugs))
				if local > max {
					max = local
					rotorAFinal = rotorA
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
				var local = calcIC(enigmaSimulate(rotorA, rotorB, content, plugs))
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

func enigmaSimulate(rotorA enigma.RotorConfig, rotorB enigma.RotorConfig, content string, plugs []string) string {
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

	// plugs := make([]string, 0)

	var trial = enigma.NewEnigma(config, "C-thin", plugs)
	return trial.EncodeString(content)
}

func main() {

	content := readFile("ct.txt")
	content = enigma.SanitizePlaintext(content)
	tris := readFile("english_tri.txt")
	var score = trigramScore(processTris(tris), content)
	fmt.Println(content)
	fmt.Println(calcIC(content))
	fmt.Println(score)
	var config = rotorBrute(content)
	var swaps = extractBetter(findSwaps(content, config))

	fmt.Println(len(swaps))
}
