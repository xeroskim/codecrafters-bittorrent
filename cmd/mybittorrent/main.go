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

		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(jsonOutput))
	case "info":
		fileName := os.Args[2]

		t, err := MakeTorrent(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Tracker URL: %s\n", t.Announce)
		fmt.Printf("Length: %d\n", t.Info.Length)
		fmt.Printf("Info Hash: %x\n", t.Hash())
		fmt.Printf("Piece Length: %d\n", t.Info.PieceLength)
		fmt.Printf("Piece Hashes:\n")

		for i := 0; i < len(t.Info.Pieces); i += 20 {
			fmt.Printf("%x\n", t.Info.Pieces[i:i+20])
		}
	case "peers":
		fileName := os.Args[2]

		t, err := MakeTorrent(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		p, err := t.GetPeers()
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, addr := range peerList {
			fmt.Printf("%s:%d\n", addr.Ip, addr.Port)
		}

	case "handshake":
		fileName := os.Args[2]
		peerAddr := os.Args[3]

		t, err := MakeTorrent(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		rb, err := handshake(t, peerAddr)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Peer ID: %x\n", rb[48:68])
	case "download_piece":
		outName := os.Args[3]
		fileName := os.Args[4]
		pieceNum := os.Args[5]

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
