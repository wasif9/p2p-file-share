package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

// UploadUI manages the directory browsing + file selection + DHT providing
type UploadUI struct {
	list       widget.List
	files      []os.DirEntry
	dirPath    string
	refreshBtn widget.Clickable
	backBtn    widget.Clickable
	fileBtns   []widget.Clickable
	selected   os.DirEntry
	confirmBtn widget.Clickable

	mu sync.RWMutex
}

// LoadFiles loads the list of files and directories in the current dirPath
func (upload *UploadUI) LoadFiles() {
	upload.mu.Lock()
	defer upload.mu.Unlock()

	log.Println("Current Dir =", upload.dirPath)

	files, err := os.ReadDir(upload.dirPath)
	if err != nil {
		log.Println("Error reading directory:", err)
		PopupMessage("Cannot read current directory!")
		upload.files = nil
		return
	}
	// Sort: directories first, then files
	sort.Slice(files, func(i, j int) bool {
		iIsDir := files[i].IsDir()
		jIsDir := files[j].IsDir()
		if iIsDir == jIsDir {
			return files[i].Name() < files[j].Name()
		}
		return iIsDir
	})
	upload.files = files
	upload.fileBtns = make([]widget.Clickable, len(files))
	upload.selected = nil
}

// UploadLayout is the main layout method for the Upload tab
func (upload *UploadUI) UploadLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
		// Title label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Padding
						inset := layout.Inset{Top: 10, Bottom: 10}
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							text := material.Body1(th, "Upload Files")
							text.TextSize = unit.Sp(20)
							return text.Layout(gtx)
						})
					})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// "Back" button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Left: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &upload.backBtn, "🢨")
						btn.TextSize = 25
						if upload.backBtn.Clicked(gtx) {
							upload.navigateUp()
						}
						return btn.Layout(gtx)
					})
				}),
				// Current directory
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Left: 20, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Label(th, th.TextSize, upload.dirPath)
						label.Alignment = text.Start
						return label.Layout(gtx)
					})
				}),
				// "Refresh" button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &upload.refreshBtn, "⟳")
						btn.TextSize = 25
						if upload.refreshBtn.Clicked(gtx) {
							upload.LoadFiles()
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
		// File List
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return upload.list.Layout(gtx, len(upload.files), func(gtx layout.Context, i int) layout.Dimensions {
				if i >= len(upload.files) {
					return layout.Dimensions{}
				}
				return upload.fileLayout(gtx, th, i)
			})
		}),
		// Divider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Top: 10, Bottom: 10}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				width := gtx.Constraints.Max.X
				thickness := gtx.Dp(3)
				rect := clip.Rect{Max: image.Point{X: width, Y: thickness}}.Op()
				paint.FillShape(gtx.Ops, color.NRGBA{R: 210, G: 210, B: 210, A: 255}, rect)
				return layout.Dimensions{Size: image.Point{X: width, Y: thickness}}
			})
		}),
		// Confirm/Upload button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.End}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Bottom: 10, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &upload.confirmBtn, "Upload")
						if upload.confirmBtn.Clicked(gtx) {
							upload.uploadFile()
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
	)
}

// FileLayout lays out a single file/folder row in the list
func (upload *UploadUI) fileLayout(gtx layout.Context, th *material.Theme, i int) layout.Dimensions {
	file := upload.files[i]
	btn := &upload.fileBtns[i]
	label := file.Name()

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Top: 8, Bottom: 8, Left: 20, Right: 20}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				fileBtn := material.Button(th, btn, label)

				if file.IsDir() {
					// Blue for folders
					fileBtn.Background = color.NRGBA{R: 100, G: 150, B: 255, A: 255}
				} else if file == upload.selected {
					// Green for selected files
					fileBtn.Color = th.ContrastFg
					fileBtn.Background = color.NRGBA{R: 60, G: 179, B: 113, A: 255}
				} else {
					fileBtn.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
					fileBtn.Background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				}

				if btn.Clicked(gtx) {
					if file.IsDir() {
						upload.navigateTo(file.Name())
					} else {
						upload.selected = file
						log.Println("Selected File:", file.Name())
					}
				}
				return fileBtn.Layout(gtx)
			})
		}),
	)
}

