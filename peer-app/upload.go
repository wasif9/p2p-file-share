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

	// Kademlia DHT
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"

	types "github.com/wasif9/p2p-file-share/pkg/models"
)

// UploadUI manages the directory browsing + file selection + DHT providing
type UploadUI struct {
	list       widget.List
	files      []os.DirEntry
	dirPath    string
	errorMsg   string
	refreshBtn widget.Clickable
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
							upload.NavigateUp()
						}
						return btn.Layout(gtx)
					})
				}),
				// Current directory
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Bottom: 10, Left: 20, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						pwd, err := os.Getwd()
						if err != nil {
							log.Println("Error when getting current directory", err)
							pwd = ""
						}
						pwd = filepath.Join(pwd, upload.dirPath)
						label := material.Label(th, th.TextSize, pwd)
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
	if upload.dirPath != dataDir {
		upload.dirPath = filepath.Dir(upload.dirPath)
		upload.LoadFiles()
	}
}

// UploadFile performs the final upload step correctly:
// 1) Generates CID from file data.
// 2) Provides the CID to the Kademlia DHT, ensuring peer discoverability.
// 3) If successful, posts the manifest (with CID) to the load balancer/DB.
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
		log.Println("Error getting file size:", err)
		PopupMessage("Canoot upload file \ndue to file corruption")
		return
	}
	log.Println("File Size =", strconv.FormatInt(fileSize, 10), "bytes")

	// 2) Generate CID from file data
	cid, err := cidFromFile(filePath)
	if err != nil {
		log.Println("Error generating CID from file:", err)
		PopupMessage("Canoot upload file \ndue to file corruption")
		return
	}
	log.Println("CID =", cid.String())

	// 3) Provide the file CID to the Kademlia DHT
	ctx := context.Background()
	if err := upload.kadDHT.Provide(ctx, cid, true); err != nil {
		log.Println("Error providing CID to DHT:", err)
		PopupMessage("Canoot upload file \ndue to DHT error")
		return
	}
	log.Println("Successfully provided CID to DHT:", cid.String())

	// 4) Build the manifest with CID
	manifest := types.Manifest{
		Name: fileName,
		Hash: cid.String(), // CID instead of raw file hash
		Size: fileSize,
	}

	// 5) Marshal manifest to JSON
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		PopupMessage("Canoot upload file \ndue to JSON type error")
		return
	}

	// 6) POST manifest JSON to load balancer
	postReq := "/api/" + DBManagerVer + "/manifests"
	log.Println("Sending POST", postReq, "to", reverseProxyAddr)

	// Set timeout for whole POST request
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "http://"+reverseProxyAddr+postReq, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating POST request:", err)
		PopupMessage("Canoot upload file \ndue to POST request failure")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PopupMessage("Canoot upload file \ndue to busy servers")
		} else {
			PopupMessage("Canoot upload file \ndue to proxy error")
		}
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	respSer, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		PopupMessage("Cannot upload file \ndue to incorrect response from servers")
		return
	}
	log.Println("Resp Status:", resp.Status)
	log.Println("Resp Body:", string(respSer))

	if resp.Status == "201 Created" {
		PopupMessage("Uploaded & Provided: " + fileName)
	} else {
		PopupMessage("Cannot upload file \ndue to server error")
	}
}

// Helper: get file size
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}
	return fileInfo.Size(), err
}
