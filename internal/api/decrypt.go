package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const decryptionKey = "SimtVuagFbGR2K7P"

// decryptEpisodeSources decrypts the AES-256-CTR encrypted tobeparsed field
func decryptEpisodeSources(tobeparsed string) ([]EpisodeSource, error) {
	plaintext, err := decryptTobeparsed(tobeparsed)
	if err != nil {
		return nil, fmt.Errorf("decrypt tobeparsed: %w", err)
	}

	var resp decryptedEpisodeResponse
	if err := json.Unmarshal(plaintext, &resp); err != nil {
		return nil, fmt.Errorf("parse decrypted response: %w", err)
	}

	return resp.Episode.SourceUrls, nil
}

// decryptTobeparsed performs AES-256-CTR decryption on the base64-encoded payload
func decryptTobeparsed(blob string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	if len(data) < 28 {
		return nil, fmt.Errorf("encrypted data too short (%d bytes)", len(data))
	}

	// Extract IV (first 12 bytes) and ciphertext (skip last 16 bytes auth tag)
	iv := data[:12]
	ciphertext := data[12 : len(data)-16]

	// Key = SHA-256 hash of the secret
	keyHash := sha256.Sum256([]byte(decryptionKey))

	// Build CTR counter: 12-byte IV + 0x00000002 (big-endian)
	ctrIV := make([]byte, aes.BlockSize)
	copy(ctrIV, iv)
	ctrIV[12] = 0
	ctrIV[13] = 0
	ctrIV[14] = 0
	ctrIV[15] = 2

	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	stream := cipher.NewCTR(block, ctrIV)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}
