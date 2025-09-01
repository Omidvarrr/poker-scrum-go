package utils

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/fernet/fernet-go"
)

type OAuthState struct {
	SessionID string `json:"sessionId"`
	ReturnURL string `json:"returnUrl"`
}

var stateKey []byte

func SetOAuthStateSecret(secret []byte) {
	stateKey = secret
}

func EncryptState(s OAuthState) (string, error) {
	if len(stateKey) == 0 {
		return "", errors.New("state secret not set")
	}
	k := fernet.Key{}
	copy(k[:], deriveFernetKey(stateKey))
	b, _ := json.Marshal(s)
	ct, err := fernet.EncryptAndSign(b, &k)
	if err != nil {
		return "", err
	}
	return string(ct), nil
}

func DecryptState(token string) (OAuthState, error) {
	var out OAuthState
	if len(stateKey) == 0 {
		return out, errors.New("state secret not set")
	}
	k := fernet.Key{}
	copy(k[:], deriveFernetKey(stateKey))
	msg := fernet.VerifyAndDecrypt([]byte(token), 5*time.Minute, []*fernet.Key{&k})
	if msg == nil {
		return out, errors.New("invalid state")
	}
	if err := json.Unmarshal(msg, &out); err != nil {
		return out, err
	}
	return out, nil
}

// deriveFernetKey normalizes any length input secret into a 32-byte key for Fernet
func deriveFernetKey(src []byte) []byte {
	// Simple zero-padded/truncated normalization
	out := make([]byte, 32)
	copy(out, src)
	return out
}