// NavigateTo goes deeper into a subdirectory
func (upload *UploadUI) navigateTo(dir string) {
	upload.dirPath = filepath.Join(upload.dirPath, dir)
	upload.LoadFiles()
}

// NavigateUp moves up one directory
func (upload *UploadUI) navigateUp() {
	if upload.dirPath != DataDir {
		upload.dirPath = filepath.Dir(upload.dirPath)
		upload.LoadFiles()
	}
}

// UploadFile performs the full upload process:
// 1. Splits the file into chunks and stores them.
// 2. Generates CIDs for each chunk and the full file.
// 3. Creates a manifest.json file mapping chunk index to CID.
// 4. Registers the full file CID and manifest CID to the DHT.
// 5. Sends the manifest CID to the reverse proxy server.
// 6. Deletes chunk files locally to conserve space.
func (upload *UploadUI) uploadFile() {
	if upload.selected == nil {
		PopupMessage("No file is selected!")
		return
	}

	fileName := upload.selected.Name()
	filePath := filepath.Join(upload.dirPath, fileName)
	log.Println("File Path =", filePath)

	// 1. Get file size
	fileSize, err := getFileSize(filePath)
	if err != nil {
		log.Println("Error getting file size:", err)
		PopupMessage("Cannot upload file due to file corruption")
		return
	}
	log.Println("File Size =", strconv.FormatInt(fileSize, 10), "bytes")

	// 2. Chunk the file, generate CIDs for each chunk and the full file
	chunkMap, fullFileCID, err := chunkAndStoreChunks(filePath)
	if err != nil {
		log.Println("Error during chunking:", err)
		PopupMessage("Failed to chunk file")
		return
	}

	// 3. Create and save manifest.json to disk using internal struct
	manifest := types.ManifestData{
		FileName: fileName,
		FileCID:  fullFileCID.String(),
		Chunks:   chunkMap,
	}
	manifestPath, err := writeManifestToDisk(manifest)
	if err != nil {
		log.Println("Error writing manifest file:", err)
		PopupMessage("Cannot upload file due to manifest error")
		return
	}

	// 4. Get CID for manifest file and register CIDs to DHT
	ctx := context.Background()
	manifestCID, err := cidFromFile(manifestPath)
	if err != nil {
		log.Println("Error generating CID for manifest file:", err)
		PopupMessage("Cannot upload file due to manifest CID failure")
		return
	}
	log.Println("Registering File CID and Manifest CID to DHT")
	KadDHT.Provide(ctx, fullFileCID, true)
	KadDHT.Provide(ctx, manifestCID, true)

	// 5. Upload manifest CID to reverse proxy
	manifestToSend := types.Manifest{
		Name: fileName,
		Hash: manifestCID.String(),
		Size: fileSize,
	}
	jsonData, err := json.Marshal(manifestToSend)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		PopupMessage("Cannot upload file due to JSON error")
		return
	}
	postReq := "/api/" + DBManagerVer + "/manifests"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "http://"+ReverseProxyAddr+postReq, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating POST request:", err)
		PopupMessage("Cannot upload file due to POST request failure")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PopupMessage("Cannot upload file due to busy servers")
		} else {
			PopupMessage("Cannot upload file due to proxy error")
		}
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	respSer, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		PopupMessage("Cannot upload file due to incorrect server response")
		return
	}
	log.Println("Resp Status:", resp.Status)
	log.Println("Resp Body:", string(respSer))

	// 6. Cleanup chunk directory
	os.RemoveAll(filepath.Join(DataDir, ".p2p", fileName+"_chunks"))

	if resp.StatusCode == http.StatusCreated {
		PopupMessage("Uploaded & Provided: " + fileName)
	} else {
		PopupMessage("Cannot upload file due to server error: \n" + string(respSer))
	}
}

// Helper to write the manifest.json file and return the full path
func writeManifestToDisk(manifest types.ManifestData) (string, error) {
	manifestDir := filepath.Join(DataDir, ".p2p")
	os.MkdirAll(manifestDir, os.ModePerm)
	manifestPath := filepath.Join(manifestDir, manifest.FileName+".json")
	f, err := os.Create(manifestPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(manifest); err != nil {
		return "", err
	}
	return manifestPath, nil
}

// Helper: get file size
func getFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}
	return fileInfo.Size(), err
}
