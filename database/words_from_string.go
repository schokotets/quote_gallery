package database

import "bytes"

var characterLookup = [256]byte{}

// sets up the characterLookup table
// needs to be called before using getWordsFromString
func setupCharacterLookup() {
	for i := 0; i < 256; i++ {
		if '0' <= i && i <= '9' {
			characterLookup[i] = byte(i)
		}

		if 'A' <= i && i <= 'Z' {
			characterLookup[i] = byte(i) - 'A' + 'a'
		}

		if 'a' <= i && i <= 'z' {
			characterLookup[i] = byte(i)
		}

		// Special Characters
		switch i {
		case 'ä':
			characterLookup[i] = 'ä'
		case 'ö':
			characterLookup[i] = 'ö'
		case 'ü':
			characterLookup[i] = 'ü'
		case 'ß':
			characterLookup[i] = 'ß'
		case 'Ä':
			characterLookup[i] = 'ä'
		case 'Ö':
			characterLookup[i] = 'ö'
		case 'Ü':
			characterLookup[i] = 'ü'
		}
	}
}

// Processes a given string into a words
// How the string is sliced is mainly dependend on the characterLookup which itself is created by setupCharacterLookup
// setupCharacterLookup needs to have been called
func getWordsFromString(s string) map[string]uint32 {
	var buffer bytes.Buffer
	buffer.Reset()
	wordCountMap := make(map[string]uint32)
	for _, r := range s {
		var c byte = byte(r)

		if characterLookup[c] != 0 {
			buffer.WriteRune(rune(characterLookup[c]))
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
