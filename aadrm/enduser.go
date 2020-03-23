package aadrm

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Key struct {
	Value      *string `json:"Value,omitempty"`
	CipherMode *string `json:"CipherMode,omitempty"`
	Algorithm  *string `json:"Algorithm,omitempty"`
	Size       *int    `json:"Size,omitempty"`
}

// Decrypt data using this key
func (k *Key) Decrypt(ciphertext []byte) ([]byte, error) {
	if k == nil {
		return nil, errors.New("Key is nil")
	}

	if k.Algorithm == nil {
		return nil, errors.New("Algorithm is nil")
	} else if *k.Algorithm != "AES" {
		return nil, errors.Errorf("Unsupported Algorithm %s", *k.Algorithm)
	}
	if k.CipherMode == nil {
		return nil, errors.New("CipherMode is nil")
	} else if *k.CipherMode != "MICROSOFT.ECB" {
		return nil, errors.Errorf("Unsupported CipherMode %s", *k.CipherMode)
	}

	if k.Value == nil {
		return nil, errors.New("Value is nil")
	}
	value, err := base64.StdEncoding.DecodeString(*k.Value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to base64 decode Value")
	}

	cipher, err := aes.NewCipher(value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AES cipher")
	}

	if k.Size == nil {
		return nil, errors.New("Size is nil")
	} else if *k.Size != cipher.BlockSize() {
		return nil, errors.Errorf("Mismatched block size %d and %d", *k.Size, cipher.BlockSize())
	}

	// TODO: Hacky af, no idea if this is "correct", seems very wrong
	bs := cipher.BlockSize()
	ciphertext = ciphertext[len(ciphertext) % bs:]

	// Go "intentionally" never implemented ECB because it's insecure, implement by hand
	tmp := make([]byte, len(ciphertext))
	plaintext := tmp[:]
	for len(ciphertext) > 0 {
		cipher.Decrypt(tmp, ciphertext)
		tmp = tmp[bs:]
		ciphertext = ciphertext[bs:]
	}
	return plaintext, nil
}

type UserRight struct {
	Users  []string `json:"Users,omitempty"`
	Rights []string `json:"Rights,omitempty"`
}

type Policy struct {
	AllowAuditedExtraction *bool       `json:"AllowAuditedExtraction,omitempty"`
	UserRoles              []string    `json:"UserRoles,omitempty"`
	UserRights             []UserRight `json:"UserRights,omitempty"`
	IntervalTimeInDays     *int        `json:"IntervalTimeInDays,omitempty"`
	LicenseValidUntil      *string     `json:"LicenseValidUntil,omitempty"`
}

type EndUserLicense struct {
	ID                       *string                    `json:"Id,omitempty"`
	Name                     *string                    `json:"Name,omitempty"`
	Description              *string                    `json:"Description,omitempty"`
	Referrer                 *string                    `json:"Referrer,omitempty"`
	Owner                    *string                    `json:"Owner,omitempty"`
	AccessStatus             *string                    `json:"AccessStatus,omitempty"`
	Key                      *Key                       `json:"Key,omitempty"`
	Rights                   []string                   `json:"Rights,omitempty"`
	Roles                    []string                   `json:"Roles,omitempty"`
	IssuedTo                 *string                    `json:"IssuedTo,omitempty"`
	ContentValidUntil        *string                    `json:"ContentValidUntil,omitempty"`
	LicenseValidUntil        *string                    `json:"LicenseValidUntil,omitempty"`
	ContentID                *string                    `json:"ContentId,omitempty"`
	DocumentID               *string                    `json:"DocumentId,omitempty"`
	LabelID                  *string                    `json:"LabelId,omitempty"`
	OnlineAccessOnly         *bool                      `json:"OnlineAccessOnly,omitempty"`
	SignedApplicationData    map[string]json.RawMessage `json:"SignedApplicationData,omitempty"`
	EncryptedApplicationData map[string]json.RawMessage `json:"EncryptedApplicationData,omitempty"`
	FromTemplate             *bool                      `json:"FromTemplate,omitempty"`
	Policy                   *Policy                    `json:"Policy,omitempty"`
	ErrorMessage             *string                    `json:"ErrorMessage,omitempty"`
}

func (l *EndUserLicense) String() string {
	b, _ := json.MarshalIndent(&l, "", "\t")
	return string(b)
}

// DecodeEndUserLicense decodes a *EndUserLicense from a io.Reader
func DecodeEndUserLicense(r io.Reader) (*EndUserLicense, error) {
	var l EndUserLicense
	if err := json.NewDecoder(r).Decode(&l); err != nil {
		return nil, errors.Wrap(err, "failed to JSON decode")
	}
	return &l, nil
}

// GetEndUserLicense calls /my/v2/enduserlicenses
func (c *Client) GetEndUserLicense(ctx context.Context, license []byte) (*EndUserLicense, []byte, *http.Response, error) {
	serializedLicense := base64.StdEncoding.EncodeToString(license)
	reqBytes, err := json.Marshal(&struct {
		SerializedPublishingLicense string `json:"SerializedPublishingLicense"`
	}{
		SerializedPublishingLicense: serializedLicense,
	})
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to wrap license")
	}
	req, err := c.NewRequest(ctx, "POST", "/my/v2/enduserlicenses", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to create Request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, nil, resp, errors.Wrap(err, "failed to do Request")
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	r := io.TeeReader(resp.Body, &buf)
	l, err := DecodeEndUserLicense(r)
	if err != nil {
		return nil, buf.Bytes(), resp, errors.Wrap(err, "failed to decode EndUserLicense")
	}
	return l, buf.Bytes(), resp, nil
}
