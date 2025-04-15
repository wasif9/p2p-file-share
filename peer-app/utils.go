package main

import (
	"crypto/sha256"
	"io"
	"log"
	"os"
	"path/filepath"

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
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal("Error when close the file used for CID", err)
		}
	}()

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
func cidFromBytes(data []byte) (cid.Cid, error) {
	h := sha256.Sum256(data)
	mhBytes, err := mh.Encode(h[:], mh.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}
	return cid.NewCidV1(cid.Raw, mhBytes), nil
}

func chunkAndStoreChunks(filePath string) (map[int]string, cid.Cid, error) {
	const chunkSize = 512 * 1024
	file, err := os.Open(filePath)
	if err != nil {
		return nil, cid.Cid{}, err
	}
	defer file.Close()

	chunkMap := make(map[int]string)
	chunkFolder := filepath.Join(".p2p", filepath.Base(filePath)+"_chunks")
	os.MkdirAll(chunkFolder, os.ModePerm)

	var fullData []byte
	buf := make([]byte, chunkSize)
	chunkIndex := 0

	for {
		n, err := file.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			fullData = append(fullData, chunk...)
			chunkCID, err := cidFromBytes(chunk)
			if err != nil {
				return nil, cid.Cid{}, err
			}

			chunkFilePath := filepath.Join(chunkFolder, chunkCID.String())
			if err := os.WriteFile(chunkFilePath, chunk, 0644); err != nil {
				return nil, cid.Cid{}, err
			}
			chunkMap[chunkIndex] = chunkCID.String()
			chunkIndex++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, cid.Cid{}, err
		}
	}

	fullHash := sha256.Sum256(fullData)
	mhBytes, _ := mh.Encode(fullHash[:], mh.SHA2_256)
	fullFileCID := cid.NewCidV1(cid.Raw, mhBytes)
	return chunkMap, fullFileCID, nil
}
