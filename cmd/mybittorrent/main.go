package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	// Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeStr(bencodedString)
	} else if strings.HasPrefix(bencodedString, "i") {
		return decodeInt(bencodedString)
	} else if strings.HasPrefix(bencodedString, "l") && strings.HasSuffix(bencodedString, "e") {
		return decodeList(bencodedString)
	} else {
		return "", fmt.Errorf("Only strings are supported at the moment")
	}
}

func decodeList(bencodedString string) (interface{}, error) {
	bencodedList := []interface{}{}
	var err error

	bencodedString = bencodedString[1 : len(bencodedString)-1]

	if len(bencodedString) == 0 {
		return bencodedList, err
	}

	parsedLen := 0
	for i := 0; len(bencodedString) != 0; i++ {
		fmt.Printf("bencodedString : %s\n", bencodedString)

		singleBencode, err := decodeBencode(bencodedString)
		if err != nil {
			return "", err
		}

		bencodedList = append(bencodedList, singleBencode)

		bencodeType := reflect.TypeOf(singleBencode).Kind()
		if bencodeType == reflect.String {
			parsedLen = len(singleBencode.(string)) + 2
		} else if bencodeType == reflect.Int {
			parsedLen = len(strconv.Itoa(singleBencode.(int))) + 2
		} else {
			parsedLen = 12
		}

		bencodedString = bencodedString[parsedLen:]
	}

	return bencodedList, err
}

func decodeStr(bencodedString string) (interface{}, error) {
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
		return "", err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
}

func decodeInt(bencodedString string) (interface{}, error) {
	var singleBencode string

	fmt.Printf("bencodedString in decodeInt: %s\n", bencodedString)
	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			singleBencode = bencodedString[1:i]
			fmt.Printf("singleBencode : %s, i : %d\n", singleBencode, i)
			break
		}
	}
	decodedInt, err := strconv.Atoi(singleBencode)
	if err != nil {
		return "", err
	}

	return decodedInt, nil
}

func main() {

	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
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
