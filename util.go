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

func doGzip(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func unGzip(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(data)
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer func(r *gzip.Reader) {
		err = r.Close()
		if err != nil {
			log.Println(err)
		}
	}(r)
	var out bytes.Buffer
	_, err = io.Copy(&out, r)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func copyStringSlice(src []string) []string {
	return append([]string(nil), src...)
}

func copyMethodFuncSlice(src []MethodFunc) []MethodFunc {
	return append([]MethodFunc(nil), src...)
}
