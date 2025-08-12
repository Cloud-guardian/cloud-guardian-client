package cloudguardian_crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
)

func ValidatePayload(publicKey, payload, signature string) (bool, error) {
	// Decode the public key from hex
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, err
	}

	// Decode the signature from hex
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	// Hash the payload using SHA-256
	hash := sha256.Sum256([]byte(payload))

	// Verify the signature using secp256k1
	valid := crypto.VerifySignature(publicKeyBytes, hash[:], signatureBytes)

	return valid, nil
}
