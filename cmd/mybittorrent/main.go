package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/jackpal/bencode-go"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		stringReader := strings.NewReader(bencodedValue)
		decoded, err := bencode.Decode(stringReader)
		check(err)

		jsonOutput, err := json.Marshal(decoded)
		check(err)

		fmt.Println(string(jsonOutput))
	case "info":
		fileName := os.Args[2]

		t, err := MakeTorrent(fileName)
		check(err)

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
		check(err)

		peerList, err := t.GetPeers()
		check(err)

		for _, addr := range peerList {
			fmt.Println(addr)
		}

	case "handshake":
		fileName := os.Args[2]
		peerAddr := os.Args[3]

		t, err := MakeTorrent(fileName)
		check(err)

		conn, err := net.Dial("tcp", peerAddr)
		defer conn.Close()
		rb, err := t.handshake(conn)
		check(err)

		fmt.Printf("Peer ID: %x\n", rb[48:68])
	case "download_piece":
		outName := os.Args[3]
		fileName := os.Args[4]
		pieceNum, _ := strconv.Atoi(os.Args[5])

		t, err := MakeTorrent(fileName)
		check(err)

		peerList, err := t.GetPeers()
		check(err)

		conn, err := net.Dial("tcp", peerList[1])
		defer conn.Close()
		pieceData, err := t.DownloadPiece(conn, pieceNum)
		check(err)

		err = ioutil.WriteFile(outName, pieceData, 0644)
		check(err)

		fmt.Printf("Piece %d downloaded to %s.\n", pieceNum, outName)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
