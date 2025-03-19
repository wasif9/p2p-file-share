package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// cidFromString is a helper to produce a "fake" CID from any string (like your manifest.Hash).
// In a real system, you'd have an actual content-based CID from the data itself.
func cidFromString(str string) (cid.Cid, error) {
	h := sha256.Sum256([]byte(str))
	mhBytes, err := mh.Encode(h[:], mh.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}
	return cid.NewCidV1(cid.Raw, mhBytes), nil
}
func cidFromFile(filepath string) (cid.Cid, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return cid.Cid{}, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return cid.Cid{}, err
	}

	h := sha256.Sum256(data)
	mhBytes, err := mh.Encode(h[:], mh.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}

	return cid.NewCidV1(cid.Raw, mhBytes), nil
}
func cidFromHash(fileHash string) (cid.Cid, error) {
	hashBytes, err := hex.DecodeString(fileHash)
	if err != nil {
		return cid.Cid{}, err
	}
	mhBytes, err := mh.Encode(hashBytes, mh.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}
	return cid.NewCidV1(cid.Raw, mhBytes), nil
}
