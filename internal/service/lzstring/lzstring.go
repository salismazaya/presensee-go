package lzstring

import (
	"errors"
	"strings"
)

const keyStrBase64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="

func getBaseValue(alphabet string, character byte) int {
	return strings.IndexByte(alphabet, character)
}

type lzData struct {
	val      int
	position int
	index    int
}

func decompress(length int, resetValue int, getNextValue func(int) int) (string, error) {
	if length == 0 {
		return "", nil
	}

	dictionary := make(map[int]string)
	enlargeIn := 4
	dictSize := 4
	numBits := 3
	var result strings.Builder

	data := &lzData{
		val:      getNextValue(0),
		position: resetValue,
		index:    1,
	}

	for i := 0; i < 3; i++ {
		dictionary[i] = string(rune(i))
	}

	readBits := func(nBits int) int {
		bits := 0
		maxpower := 1 << nBits
		power := 1
		for power != maxpower {
			resb := data.val & data.position
			data.position >>= 1
			if data.position == 0 {
				data.position = resetValue
				data.val = getNextValue(data.index)
				data.index++
			}
			if resb > 0 {
				bits |= power
			}
			power <<= 1
		}
		return bits
	}

	next := readBits(2)
	var c string
	switch next {
	case 0:
		c = string(rune(readBits(8)))
	case 1:
		c = string(rune(readBits(16)))
	case 2:
		return "", nil
	}

	dictionary[3] = c
	w := c
	result.WriteString(c)

	for {
		if data.index > length+1 {
			return "", nil
		}

		bits := readBits(numBits)
		cVal := bits

		switch cVal {
		case 0:
			bits = readBits(8)
			dictionary[dictSize] = string(rune(bits))
			dictSize++
			cVal = dictSize - 1
			enlargeIn--
		case 1:
			bits = readBits(16)
			dictionary[dictSize] = string(rune(bits))
			dictSize++
			cVal = dictSize - 1
			enlargeIn--
		case 2:
			return result.String(), nil
		}

		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}

		var entry string
		if val, ok := dictionary[cVal]; ok {
			entry = val
		} else {
			if cVal == dictSize {
				entry = w + string([]rune(w)[0])
			} else {
				return "", errors.New("invalid compressed stream")
			}
		}

		result.WriteString(entry)

		dictionary[dictSize] = w + string([]rune(entry)[0])
		dictSize++
		enlargeIn--

		w = entry

		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}
	}
}

func CompressToBase64(uncompressed string) string {
	if uncompressed == "" {
		return ""
	}
	res := compress(uncompressed, 6, func(a int) byte {
		return keyStrBase64[a]
	})
	end := len(res) % 4
	if end > 0 {
		res += strings.Repeat("=", 4-end)
	}
	return res
}

