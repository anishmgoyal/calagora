package utils

import (
	"bytes"

	"github.com/anishmgoyal/calagora/constants"
)

var stopWords map[string]int

func isStopWord(word string) bool {
	if stopWords == nil {
		stopWords = GetSearchTermsForString(constants.StopWords, false)
	}
	_, ok := stopWords[word]
	return ok
}

// GetSearchTermsForString extracts alphanumeric words from
// a string, with the requirement that each word is at least
// 3 letters long, and the first character is a letter.
func GetSearchTermsForString(s string, enforceStopwords bool) map[string]int {
	termMap := make(map[string]int)
	var buff bytes.Buffer

	for pos := 0; pos < len(s); pos++ {
		isFirst := buff.Len() == 0
		if isSearchChar(s[pos], isFirst) {
			buff.WriteByte(toLower(s[pos]))
		} else {
			if buff.Len() >= 3 {
				currTerm := buff.String()
				if !enforceStopwords || !isStopWord(currTerm) {
					termMap[currTerm] = termMap[currTerm] + 1
				}
			}
			buff.Reset()
		}
	}

	if buff.Len() >= 3 {
		currTerm := buff.String()
		if !enforceStopwords || !isStopWord(currTerm) {
			termMap[currTerm] = termMap[currTerm] + 1
		}
	}

	return termMap
}

func isSearchChar(c byte, isFirst bool) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9' && !isFirst)
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		c = c + 32
	}
	return c
}
