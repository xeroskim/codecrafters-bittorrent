package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

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

	b, err := io.ReadAll(resp.Body)
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

type PieceInfo struct {
	pieceData  []byte
	pieceIndex int
}

func (t *TorrentFile) Download() ([]byte, error) {
	var wg sync.WaitGroup
	var fileData []byte
	pieceNum := int64(math.Ceil(float64(t.Info.Length) / float64(t.Info.PieceLength)))
	pieceDataList := make([][]byte, pieceNum)
	downloadQueue := make(chan int, pieceNum)

	//fmt.Printf("Download start! Total piece num : %d\n", pieceNum)

	for i := 0; i < int(pieceNum); i++ {
		downloadQueue <- i
	}

	peerList, err := t.GetPeers()
	if err != nil {
		return nil, err
	}

	errChannel := make(chan error, len(peerList))
	wg.Add(len(peerList))
	// goroutine per peer
	for _, peer := range peerList {
		go func(peer string) {
			defer wg.Done()

			for ; len(downloadQueue) != 0; {
				pieceIndex := <- downloadQueue

				conn, err := net.Dial("tcp", peer)
				if err != nil {
					downloadQueue <- pieceIndex
					errChannel <- err
					continue
				}

				pieceData, err := t.DownloadPiece(conn, pieceIndex)
				if err != nil {
					downloadQueue <- pieceIndex
					errChannel <- err
					continue
				}

				pieceDataList[pieceIndex] = pieceData
				fmt.Printf("Piece %d downloaded\n", pieceIndex)
				conn.Close()
			}
		}(peer)
	}

	wg.Wait()
	close(errChannel)

	for err := range errChannel {
		if err != nil {
			fmt.Println("error occured")
			return nil, err
		}
	}

	for i := 0; i < int(pieceNum); i++ {
		fileData = append(fileData, pieceDataList[i]...)
	}

	return fileData, nil
}

func (t *TorrentFile) DownloadPiece(conn net.Conn, pieceIndex int) ([]byte, error) {
	_, err := t.handshake(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to handshake (%w)", err)
	}

	// bitfield message.
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

	var PieceLength int
	if (pieceIndex+1)*t.Info.PieceLength > t.Info.Length {
		PieceLength = t.Info.Length - t.Info.PieceLength*pieceIndex
	} else {
		PieceLength = t.Info.PieceLength
	}

	for i := 0; i < PieceLength; i += 1 << 14 {
		length := math.Min(float64(PieceLength-i), float64(1<<14))

		binary.BigEndian.PutUint32(sb[5:], uint32(pieceIndex))
		binary.BigEndian.PutUint32(sb[9:], uint32(i))
		binary.BigEndian.PutUint32(sb[13:], uint32(length))
		conn.Write(sb)

		header := make([]byte, 5)
		if _, err := io.ReadFull(conn, header); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		bodyLength := binary.BigEndian.Uint32(header[0:4]) - 1
		body := make([]byte, bodyLength)
		if _, err := io.ReadFull(conn, body); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		pb = append(pb, body[8:]...)
	}

	h := sha1.New()
	h.Write(pb)
	hsum := string(h.Sum(nil))

	pieceHash := t.Info.Pieces[20*pieceIndex : 20*pieceIndex+20]
	if hsum != pieceHash {
		return nil, fmt.Errorf("Wrong hash got %x, want %x\n", hsum, pieceHash)
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