func compress(uncompressed string, bitsPerChar int, getCharFromInt func(int) byte) string {
	if uncompressed == "" {
		return ""
	}

	contextDictionary := make(map[string]int)
	contextDictionaryToCreate := make(map[string]bool)
	contextC := ""
	contextW := ""
	contextEnlargeIn := 2
	contextDictSize := 3
	contextNumBits := 2
	var contextData []byte
	contextDataVal := 0
	contextDataPosition := 0

	runes := []rune(uncompressed)
	for ii := 0; ii < len(runes); ii++ {
		contextC = string(runes[ii])
		if _, ok := contextDictionary[contextC]; !ok {
			contextDictionary[contextC] = contextDictSize
			contextDictSize++
			contextDictionaryToCreate[contextC] = true
		}

		contextWC := contextW + contextC
		if _, ok := contextDictionary[contextWC]; ok {
			contextW = contextWC
		} else {
			if contextDictionaryToCreate[contextW] {
				firstRune := []rune(contextW)[0]
				if firstRune < 256 {
					for i := 0; i < contextNumBits; i++ {
						contextDataVal = (contextDataVal << 1)
						if contextDataPosition == bitsPerChar-1 {
							contextDataPosition = 0
							contextData = append(contextData, getCharFromInt(contextDataVal))
							contextDataVal = 0
						} else {
							contextDataPosition++
						}
					}
					value := int(firstRune)
					for i := 0; i < 8; i++ {
						contextDataVal = (contextDataVal << 1) | (value & 1)
						if contextDataPosition == bitsPerChar-1 {
							contextDataPosition = 0
							contextData = append(contextData, getCharFromInt(contextDataVal))
							contextDataVal = 0
						} else {
							contextDataPosition++
						}
						value = value >> 1
					}
				} else {
					value := 1
					for i := 0; i < contextNumBits; i++ {
						contextDataVal = (contextDataVal << 1) | value
						if contextDataPosition == bitsPerChar-1 {
							contextDataPosition = 0
							contextData = append(contextData, getCharFromInt(contextDataVal))
							contextDataVal = 0
						} else {
							contextDataPosition++
						}
						value = 0
					}
					value = int(firstRune)
					for i := 0; i < 16; i++ {
						contextDataVal = (contextDataVal << 1) | (value & 1)
						if contextDataPosition == bitsPerChar-1 {
							contextDataPosition = 0
							contextData = append(contextData, getCharFromInt(contextDataVal))
							contextDataVal = 0
						} else {
							contextDataPosition++
						}
						value = value >> 1
					}
				}
				contextEnlargeIn--
				if contextEnlargeIn == 0 {
					contextEnlargeIn = 1 << contextNumBits
					contextNumBits++
				}
				delete(contextDictionaryToCreate, contextW)
			} else {
				value := contextDictionary[contextW]
				for i := 0; i < contextNumBits; i++ {
					contextDataVal = (contextDataVal << 1) | (value & 1)
					if contextDataPosition == bitsPerChar-1 {
						contextDataPosition = 0
						contextData = append(contextData, getCharFromInt(contextDataVal))
						contextDataVal = 0
					} else {
						contextDataPosition++
					}
					value = value >> 1
				}
			}

			contextEnlargeIn--
			if contextEnlargeIn == 0 {
				contextEnlargeIn = 1 << contextNumBits
				contextNumBits++
			}

			contextDictionary[contextWC] = contextDictSize
			contextDictSize++
			contextW = contextC
		}
	}

	if contextW != "" {
		if contextDictionaryToCreate[contextW] {
			firstRune := []rune(contextW)[0]
			if firstRune < 256 {
				for i := 0; i < contextNumBits; i++ {
					contextDataVal = (contextDataVal << 1)
					if contextDataPosition == bitsPerChar-1 {
						contextDataPosition = 0
						contextData = append(contextData, getCharFromInt(contextDataVal))
						contextDataVal = 0
					} else {
						contextDataPosition++
					}
				}
				value := int(firstRune)
				for i := 0; i < 8; i++ {
					contextDataVal = (contextDataVal << 1) | (value & 1)
					if contextDataPosition == bitsPerChar-1 {
						contextDataPosition = 0
						contextData = append(contextData, getCharFromInt(contextDataVal))
						contextDataVal = 0
					} else {
						contextDataPosition++
					}
					value = value >> 1
				}
			} else {
				value := 1
				for i := 0; i < contextNumBits; i++ {
					contextDataVal = (contextDataVal << 1) | value
					if contextDataPosition == bitsPerChar-1 {
						contextDataPosition = 0
						contextData = append(contextData, getCharFromInt(contextDataVal))
						contextDataVal = 0
					} else {
						contextDataPosition++
					}
					value = 0
				}
				value = int(firstRune)
				for i := 0; i < 16; i++ {
					contextDataVal = (contextDataVal << 1) | (value & 1)
					if contextDataPosition == bitsPerChar-1 {
						contextDataPosition = 0
						contextData = append(contextData, getCharFromInt(contextDataVal))
						contextDataVal = 0
					} else {
						contextDataPosition++
					}
					value = value >> 1
				}
			}
			contextEnlargeIn--
			if contextEnlargeIn == 0 {
				contextEnlargeIn = 1 << contextNumBits
				contextNumBits++
			}
			delete(contextDictionaryToCreate, contextW)
		} else {
			value := contextDictionary[contextW]
			for i := 0; i < contextNumBits; i++ {
				contextDataVal = (contextDataVal << 1) | (value & 1)
				if contextDataPosition == bitsPerChar-1 {
					contextDataPosition = 0
					contextData = append(contextData, getCharFromInt(contextDataVal))
					contextDataVal = 0
				} else {
					contextDataPosition++
				}
				value = value >> 1
			}
		}
	}

	contextEnlargeIn--
	if contextEnlargeIn == 0 {
		contextEnlargeIn = 1 << contextNumBits
		contextNumBits++
	}

	// Mark end of stream
	value := 2
	for i := 0; i < contextNumBits; i++ {
		contextDataVal = (contextDataVal << 1) | (value & 1)
		if contextDataPosition == bitsPerChar-1 {
			contextDataPosition = 0
			contextData = append(contextData, getCharFromInt(contextDataVal))
			contextDataVal = 0
		} else {
			contextDataPosition++
		}
		value = value >> 1
	}

	// Flush
	for {
		contextDataVal = (contextDataVal << 1)
		if contextDataPosition == bitsPerChar-1 {
			contextData = append(contextData, getCharFromInt(contextDataVal))
			break
		} else {
			contextDataPosition++
		}
	}

	return string(contextData)
}

func DecompressFromBase64(compressed string) (string, error) {
	if compressed == "" {
		return "", nil
	}
	clean := strings.TrimRight(compressed, "=")
	return decompress(len(clean), 32, func(index int) int {
		if index >= len(clean) {
			return 0
		}
		val := getBaseValue(keyStrBase64, clean[index])
		if val < 0 {
			return 0
		}
		return val
	})
}
