package cloudguardian_crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
)

func ValidatePayload(publicKey, payload, signature string) (bool, error) {
	// Decode the public key from hex
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, err
	}

	// Decode the payload from hex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return false, err
	}

	// Decode the signature from hex
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	// Hash the payload using SHA-256
	hash := sha256.Sum256(payloadBytes)

	// Verify the signature using secp256k1
	valid := crypto.VerifySignature(pubKeyBytes, hash[:], signatureBytes)

	return valid, nil
}

// func main() {
// 	payload := `{"command":"curl ifconfig.me"}`
// 	signature := "dad5578ac66504e1e8173cda2edef04cebd68e1ef854f9bfec11f5dfe0f55e0510ebf5c4c4ce5dd51afc9f5964135d56ea32ab29554b225d7a45c024dfe25f61"
// 	publicKey := "049634ad321cac90c3f35a5f6563a0d71114c0e5728a3c47fe296a2e61071ed4385b8d8f19770deebebc8aa423242d7137f87986b32bff00c8d1b7c2b551ae4409"
// 	hash := sha256.Sum256([]byte(payload))
// 	payloadBytes := hash[:]
// 	signatureBytes, _ := hex.DecodeString(signature)
// 	publicKeyBytes, _ := hex.DecodeString(publicKey)
// 	if secp256k1.VerifySignature(publicKeyBytes, payloadBytes, signatureBytes) {
// 		println("Signature is valid")
// 	} else {
// 		println("Signature is invalid")
// 	}
// }
