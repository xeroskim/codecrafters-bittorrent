package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {

	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	case "info":
		fileName := os.Args[2]

		url, length, hash, err := TorrentInfo(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Tracker URL: %s\n", url)
		fmt.Printf("Length: %d\n", length)
		fmt.Printf("Info Hash: %s\n", hash)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
