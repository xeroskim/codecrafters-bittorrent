package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

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

func MakeTorrent(fileName string) (*TorrentFile, error) {
	var t TorrentFile

	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file (%w)", err)
	}

	err = bencode.Unmarshal(f, &t)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal file (%w)", err)
	}

	return &t, err
}

func (t *TorrentFile) handshake(conn net.Conn) ([]byte, error) {
	var syn []byte
	syn = append(syn, 19)
	syn = append(syn, "BitTorrent protocol"...)
	syn = append(syn, strings.Repeat("\x00", 8)...)
	syn = append(syn, t.Hash()...)
	syn = append(syn, "00112233445566778899"...)

	conn.Write(syn)

	rb := make([]byte, 68)
	_, err := conn.Read(rb)
	if err != nil {
		return nil, fmt.Errorf("Failed read from peer address (%w)", err)
	}

	return rb, err
}

func (t *TorrentFile) GetPeers() ([]string, error) {
	v := url.Values{}
	v.Set("info_hash", t.Hash())
	v.Add("peer_id", "00112233445566778899")
	v.Add("port", "6881")
	v.Add("uploaded", "0")
	v.Add("downloaded", "0")
	v.Add("left", fmt.Sprint(t.Info.Length))
	v.Add("compact", "1")
	u := t.Announce + "?" + v.Encode()

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("Failed to send GET request (%w)", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body (%w)", err)
	}

	stringReader := strings.NewReader(string(b))
	decoded, err := bencode.Decode(stringReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode bencoded string (%w)", err)
	}

	peers := []byte(decoded.(map[string]interface{})["peers"].(string))
	peerList := make([]string, 0, len(peers)/6)
	for i := 0; i < len(peers); i += 6 {
		peerList = append(peerList, fmt.Sprintf(
			"%s:%d",
			net.IP(peers[i:i+4]),
			binary.BigEndian.Uint16(peers[i+4:i+6]),
		))
	}

	return peerList, nil
}

func (t *TorrentFile) DownloadPiece(conn net.Conn, pieceNum int) ([]byte, error) {
	_, err := t.handshake(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to handshake (%w)", err)
	}

	// bitfield message
	rb := make([]byte, 8)
	_, err = conn.Read(rb)
	if err != nil {
		return nil, fmt.Errorf("Failed read from peer address (%w)", err)
	}

	// interested message
	sb := []byte{0, 0, 0, 5, 2}
	conn.Write(sb)

	// unchoke message
	rb = make([]byte, 8)
	_, err = conn.Read(rb)

	// request message
	sb = make([]byte, 17)
	pb := make([]byte, 0)
	sb[3] = 17
	sb[4] = 6
	for i := 0; i < t.Info.PieceLength; i += 1 << 14 {
		length := math.Min(float64(t.Info.PieceLength-i), float64(16*1024))

		binary.BigEndian.PutUint32(sb[5:], uint32(pieceNum))
		binary.BigEndian.PutUint32(sb[9:], uint32(i))
		binary.BigEndian.PutUint32(sb[13:], uint32(length))
		conn.Write(sb)

		rb = make([]byte, uint32(length))
		conn.Read(rb)
		pb = append(pb, rb[13:]...)
	}

	return pb, err
}

func (t *TorrentFile) Hash() string {
	bencodedString := t.encode()
	h := sha1.New()
	h.Write([]byte(bencodedString))
	return string(h.Sum(nil))
}

func (t *TorrentFile) encode() string {
	b := bytes.NewBufferString("")

	bencode.Marshal(b, t.Info)

	return b.String()
}

func (t *TorrentFile) decode(s string) interface{} {
	b := bytes.NewBufferString(s)
	decoded, _ := bencode.Decode(b)

	return decoded
}
