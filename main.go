package main

import "fmt"
import "strings"

// create a text corpus divided by sentenses
// create bigram and next word map
// while creating next word calculate the P(w1/ prev 2 words)

// this will be the base model

// v2- for any new bigrams for every user search update the mapped next world prrob

func getTestCorpus() []string {
	corpus := make([]string, 0)
	corpus = append(corpus, "this is a the house that jack built")
	corpus = append(corpus, "this is the malt")
	corpus = append(corpus, "this is a rat")
	corpus = append(corpus, "this is a cat")
	corpus = append(corpus, "that killed the rat")
	corpus = append(corpus, "that ate the malt")
	return corpus
}

type bigramNextValue struct {
	word string
	prob float32
}

func calculateWordProbability(corpus []string, nGramString string, word string) float32 {
	if word == "#" {
		return 0 // end of the sentence; this should be chosen at the end
	}

	fmt.Println("complete - calculating probability for ", word)
	var currentString = nGramString + " " + word

	// search for respective strings in text corpus
	var countOfTheCurrentString = 0 // this is car
	var countOfBasePrevString = 0   // this is

	for _, sentence := range corpus {
		fmt.Println("current sentence for prob : ", sentence, "/", currentString)
		if strings.Contains(sentence, currentString) {
			countOfTheCurrentString += 1
		}

		if strings.Contains(sentence, nGramString) {
			countOfBasePrevString += 1
		}
	}

	// calculating probability
	var p float32 = 0
	if countOfBasePrevString > 0 {
		p = float32(countOfTheCurrentString) / float32(countOfBasePrevString)
	}
	return p
}

func createMapforSentence(corpus []string, corpusSentence int, window int,
	currentMap map[string][]bigramNextValue) {

	var splitWords = strings.Split(corpus[corpusSentence], " ")
	fmt.Println("split sentence : ", splitWords)
	for i := 0; i < len(splitWords)-(window-1); i++ {
		// convert to n-gram
		var currentBiGramString = splitWords[i] + " " + splitWords[i+1]

		// get the next word
		var nextWords = currentMap[currentBiGramString]
		if len(nextWords) == 0 {
			nextWords = make([]bigramNextValue, 0)
		}
		var nextWord = bigramNextValue{"", 0}
		if i+2 >= len(splitWords) { // convert to n-gram
			nextWord.word = "#" // terminating string
		} else {
			nextWord.word = splitWords[i+2] // convert to n-gram
		}

		// calculate probability of the word
		nextWord.prob = calculateWordProbability(corpus, currentBiGramString, nextWord.word)

		//add if the word is not already present
		var wordAlreadySeen = false
		for _, word := range nextWords {
			if nextWord == word {
				wordAlreadySeen = true
				break
			}
		}

		if !wordAlreadySeen {
			nextWords = append(nextWords, nextWord)
		}
		currentMap[currentBiGramString] = nextWords
	}
}

func createNGramNextWordMap(corpus []string, window int) map[string][]bigramNextValue {
	fmt.Println("preparing to create bigram next word map")
	if window == 0 {
		window = 2 // bigram model
	}
	var m = make(map[string][]bigramNextValue)

	for i := 0; i < len(corpus); i++ {
		var currentSentence string = corpus[i]
		fmt.Println("Tokenizing sentence : ", currentSentence)
		createMapforSentence(corpus, i, window, m)
	}
	fmt.Println("These are the values of map:", m)
	return m
}

func predictNextWord(nGramNextWordMap map[string][]bigramNextValue, inputSentence string) string {
	fmt.Println("Predict next word for ", inputSentence)

	// get last 2 words of the input sentence
	var splitWords = strings.Split(inputSentence, " ")
	var splitWordsLen = len(splitWords)

	var biGramString = splitWords[splitWordsLen-2] + " " + splitWords[splitWordsLen-1]
	var possibleNextWords = nGramNextWordMap[biGramString]

	// find word with max prob
	var maxProbWord = bigramNextValue{"", -1}
	for _, value := range possibleNextWords {
		if value.prob > maxProbWord.prob {
			maxProbWord = value
		}
	}
	// var prediction = possibleNextWords[rand.Intn(len(possibleNextWords))].word
	fmt.Println("Predicted word :", maxProbWord.word, "/", maxProbWord.prob)
	var prediction = maxProbWord.word
	return prediction
}

func main() {
	corpus := getTestCorpus()
	fmt.Println("number of sentenses in the corpus: ", len(corpus))
	var m map[string][]bigramNextValue = createNGramNextWordMap(corpus, 2)
	fmt.Println("bigram map: ", m)

	fmt.Println("-----------------Prediction----------------")
	fmt.Println(predictNextWord(m, "this is a"))
	fmt.Println(predictNextWord(m, "the house"))
	fmt.Println(predictNextWord(m, "jack built"))
}
