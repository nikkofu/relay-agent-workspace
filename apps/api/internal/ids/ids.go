package ids

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func NewPrefixedUUID(prefix string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "-" + time.Now().UTC().Format("20060102150405.000000")
	}

	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(buf)
	return prefix + "-" + encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}
