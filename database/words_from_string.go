package database

import (
	"bytes"
	"unicode"
)

// Processes a given string into a words
// How the string is sliced is mainly dependend on the characterLookup which itself is created by setupCharacterLookup
// setupCharacterLookup needs to have been called
func getWordsFromString(s string) map[string]uint32 {
	var buffer bytes.Buffer
	buffer.Reset()
	wordCountMap := make(map[string]uint32)
	for _, r := range s {

		if unicode.IsLetter(r) {
			buffer.WriteRune(unicode.ToLower(r))
			continue
		}

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
