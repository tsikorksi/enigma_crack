package main

import (
	"fmt"
	"github.com/emedvedev/enigma"
	"io/ioutil"
	"math"
	"sort"
	"strconv"
	"strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var trisMap map[string]int64

// Generates Trigram score lookup table
// Trigrams sourced from:
// http://practicalcryptography.com/cryptanalysis/letter-frequencies-various-languages/english-letter-frequencies/
func processGrams(tris string) map[string]int64 {
	trisMap = make(map[string]int64)
	trisArray := strings.Split(tris, "\n")
	for _, line := range trisArray {
		lineData := strings.Split(line, " ")
		trisMap[lineData[0]], _ = strconv.ParseInt(strings.Trim(lineData[1], "\r"), 10, 32)
	}
	return trisMap
}

//const englishIOC = 0.06577359255736807

// IOC of english text, calculated using one time script from the corpus 'The Count of Monte Cristo'
// 0.06577359255736807
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

// Generates Trigram score for content from pre-generated trigram map
func gramScore(content string) int64 {
	var score int64 = 0
	for key, value := range trisMap {
		var count = int64(strings.Count(content, key))
		score += count * value
	}
	return score
}

// Swaps two letters in ciphertext and checks the IoC for the new version
func singleSwapIOC(content string, config []enigma.RotorConfig, swapList []string) float64 {
	return math.Round(calcIC(enigmaSimulate(config[0], config[1], content, swapList))*10000) / 10000
}

// The lack of generics is literally killing me
func singleSwapGram(content string, config []enigma.RotorConfig, swapList []string) int64 {
	return gramScore(enigmaSimulate(config[0], config[1], content, swapList))
}

// Small wrapper on standard swap function
func conductSwapIOC(swapList []string, content string, config []enigma.RotorConfig, joined string, i int) []string {
	var initial = singleSwapIOC(content, config, swapList)
	var temp = swapList[i]
	swapList[i] = joined
	var score = singleSwapIOC(content, config, swapList)
	if score < initial {
		swapList[i] = temp
	}
	return swapList
}

//// Small wrapper on standard swap function, but with grams
//func conductSwapGram(swapList []string, content string, config []enigma.RotorConfig, joined string, i int) []string {
//	var initial = singleSwapGram(content, config, swapList)
//	var temp = swapList[i]
//	swapList[i] = joined
//	var score = singleSwapGram(content, config, swapList)
//	if score < initial {
//		swapList[i] = temp
//	}
//	return swapList
//}

// Finds all legal swaps and their IoC after such a swap is made
func hillclimbIOC(content string, config []enigma.RotorConfig, swapList []string) ([]string, map[string]float64) {
	var scoreMap = make(map[string]float64)
	for _, letter := range alphabet {
		for _, letter2 := range alphabet {
			var swapped = false
			// Because AB = BA ...
			s := strings.Split(string(letter)+string(letter2), "")
			sort.Strings(s)
			var joined = strings.Join(s, "")
			_, ok := scoreMap[joined]
			// Because AA etc. is not possible...
			if letter != letter2 && !ok {
				scoreMap[joined] = 0.0
				if len(swapList) == 0 {
					swapList = append(swapList, joined)
					swapped = true
				} else if !swapped && strings.ContainsAny(strings.Join(swapList, ""), string(letter)) && strings.ContainsAny(strings.Join(swapList, ""), string(letter2)) {
					var initial = singleSwapIOC(content, config, swapList)
					var tempSwaps []string
					for i, v := range swapList {
						if strings.ContainsAny(v, joined) {
							tempSwaps = append(tempSwaps, v)
							swapList[i] = ""
						}
					}
					sort.Strings(swapList)
					swapList = swapList[2:]
					swapList = append(swapList, joined)
					var score = singleSwapIOC(content, config, swapList)

					if initial > score {
						swapList = swapList[:len(swapList)-1]
						swapList = append(swapList, tempSwaps...)
					}
					sort.Strings(swapList)
				} else {
					for i, v := range swapList {
						if string(v[0]) == string(joined[0]) && !strings.ContainsAny(strings.Join(swapList, ""), string(letter2)) {
							conductSwapIOC(swapList, content, config, joined, i)
							swapped = true
							break
						} else if string(v[1]) == string(joined[1]) && !swapped && !strings.ContainsAny(strings.Join(swapList, ""), string(letter)) {
							conductSwapIOC(swapList, content, config, joined, i)
							swapped = true
						} else if !strings.ContainsAny(strings.Join(swapList, ""), joined) {
							swapList = append(swapList, joined)
							swapped = true
						}
						swapped = false

					}
				}

			}
		}
	}

	return swapList[:10], scoreMap
}

type Pair struct {
	Key   string
	Value int64
}

// PairList All the Pairlist nonsense stolen from:
// https://groups.google.com/g/golang-nuts/c/FT7cjmcL7gw
type PairList []Pair

// Since golang maps are bad, and I couldn't work out another way to sort a map by value

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

func hillclimbTrigram(swaps map[string]float64, config []enigma.RotorConfig, content string, plugboard []string) []string {
	var finalSwaps []string
	for key := range swaps {
		if !strings.ContainsAny(strings.Join(plugboard, ""), key) {
			finalSwaps = append(finalSwaps, key)
		}
	}
	//var plugboard = finalSwaps[0:10]
	//var score = gramScore(trisMap, enigmaSimulate(config[0], config[1], content, []string{}))
	var plugboardMap = make(map[string]int64)

	for _, combo := range finalSwaps {
		var tempScore = singleSwapGram(content, config, append(plugboard, combo))
		plugboardMap[combo] = tempScore

	}

	var pairList PairList

	for key, value := range plugboardMap {
		pairList = append(pairList, Pair{key, value})
	}
	sort.Sort(sort.Reverse(pairList))
	var count = 7
	// remove duplicate letter swaps, selecting best
	for len(plugboard) < 10 {
		if !strings.ContainsAny(strings.Join(plugboard[7:], ""), pairList[count].Key) {
			plugboard = append(plugboard, pairList[count].Key)
		}
		count++
	}

	return plugboard
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
	var rotorsFirst = [2]string{"Beta", "Gamma"}
	var rotorsSecond = [4]string{"I", "II", "V", "VI"}
	var max = calcIC(content)
	var rotorA enigma.RotorConfig
	var rotorB enigma.RotorConfig
	var rotorAFinal enigma.RotorConfig
	var rotorBFinal enigma.RotorConfig
	for _, rotor := range rotorsFirst {
		for _, letter := range alphabet {
			rotorA = enigma.RotorConfig{
				ID:    rotor,
				Start: string(letter)[0],
				Ring:  1,
			}
			for _, rotor2 := range rotorsSecond {
				for _, letter2 := range alphabet {

					rotorB = enigma.RotorConfig{
						ID:    rotor2,
						Start: string(letter2)[0],
						Ring:  1,
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
	trisMap = processGrams(tris)
	var config = rotorBrute(content)
	var swapList = make([]string, 0, 48)
	//var scoreMap map[string]float64
	plugboard, _ := hillclimbIOC(content, config, swapList)
	//var plugboard = make([]string, 0, 10)
	//plugboard = hillclimbTrigram(scoreMap ,config, content, swapList)
	sort.Strings(plugboard)

	var id []string
	for _, v := range config {
		id = append(id, v.ID)
	}
	fmt.Println(strings.Join(id, " "))
	var starts []string
	for _, v := range config {
		starts = append(starts, string(v.Start))
	}
	fmt.Println(strings.Join(starts, " "))

	fmt.Println(strings.Join(plugboard, " "))
	//fmt.Print(enigmaSimulate(config[0], config[1], content, plugboard))
}
