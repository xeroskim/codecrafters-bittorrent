package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func encodeBencode(decodedData interface{}) (string, error) {
	var bencodedDic strings.Builder
	var err error
	keys := [4]string{"Length", "Name", "PieceLength", "Pieces"}

	switch decodedData.(type) {
	case map[string]interface{}:
		bencodedDic.WriteString("d")

		for _, v := range keys {
			encodedKey, _ := encodeBencode(v)
			bencodedDic.WriteString(encodedKey)

			encodedValue, _ := encodeBencode(decodedData.(map[string]interface{})[v])
			bencodedDic.WriteString(encodedValue)
		}

		bencodedDic.WriteString("e")
	case []interface{}:
		bencodedDic.WriteString("l")

		for _, v := range decodedData.([]interface{}) {
			out, _ := encodeBencode(v)
			bencodedDic.WriteString(out)
		}

		bencodedDic.WriteString("e")
	case string:
		data := decodedData.(string)
		bencodedDic.WriteString(strconv.Itoa(len(data)))
		bencodedDic.WriteString(":")
		bencodedDic.WriteString(data)
	case int:
		data := decodedData.(int)
		bencodedDic.WriteString("i")
		bencodedDic.WriteString(strconv.Itoa(data))
		bencodedDic.WriteString("e")
	default:
		return "", fmt.Errorf("Wrong data format while encoding")
	}

	return bencodedDic.String(), err
}

func decodeBencode(bencodedString string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeStr(bencodedString)
	} else if strings.HasPrefix(bencodedString, "i") {
		return decodeInt(bencodedString)
	} else if strings.HasPrefix(bencodedString, "l") {
		return decodeList(bencodedString)
	} else if strings.HasPrefix(bencodedString, "d") {
		return decodeDictionary(bencodedString)
	} else {
		return "", 0, fmt.Errorf("Only strings are supported at the moment")
	}
}

func decodeDictionary(bencodedString string) (interface{}, int, error) {
	bencodedDic := map[string]interface{}{}
	var singleBencoded interface{}
	var key string
	var singleBencodedLen int
	var dicLen int
	var err error
	var isValue bool

	for i := 1; i < len(bencodedString); i += singleBencodedLen {
		if bencodedString[i] == 'e' {
			break
		}

		singleBencoded, singleBencodedLen, err = decodeBencode(bencodedString[i:])

		if !isValue {
			key = singleBencoded.(string)
			isValue = true
		} else {
			bencodedDic[key] = singleBencoded
			fmt.Printf("key is : %s\n", key)
			if reflect.TypeOf(singleBencoded).Kind() == reflect.String {
				fmt.Println([]byte(singleBencoded.(string)))
			}
			isValue = false
		}

		dicLen += singleBencodedLen
	}

	return bencodedDic, dicLen + 2, err
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
