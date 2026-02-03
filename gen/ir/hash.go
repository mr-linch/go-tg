package ir

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// hashableAPI contains only the fields that should be included in the hash computation.
type hashableAPI struct {
	Types   []Type   `json:"types,omitempty"`
	Methods []Method `json:"methods,omitempty"`
	Enums   []Enum   `json:"enums,omitempty"`
}

// ComputeHash computes a SHA256 hash of the API structure.
// Only Types, Methods, and Enums are included in the hash computation.
// Version, ReleaseDate, and Hash fields are excluded.
// Returns the first 12 characters of the hex-encoded hash.
func (api *API) ComputeHash() string {
	h := hashableAPI{
		Types:   api.Types,
		Methods: api.Methods,
		Enums:   api.Enums,
	}

	data, err := json.Marshal(h)
	if err != nil {
		return ""
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:12]
}
