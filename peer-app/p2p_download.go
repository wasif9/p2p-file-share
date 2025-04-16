package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

func DownloadManifest(fileName, fileCID, filePath string) {
	var providers []peer.ID
	var err error
	if providers, err = dhtLookup(fileCID); err != nil {
		PopupMessage("Fail to download file due to " + err.Error())
	}
	for {

		for _, provider := range providers {
			if err := requestFile(provider, fileName+".json", filePath, fileCID, "manifest", 0); err != nil {
				log.Println(err)
				continue
			}
			return
		}
	}
}

func DownloadFile(fileName, filePath string, downloadFile *Download) {
	// Load manifest
	file, err := os.Open(filepath.Join(filePath, ".p2p", fileName+".json"))
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	var manifest types.ManifestData
	if err := json.Unmarshal(data, &manifest); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Prepare for parallel download
	totalChunks := len(manifest.Chunks)
	chunkData := make([][]byte, totalChunks)
	errChan := make(chan error, totalChunks)
	var wg sync.WaitGroup

	providers, err := dhtLookup(manifest.FileCID)

	if err != nil {
		errChan <- fmt.Errorf("File lookup failed: %w", err)
		return
	}

	for index, chunkCID := range manifest.Chunks {
		wg.Add(1)

		// Round-robin select a peer for this chunk
		provider := providers[index%len(providers)]

		go func(idx int, cidStr string, provider peer.ID) {
			defer wg.Done()

			buffer, err := requestChunk(provider, fileName, idx)
			if err != nil {
				log.Printf("Chunk %d failed from provider %s: %v", idx, provider, err)
				errChan <- fmt.Errorf("chunk %d failed from provider %s", idx, provider)
				return
			}

			computedCID, err := cidFromBytes(buffer)
			if err != nil || computedCID.String() != cidStr {
				log.Printf("Chunk %d CID mismatch from provider %s", idx, provider)
				errChan <- fmt.Errorf("chunk %d integrity check failed", idx)
				return
			}

			chunkData[idx] = buffer
			downloadFile.UpdateProgress(totalChunks)
		}(index, chunkCID, provider)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		for err := range errChan {
			log.Println("Error during chunk download:", err)
		}
		PopupMessage("Some chunks failed to download")
		return
	}

	// Reconstruct file
	savePath := filepath.Join(filePath, fileName)
	f, err := os.Create(savePath)
	if err != nil {
		log.Println("Failed to create file for reconstruction:", err)
		return
	}
	defer f.Close()
	for _, chunk := range chunkData {
		f.Write(chunk)
	}

	// Verify file integrity
	computedCID, err := cidFromFile(savePath)
	if err != nil {
		log.Println("Error computing CID for reconstructed file:", err)
		return
	}
	if computedCID.String() == manifest.FileCID {
		log.Println("✅ File reconstruction integrity verified")

		// Automatically announce that this peer now serves the file and manifest
		ctx := context.Background()
		log.Println("📢 Announcing file availability to DHT...")

		// Announce file CID
		if err := KadDHT.Provide(ctx, computedCID, true); err != nil {
			log.Println("Error providing file CID:", err)
		}

		// Compute manifest CID and announce it too
		manifestPath := filepath.Join(filePath, ".p2p", fileName+".json")
		manifestCID, err := cidFromFile(manifestPath)
		if err != nil {
			log.Println("Error computing CID for manifest:", err)
		} else {
			if err := KadDHT.Provide(ctx, manifestCID, true); err != nil {
				log.Println("Error providing manifest CID:", err)
			}
		}

	} else {
		log.Println("❌ File CID mismatch after reconstruction")
		log.Println("Expected:", manifest.FileCID)
		log.Println("Got:", computedCID.String())
	}
}

func dhtLookup(fileCID string) ([]peer.ID, error) {
	ctx := context.Background()

	// Decode CID
	c, err := cid.Decode(fileCID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert CID: %v", err)
	}

	log.Print("FINDING PROVIDERS...")
	providerChan := KadDHT.FindProvidersAsync(ctx, c, 10)

	// Temporary slice to hold providers before filtering
	var aliveProviders []peer.ID
	providerSet := make(map[peer.ID]bool)

	for p := range providerChan {
		if p.ID == Node.ID() {
			continue // skip self
		}

		if providerSet[p.ID] {
			continue // skip duplicate
		}
		providerSet[p.ID] = true

		log.Println("Discovered provider:", p.ID.String())

		// Attempt to connect
		if err := Node.Connect(ctx, p); err != nil {
			log.Printf("Skipping %s (unreachable)", p.ID)
			continue
		}

		log.Printf("✅ Peer %s is reachable", p.ID.ShortString())
		aliveProviders = append(aliveProviders, p.ID)
	}

	if len(aliveProviders) == 0 {
		return nil, fmt.Errorf("no reachable providers found")
	}

	return aliveProviders, nil
}

func requestChunk(providerID peer.ID, fileName string, chunkID int) ([]byte, error) {
	ctx := context.Background()
	s, err := Node.NewStream(ctx, providerID, Protocol)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	request := types.DownloadRequest{
		FileName:   fileName,
		ChunkIndex: chunkID,
		Type:       "file",
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	if _, err := s.Write(jsonData); err != nil {
		return nil, err
	}

	return io.ReadAll(s)
}

func requestFile(providerID peer.ID, fileName, dirPath, expectedCID, downloadType string, chunkID int) error {
	// Open a new stream to the provider
	ctx := context.Background()

	// Log which peer we're contacting
	log.Println("Attempting to request file from provider:", providerID.String())

	s, err := Node.NewStream(ctx, providerID, Protocol)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			// Closed by the other side (provider)
			log.Println("Peer app error when close request file request ", err)
		}
	}()

	// Send the file request
	request := types.DownloadRequest{
		FileName:   fileName,
		ChunkIndex: chunkID,
		Type:       downloadType,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		PopupMessage("Cannot download file due to JSON error")
		return err
	}
	log.Printf("Requesting file: %v\n", request)
	if _, err := s.Write(jsonData); err != nil {
		return err
	}

	// Read the response
	data, err := io.ReadAll(s)
	if err != nil {
		return err
	}

	// Check if the response indicates an error
	response := string(data)
	if strings.HasPrefix(response, "Error:") || strings.HasPrefix(response, "File not found") {
		log.Printf("Failed to get file %s from the provider %s, %s\n", fileName, providerID.String(), response)
		return fmt.Errorf("file Not Found from the provider")
	}

	// Construct the correct save path in the peer's directory
	savePath := filepath.Join(dirPath, fileName)

	// Try to create a parent folder
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Fatalf("Failed to create .p2p folder in %v dir: %v", savePath, err)
	}
	// Write the file data to correct directory
	err = os.WriteFile(savePath, data, 0644)
	if err != nil {
		log.Fatalf("Failed to create file %v: %v", savePath, err)
	}

	log.Printf("Received and saved file as %s\n", fileName)

	// Integrity Check: Compare computed CID with expected CID
	computedCID, err := cidFromFile(savePath)
	if err != nil {
		log.Println("Error computing CID for downloaded file:", err)
		return err
	}
	if computedCID.String() == expectedCID {
		log.Println("✅ Integrity check passed: File matches expected CID.")
	} else {
		log.Println("❌ Integrity check failed: File does NOT match expected CID.")
		log.Printf("Expected CID: %s\n", expectedCID)
		log.Printf("Computed CID: %s\n", computedCID.String())
		return fmt.Errorf("integrity check failed")
	}

	return nil
}
