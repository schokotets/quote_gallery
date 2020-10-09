package database

import (
	"bytes"
	"unicode"
)

// turns a given string to a wordCountMap
func getWordsFromString(s string) map[string]uint32 {
	var buffer bytes.Buffer
	buffer.Reset()
	wordCountMap := make(map[string]uint32)

	for _, r := range s {

		if unicode.IsLetter(r) {
			buffer.WriteRune(unicode.ToLower(r))
			continue
		}

		// character is not a letter -> word ended
		if buffer.Len() != 0 {
			wordCountMap[buffer.String()] = wordCountMap[buffer.String()] + 1
			buffer.Reset()
		}
	}

	if buffer.Len() != 0 {
		wordCountMap[buffer.String()] = wordCountMap[buffer.String()] + 1
	}

	return wordCountMap
}
