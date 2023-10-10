package comborpc

import "encoding/binary"

func int64ToBytes(num int64) []byte {
	byteArray := make([]byte, TCPHeaderLen)
	binary.LittleEndian.PutUint64(byteArray, uint64(num))
	return byteArray
}

func bytesToInt64(bytes []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bytes[:]))
}
