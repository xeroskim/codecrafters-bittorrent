package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"

	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce string
	Info     struct {
		Length      int    `bencode:"length"`
		Name        string `bencode:"name"`
		PieceLength int    `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
	}
}

func (t TorrentFile) hash() string {
	bencodedString := t.encode()
	h := sha1.New()
	h.Write([]byte(bencodedString))
	return string(h.Sum(nil))
}

func (t TorrentFile) encode() string {
	b := bytes.NewBufferString("")

	err := bencode.Marshal(b, t.Info)
	if err != nil {
		fmt.Println(err)
	}

	return b.String()
}

func (t TorrentFile) decode(s string) any {
	b := bytes.NewBufferString(s)
	decoded, _ := bencode.Decode(b)
	return decoded
}
