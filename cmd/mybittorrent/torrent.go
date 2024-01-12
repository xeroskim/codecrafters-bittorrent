package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type TorrentFile struct {
	Announce string
	Info     struct {
		Length      int
		Name        string
		PieceLength int
		Pieces      string
	}
}

func TorrentInfo(fileName string) (string, int, error) {
	var t TorrentFile

	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("File open failed")
	}

	decodedTorrentInfo, _, err := decodeBencode(string(data))
	if err != nil {
		fmt.Println("decodeBencode failed")
	}

	jsonOutput, _ := json.Marshal(decodedTorrentInfo)
	json.Unmarshal(jsonOutput, &t)

	return t.Announce, t.Info.Length, err
}
