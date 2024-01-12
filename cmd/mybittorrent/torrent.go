package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/structs"
)

type TorrentFile struct {
	Announce string
	Info     struct {
		Length      int
		Name        string
		PieceLength int `json:"piece length"`
		Pieces      string
	}
}

func TorrentInfo(fileName string) (string, int, string, error) {
	var t TorrentFile

	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("File open failed")
	}

	decodedTorrentInfo, _, err := decodeBencode(string(data))
	if err != nil {
		fmt.Println("decodeBencode failed")
	}

	pieces := decodedTorrentInfo.(map[string]interface{})["info"].(map[string]interface{})["pieces"].(string)

	jsonOutput, _ := json.Marshal(decodedTorrentInfo)
	fmt.Println(string(jsonOutput))
	err = json.Unmarshal(jsonOutput, &t)
	t.Info.Pieces = pieces

	fmt.Println("info pieces before encodeBencode")
	fmt.Println(t.Info.PieceLength)
	fmt.Println([]byte(t.Info.Pieces))

	encodedInfo, _ := encodeBencode(structs.Map(t.Info))
	fmt.Println(encodedInfo)

	h := sha1.New()
	h.Write([]byte(encodedInfo))
	bs := h.Sum(nil)

	return t.Announce, t.Info.Length, fmt.Sprintf("%x", bs), err
}
