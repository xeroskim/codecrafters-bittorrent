package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

func main() {

	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		stringReader := strings.NewReader(bencodedValue)
		decoded, err := bencode.Decode(stringReader)
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
		fmt.Printf("Piece Length: %d\n", t.Info.PieceLength)
		fmt.Printf("Piece Hashes:\n")

		for i := 0; i < len(t.Info.Pieces); i += 20 {
			fmt.Printf("%x\n", t.Info.Pieces[i:i+20])
		}
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
