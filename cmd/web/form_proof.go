package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var errInvalidFormProof = errors.New("invalid form proof")

type formProofPayload struct {
	Form     string `json:"form"`
	IssuedAt int64  `json:"issued_at"`
	Nonce    string `json:"nonce"`
}

type formProofSigner struct {
	key       []byte
	now       func() time.Time
	random    io.Reader
	minAge    time.Duration
	maxAge    time.Duration
	clockSkew time.Duration
}

func newFormProofSigner(key []byte) *formProofSigner {
	return &formProofSigner{
		key:       key,
		now:       time.Now,
		random:    rand.Reader,
		minAge:    2 * time.Second,
		maxAge:    2 * time.Hour,
		clockSkew: 30 * time.Second,
	}
}

func (s *formProofSigner) Create(form string) (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := io.ReadFull(s.random, nonceBytes); err != nil {
		return "", fmt.Errorf("generating form proof nonce: %w", err)
	}
	payload, err := json.Marshal(formProofPayload{
		Form:     form,
		IssuedAt: s.now().Unix(),
		Nonce:    base64.RawURLEncoding.EncodeToString(nonceBytes),
	})
	if err != nil {
		return "", fmt.Errorf("encoding form proof: %w", err)
	}

	signature := s.sign(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(signature), nil
}

func (s *formProofSigner) Verify(token, expectedForm string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errInvalidFormProof
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return errInvalidFormProof
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, s.sign(payload)) {
		return errInvalidFormProof
	}

	var proof formProofPayload
	if err := json.Unmarshal(payload, &proof); err != nil {
		return errInvalidFormProof
	}
	if proof.Form != expectedForm || proof.Nonce == "" {
		return errInvalidFormProof
	}
	nonce, err := base64.RawURLEncoding.DecodeString(proof.Nonce)
	if err != nil || len(nonce) != 16 {
		return errInvalidFormProof
	}

	now := s.now()
	issuedAt := time.Unix(proof.IssuedAt, 0)
	if issuedAt.After(now.Add(s.clockSkew)) {
		return errInvalidFormProof
	}
	age := now.Sub(issuedAt)
	if age < s.minAge || age > s.maxAge {
		return errInvalidFormProof
	}
	return nil
}

func (s *formProofSigner) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, s.key)
	mac.Write(payload)
	return mac.Sum(nil)
}

func decodeFormHMACKey(value string) ([]byte, error) {
	key, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decoding FORM_HMAC_KEY: %w", err)
	}
	if len(key) < 32 {
		return nil, fmt.Errorf("FORM_HMAC_KEY must decode to at least %s bytes", strconv.Itoa(32))
	}
	return key, nil
}
