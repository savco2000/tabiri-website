package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormProofCreateAndVerify(t *testing.T) {
	now := time.Date(2026, time.October, 14, 16, 0, 0, 0, time.UTC)
	signer := newFormProofSigner([]byte("01234567890123456789012345678901"))
	signer.now = func() time.Time { return now }
	signer.random = bytes.NewReader(bytes.Repeat([]byte{1}, 16))

	token, err := signer.Create("contact")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(3 * time.Second)
	if err := signer.Verify(token, "contact"); err != nil {
		t.Fatalf("valid form proof rejected: %v", err)
	}
	if err := signer.Verify(token, "partner_contact"); !errors.Is(err, errInvalidFormProof) {
		t.Fatalf("wrong form returned %v, want errInvalidFormProof", err)
	}

	parts := strings.Split(token, ".")
	tampered := parts[0] + "." + parts[1][:len(parts[1])-1] + "A"
	if err := signer.Verify(tampered, "contact"); !errors.Is(err, errInvalidFormProof) {
		t.Fatalf("tampered proof returned %v, want errInvalidFormProof", err)
	}
}

func TestFormProofTimingWindows(t *testing.T) {
	issuedAt := time.Date(2026, time.October, 14, 16, 0, 0, 0, time.UTC)
	now := issuedAt
	signer := newFormProofSigner([]byte("01234567890123456789012345678901"))
	signer.now = func() time.Time { return now }
	signer.random = bytes.NewReader(bytes.Repeat([]byte{2}, 16))

	token, err := signer.Create("contact")
	if err != nil {
		t.Fatal(err)
	}
	if err := signer.Verify(token, "contact"); !errors.Is(err, errInvalidFormProof) {
		t.Fatalf("too-fast proof returned %v, want errInvalidFormProof", err)
	}

	now = issuedAt.Add(2*time.Hour + time.Second)
	if err := signer.Verify(token, "contact"); !errors.Is(err, errInvalidFormProof) {
		t.Fatalf("expired proof returned %v, want errInvalidFormProof", err)
	}

	now = issuedAt.Add(-31 * time.Second)
	if err := signer.Verify(token, "contact"); !errors.Is(err, errInvalidFormProof) {
		t.Fatalf("future proof returned %v, want errInvalidFormProof", err)
	}
}

func TestDecodeFormHMACKeyRequiresStrongKey(t *testing.T) {
	if _, err := decodeFormHMACKey("c2hvcnQ"); err == nil {
		t.Fatal("expected a short HMAC key to be rejected")
	}
}
