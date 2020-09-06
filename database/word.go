package database

import "bytes"

var characterLookup = [256]byte{}

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

func words(s string) []string {
	var buffer bytes.Buffer
	buffer.Reset()
	var words []string
	for _, r := range s {
		var c byte = byte(r)

		if characterLookup[c] != 0 {
			buffer.WriteRune(rune(characterLookup[c]))
			continue
		}

		if buffer.Len() != 0 {
			words = append(words, buffer.String())
			buffer.Reset()
		}
	}

	if buffer.Len() != 0 {
		words = append(words, buffer.String())
	}

	return words
}
