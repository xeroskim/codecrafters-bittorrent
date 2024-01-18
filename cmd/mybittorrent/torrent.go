package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net"
	"net/url"

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

type Addr struct {
	Ip   net.IP
	Port uint16
}

func MakeTorrent(fileName string) *TorrentFile, err {
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

func handshake(t TorrentFile, peerAddr string) []byte, err {
	var syn []byte
	syn = append(syn, 19)
	syn = append(syn, "BitTorrent protocol"...)
	syn = append(syn, strings.Repeat("\x00", 8)...)
	syn = append(syn, t.Hash()...)
	syn = append(syn, "00112233445566778899"...)

	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, fmt.Errorf("Failed connect with peer address (%w)", err)
	}
	defer conn.Close()

	conn.Write(syn)

	rb := make([]byte, 4096)
	_, err = conn.Read(rb)
	if err != nil {
		return nil, fmt.Errorf("Failed read from peer address (%w)", err)
	}

	return rb, err
}

func (t* TorrentFile) GetPeers() []Addr, err {
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
	peerList := make([]Addr, 0, len(peers)/6)
	for i := 0; i < len(peers); i += 6 {
		peerList = append(peerList, Addr{
			Ip:   net.IP(peers[i : i+4]),
			Port: binary.BigEndian.Uint16(peers[i+4 : i+6]),
		})
	}

	return peerList, nil
}

func (t* TorrentFile) Hash() string {
	bencodedString := t.encode()
	h := sha1.New()
	h.Write([]byte(bencodedString))
	return string(h.Sum(nil))
}

func (t* TorrentFile) encode() string {
	b := bytes.NewBufferString("")

	err := bencode.Marshal(b, t.Info)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal file (%w)", err)
	}

	return b.String()
}

func (t* TorrentFile) decode(s string) interface{} {
	b := bytes.NewBufferString(s)
	decoded, err := bencode.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode bencoded string (%w)", err)
	}
	return decoded
}


