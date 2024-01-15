package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
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

		f, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
		}

		var t TorrentFile
		err = bencode.Unmarshal(f, &t)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Tracker URL: %s\n", t.Announce)
		fmt.Printf("Length: %d\n", t.Info.Length)
		fmt.Printf("Info Hash: %x\n", t.hash())

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
