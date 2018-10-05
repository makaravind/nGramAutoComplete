package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type MessageData struct {
	Content    string `json:"content"`
	SenderName string `json:"sender_name"`
}
type Data struct {
	Messages []MessageData `json:"messages"`
}

type bigramNextValue struct {
	word string
	prob float32
}

// create a text corpus divided by sentenses
// create bigram and next word map
// while creating next word calculate the P(w1/ prev 2 words)

// this will be the base model

// v2- for any new bigrams for every user search update the mapped next world prrob

/* func getTestCorpus() []string {
	corpus := make([]string, 0)
	corpus = append(corpus, "this is a the house that jack built")
	corpus = append(corpus, "this is the malt")
	corpus = append(corpus, "this is a rat")
	corpus = append(corpus, "this is a cat")
	corpus = append(corpus, "that killed the rat")
	corpus = append(corpus, "that ate the malt")
	return corpus
}
*/
func sanitizeWord(word string) string {
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		fmt.Print("something went wrong while creating regex")
		return ""
	}
	// remove characters apart from alpha bets
	word = reg.ReplaceAllString(word, "")
	word = strings.ToLower(word)
	return word
}

func sanitizeSentence(sentence string, nGramn int) string {
	// santize word-by-word
	var inSentence = strings.Trim(sentence, " ")
	var splitWords = strings.Split(inSentence, " ")

	var sanitizedSentence = ""
	for _, word := range splitWords {
		// remove characters apart from alpha bet
		word = sanitizeWord(word)
		if len(word) > 0 {
			sanitizedSentence = sanitizedSentence + " " + word
		}
	}
	return sanitizedSentence
}

func getMessageDataFromFile(fileName string) []MessageData {
	contentBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Print("Error reading file")
	}

	var data Data

	//  var data interface{}
	err1 := json.Unmarshal(contentBytes, &data)

	if err1 != nil {
		fmt.Print("Error unmarshaling")
	}

	return data.Messages
}

func isValidateSentence(currSentence string, nGramn int) bool {
	if currSentence == "" || len(strings.Split(currSentence, " ")) < nGramn {
		return false
	}
	return true
}

func getMessageFilePaths(baseFolder string) []string {
	files, err := ioutil.ReadDir(baseFolder)

	var paths = make([]string, 0)
	if err != nil {
		fmt.Println("Something went wrong reading the directory for messages")
		return paths
	}

	for _, f := range files {
		paths = append(paths, f.Name())
	}
	return paths
}

func updateTextCorpusWithMessageData(corpus []string, messagesData []MessageData, nGramn int, forParticipant string) []string {
	for _, message := range messagesData {
		// fmt.Println("Evaluting current ", message)
		if message.SenderName == forParticipant {
			var corpusSentence = sanitizeSentence(message.Content, nGramn)
			if !isValidateSentence(corpusSentence, nGramn) {
				continue
			}
			// fmt.Println("Sanitized sentence", corpusSentence)
			corpus = append(corpus, corpusSentence)
		}
	}
	return corpus
}

func getTestCorpus(forParticipant string, nGramn int, messagesFileBasePathRelative string) []string {
	var capacity = 1000
	corpus := make([]string, 0, capacity)

	// get all message file paths
	var directives = getMessageFilePaths(messagesFileBasePathRelative)

	for _, path := range directives {
		var fileName = messagesFileBasePathRelative + path + "/message.json"
		var messagesData = getMessageDataFromFile(fileName)
		corpus = updateTextCorpusWithMessageData(corpus, messagesData, nGramn, forParticipant)
	}

	// fmt.Println("final corpus ", corpus[0], ",", corpus[1])
	fmt.Println("final corpus ", corpus)
	return corpus
}

func calculateWordProbability(corpus []string, nGramString string, word string) float32 {
	if word == "#" {
		return 0 // end of the sentence; this should be chosen at the end
	}

	// fmt.Println("complete - calculating probability for ", word)
	var currentString = nGramString + " " + word

	// search for respective strings in text corpus
	var countOfTheCurrentString = 0 // this is car
	var countOfBasePrevString = 0   // this is

	for _, sentence := range corpus {
		// fmt.Println("current sentence for prob : ", sentence, "/", currentString)
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
		// fmt.Println("Tokenizing sentence : ", corpus[i])
		createMapforSentence(corpus, i, window, m)
	}
	// fmt.Println("These are the values of map:", m)
	return m
}

func predictNextWord(nGramNextWordMap map[string][]bigramNextValue, nGramn int, inputSentence string) string {

	var sanitizedSentence = sanitizeSentence(inputSentence, nGramn)

	fmt.Println("Predict next word for ", inputSentence, " / ", sanitizedSentence)

	// get last 2 words of the input sentence
	var splitWords = strings.Split(sanitizedSentence, " ")
	var splitWordsLen = len(splitWords)

	fmt.Println("predictiong split words", splitWords)

	if splitWordsLen <= nGramn {
		return "invalid, nothing predicted!"
	}

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

func predictNextWordTillEnd(m map[string][]bigramNextValue, nGramn int, seed string) string {
	fullSentence := seed
	for {
		nextWord := predictNextWord(m, nGramn, fullSentence)
		fmt.Println("nextWord : ", nextWord)
		if nextWord == "#" || nextWord == ""{
			break
		}
		fullSentence += " " + nextWord
	}
	return fullSentence
}

func main() {
	nGramn := 2
	messageBasePath := "facebook-mak11195/messages_small/"
	corpus := getTestCorpus("Aravind Metku", nGramn, messageBasePath)

	if len(corpus) > 0 {
		fmt.Println("number of sentenses in the corpus: ", len(corpus))
		var m map[string][]bigramNextValue = createNGramNextWordMap(corpus, 2)
		fmt.Println("bigram map: ", m)

		fmt.Println("-----------------Prediction----------------")

		fullSentence := predictNextWordTillEnd(m, nGramn, "how are")
		fmt.Println(fullSentence)
	}

}
