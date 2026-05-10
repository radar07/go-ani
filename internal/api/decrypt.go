package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
)

func parseDecryptedSources(data string) ([]EpisodeSource, error) {
	re := regexp.MustCompile(`sourceUrl":"--([^"]+)".*?sourceName":"([^"]+)"`)

	matches := re.FindAllStringSubmatch(data, -1)

	var sources []EpisodeSource

	for _, m := range matches {
		sources = append(sources, EpisodeSource{
			SourceURL:  "--" + m[1],
			SourceName: m[2],
		})
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources parsed")
	}

	return sources, nil
}

// decryptEpisodeSources decrypts the AES-256-CTR encrypted tobeparsed field
func decryptEpisodeSources(tobeparsed string) ([]EpisodeSource, error) {
	plaintext, err := decryptTobeparsed(tobeparsed)
	if err != nil {
		return nil, fmt.Errorf("decrypt tobeparsed: %w", err)
	}

	return parseDecryptedSources(string(plaintext))
}

// decryptTobeparsed performs AES-256-CTR decryption on the base64-encoded payload
func decryptTobeparsed(blob string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return nil, err
	}

	if len(data) < 29 {
		return nil, fmt.Errorf("invalid encrypted payload")
	}

	iv := data[1:13]
	ciphertext := data[13 : len(data)-16]

	keyHash := sha256.Sum256([]byte("Xot36i3lK3:v1"))

	ctrIV := make([]byte, aes.BlockSize)
	copy(ctrIV, iv)
	ctrIV[12] = 0
	ctrIV[13] = 0
	ctrIV[14] = 0
	ctrIV[15] = 2

	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, ctrIV)

	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}
