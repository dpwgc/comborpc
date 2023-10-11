package comborpc

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"log"
)

func int64ToBytes(num int64, len int) []byte {
	byteArray := make([]byte, len)
	binary.LittleEndian.PutUint64(byteArray, uint64(num))
	return byteArray
}

func bytesToInt64(bytes []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bytes[:]))
}

func doGzipBytes(data []byte) ([]byte, error) {
	var input bytes.Buffer
	wr := gzip.NewWriter(&input)
	defer func(wr *gzip.Writer) {
		err := wr.Close()
		if err != nil {
			log.Println(err)
		}
	}(wr)
	_, err := wr.Write(data)
	if err != nil {
		return nil, err
	}
	return input.Bytes(), nil
}

func unGzipBytes(data []byte) ([]byte, error) {
	var output bytes.Buffer
	var input bytes.Buffer
	input.Write(data)
	r, err := gzip.NewReader(&input)
	if err != nil {
		return nil, err
	}
	defer func(r *gzip.Reader) {
		err := r.Close()
		if err != nil {
			log.Println(err)
		}
	}(r)
	_, err = io.Copy(&output, r)
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
