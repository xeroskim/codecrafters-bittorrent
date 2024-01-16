package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	case "peers":
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

		v := url.Values{}
		v.Set("info_hash", t.hash())
		v.Add("peer_id", "00112233445566778899")
		v.Add("port", "6881")
		v.Add("uploaded", "0")
		v.Add("downloaded", "0")
		v.Add("left", fmt.Sprint(t.Info.Length))
		v.Add("compact", "1")
		u := t.Announce + "?" + v.Encode()
		fmt.Println(u)
		fmt.Println(t.Info.Length)

		resp, err := http.Get(u)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		type TrackerResponse struct {
			Interval int
			Peers    string
		}

		decoder := json.NewDecoder(resp.Body)
		var r TrackerResponse
		err = decoder.Decode(&r)
		if err != nil {
			fmt.Println(err)
		}

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
