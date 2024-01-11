package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	// Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeStr(bencodedString)
	} else if strings.HasPrefix(bencodedString, "i") {
		return decodeInt(bencodedString)
	} else if strings.HasPrefix(bencodedString, "l") {
		return decodeList(bencodedString)
	} else {
		return "", 0, fmt.Errorf("Only strings are supported at the moment")
	}
}

func decodeList(bencodedString string) (interface{}, int, error) {
	bencodedList := []interface{}{}
	var singleBencoded interface{}
	var singleBencodedLen int
	var listLen int
	var err error

	for i := 1; i < len(bencodedString); i += singleBencodedLen {
		if bencodedString[i] == 'e' {
			break
		}

		singleBencoded, singleBencodedLen, err = decodeBencode(bencodedString[i:])
		if err != nil {
			return "", 0, err
		}

		listLen += singleBencodedLen
		bencodedList = append(bencodedList, singleBencoded)
	}

	return bencodedList, listLen + 2, err
}

func decodeStr(bencodedString string) (interface{}, int, error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], len(lengthStr) + 1 + length, nil
}

func decodeInt(bencodedString string) (interface{}, int, error) {
	var endMarkerIndex int

	for i := 1; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			endMarkerIndex = i
			break
		}
	}
	decodedInt, err := strconv.Atoi(bencodedString[1:endMarkerIndex])
	if err != nil {
		return "", 0, err
	}

	return decodedInt, endMarkerIndex + 1, nil
}

func main() {

	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
