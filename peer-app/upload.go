package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	// Kademlia DHT
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

// UploadUI manages the directory browsing + file selection + DHT providing
type UploadUI struct {
	list    widget.List
	files   []os.DirEntry
	dirPath string

	// Buttons
	backBtn    widget.Clickable
	fileBtns   []widget.Clickable
	selected   os.DirEntry
	confirmBtn widget.Clickable

	mu sync.RWMutex

	// Add references to your node + DHT so we can Provide the file
	node   host.Host
	kadDHT *dht.IpfsDHT
}

// LoadFiles loads the list of files and directories in the current dirPath
func (upload *UploadUI) LoadFiles() {
	upload.mu.Lock()
	defer upload.mu.Unlock()

	log.Println("Current Dir =", upload.dirPath)

	files, err := os.ReadDir(upload.dirPath)
	if err != nil {
		log.Println("Error reading directory:", err)
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

// func (upload *UploadUI) LoadFilesAgain(node host.Host, kadDHT *dht.IpfsDHT) {
// 	dirPath := "./p2-dir" // Ensure this matches the actual directory
// 	files, err := os.ReadDir(dirPath)
// 	if err != nil {
// 		log.Println("Error reading directory:", err)
// 		return
// 	}

// 	for _, file := range files {
// 		if !file.IsDir() {
// 			filePath := file.Name()
// 			log.Println("Announcing file:", filePath)
// 			announceFile(node, kadDHT, filePath)
// 		}
// 	}
// }

// UploadLayout is the main layout method for the Upload tab
func (upload *UploadUI) UploadLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Title label
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Left: 20, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Label(th, th.TextSize, "File Upload")
						label.Alignment = text.Start
						return label.Layout(gtx)
					})
				}),
				// "Back" button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Left: 20, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if upload.dirPath != "/" {
							btn := material.Button(th, &upload.backBtn, "Back")
							if upload.backBtn.Clicked(gtx) {
								upload.NavigateUp()
							}
							return btn.Layout(gtx)
						}
						return layout.Dimensions{}
					})
				}),
			)
		}),
		// File List
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return upload.list.List.Layout(gtx, len(upload.files), func(gtx layout.Context, i int) layout.Dimensions {
				if i >= len(upload.files) {
					return layout.Dimensions{}
				}
				return upload.FileLayout(gtx, th, i)
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
							upload.UploadFile()
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
	)
}

// FileLayout lays out a single file/folder row in the list
func (upload *UploadUI) FileLayout(gtx layout.Context, th *material.Theme, i int) layout.Dimensions {
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
					fileBtn.Color = th.Palette.ContrastFg
					fileBtn.Background = color.NRGBA{R: 60, G: 179, B: 113, A: 255}
				} else {
					fileBtn.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
					fileBtn.Background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				}

				if btn.Clicked(gtx) {
					if file.IsDir() {
						upload.NavigateTo(file.Name())
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
func (upload *UploadUI) NavigateTo(dir string) {
	upload.dirPath = filepath.Join(upload.dirPath, dir)
	upload.LoadFiles()
}

// NavigateUp moves up one directory
func (upload *UploadUI) NavigateUp() {
	upload.dirPath = filepath.Dir(upload.dirPath)
	upload.LoadFiles()
}

// UploadFile performs the final "upload" step:
// 1) Gathers file info (hash, size).
// 2) (Still) POSTs the manifest to your load balancer (if you want).
// 3) Provides the file to the Kademlia DHT so other peers can find it.
func (upload *UploadUI) UploadFile() {
	if upload.selected == nil {
		PopupMessage("No file is selected!")
		return
	}

	fileName := upload.selected.Name()
	filePath := filepath.Join(upload.dirPath, fileName)
	log.Println("File Path =", filePath)

	// 1) Get file size
	fileSize, err := GetFileSize(filePath)
	if err != nil {
		log.Println("Error when getting file size:", err)
		return
	}
	log.Println("File Size =", strconv.FormatInt(fileSize, 10), "bytes")

	// 2) Get file hash
	fileHash, err := GetFileHash(filePath)
	if err != nil {
		log.Println("Error when getting file hash:", err)
		return
	}
	log.Println("File Hash =", fileHash)

	// 3) Build the manifest
	manifest := types.Manifest{
		Name: fileName,
		Hash: fileHash,
		// Optionally set Size: fileSize if your model supports it
	}

	// 4) Marshal to JSON and POST to your load balancer (if you still want that)
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		return
	}

	postReq := "/api/" + DBManagerVer + "/records"
	log.Println("Send POST " + postReq + " to " + LoadBalancerAdr)

	req, err := http.NewRequest("POST", LoadBalancerAdr+postReq, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating POST request:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	respSer, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return
	}
	log.Println("Resp Status:", resp.Status)
	log.Println("Resp Body:", string(respSer))

	// ---------------------------------------------------------
	// 5) Provide the file to the Kademlia DHT
	c, err := cidFromString(fileHash)
	if err != nil {
		log.Println("Error creating CID from fileHash:", err)
		return
	}
	ctx := context.Background()
	if err := upload.kadDHT.Provide(ctx, c, true); err != nil {
		log.Println("Error providing file to DHT:", err)
		return
	}
	log.Println("Provided to DHT: ", filePath, "with CID:", c.String())

	PopupMessage("Uploaded & Provided: " + fileName)
}

// Helper: get file size
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}
	return fileInfo.Size(), err
}

// Helper: get file SHA256
func GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func announceFile(node host.Host, kadDHT *dht.IpfsDHT, filePath string) {
	// Convert filename to hash (CID-like)
	fileHashCID, err := cidFromString(filePath)
	if err != nil {
		log.Println("Error generating CID for file:", err)
		return
	}

	// Store the file hash in the DHT
	ctx := context.Background()
	err = kadDHT.Provide(ctx, fileHashCID, true)
	if err != nil {
		log.Println("Error announcing file to DHT:", err)
		return
	}

	log.Println("File announced in DHT:", filePath, "->", fileHashCID.String())
}
