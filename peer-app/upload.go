package main

import (
	"bytes"
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
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

type UploadUI struct {
	list       widget.List
	files      []os.DirEntry
	dirPath    string
	backBtn    widget.Clickable
	fileBtns   []widget.Clickable
	selected   os.DirEntry
	confirmBtn widget.Clickable
	mu         sync.RWMutex
}

// LoadFiles loads the list of files and directories
func (upload *UploadUI) LoadFiles() {
	// Prevent race condition when rendering
	upload.mu.Lock()
	defer upload.mu.Unlock()

	log.Println("Current Dir = " + upload.dirPath)

	// Read files from current dir
	files, err := os.ReadDir(upload.dirPath)
	if err != nil {
		log.Println("Error reading directory:", err)
		return
	}

	// Sort files by their types
	sort.Slice(files, func(i, j int) bool {
		iIsDir := files[i].IsDir()
		jIsDir := files[j].IsDir()

		// Both are dir or files
		if iIsDir == jIsDir {
			return files[i].Name() < files[j].Name()
		}

		return iIsDir
	})

	// Update UI state
	upload.files = files
	upload.fileBtns = make([]widget.Clickable, len(files))
	upload.selected = nil
}

// Upload UI layout
func (upload *UploadUI) UploadLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Title
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{
						Top:    10,
						Bottom: 10,
						Left:   20,
						Right:  20,
					}

					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						label := material.Label(th, th.TextSize, "File Upload")
						label.Alignment = text.Start
						return label.Layout(gtx)
					})
				}),

				// Up dir button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{
						Top:    10,
						Bottom: 10,
						Left:   20,
						Right:  20,
					}

					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if upload.dirPath != "/" {
							btn := material.Button(th, &upload.backBtn, "Back")

							// Move to upper dir when click button
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
				// As i will be refreshed after files updated
				if i >= len(upload.files) {
					return layout.Dimensions{}
				}
				return upload.FileLayout(gtx, th, i)
			})
		}),

		// Divider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Padding
			inset := layout.Inset{
				Top:    10,
				Bottom: 10,
			}

			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				width := gtx.Constraints.Max.X
				thickness := gtx.Dp(3)

				rect := clip.Rect{
					Min: image.Point{0, 0},
					Max: image.Point{width, thickness},
				}.Op()

				// Apply the clip and fill with color
				paint.FillShape(gtx.Ops, color.NRGBA{R: 210, G: 210, B: 210, A: 255}, rect)

				return layout.Dimensions{
					Size: image.Point{X: width, Y: thickness},
				}
			})
		}),

		// Confirm button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.End}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{
						Bottom: 10,
						Right:  20,
					}

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

// Single file layout
func (upload *UploadUI) FileLayout(gtx layout.Context, th *material.Theme, i int) layout.Dimensions {
	file := upload.files[i]
	btn := &upload.fileBtns[i]
	label := file.Name()

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			// Padding
			inset := layout.Inset{
				Top:    8,
				Bottom: 8,
				Left:   20,
				Right:  20,
			}

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
					// White for unselected files
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

// Navigate to a directory
func (upload *UploadUI) NavigateTo(dir string) {
	upload.dirPath = filepath.Join(upload.dirPath, dir)
	upload.LoadFiles()
}

// Move up one directory
func (upload *UploadUI) NavigateUp() {
	upload.dirPath = filepath.Dir(upload.dirPath)
	upload.LoadFiles()
}

// Get file size
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}
	return fileInfo.Size(), err
}

// Get file SHA256 hash
func GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashSum := hash.Sum(nil) // Hash sum to bytes

	return hex.EncodeToString(hashSum), nil
}

// Upload file
func (upload *UploadUI) UploadFile() {
	if upload.selected == nil {
		PopupMessage("No file is selected!")
		return
	}

	fileName := upload.selected.Name()

	// Get file path
	filePath := filepath.Join(upload.dirPath, fileName)
	log.Println("File Path = " + filePath)

	// Get file size
	fileSize, err := GetFileSize(filePath)
	if err != nil {
		log.Println("Error when getting file size: ", err)
		return
	}
	log.Println("File Size = " + strconv.FormatInt(fileSize, 10) + " bytes")

	// Get file hash
	fileHash, err := GetFileHash(filePath)
	if err != nil {
		log.Println("Error when getting file hash: ", err)
		return
	}
	log.Println("File Hash = " + fileHash)

	// Create manifest file
	manifest := types.Manifest{
		Name: fileName,
		Hash: fileHash,
		// ! Filesize included in manifest file
		// Size: fileSize,
	}

	// Marshal to JSON file
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		log.Println("Error when encoding JSON: ", err)
		return
	}

	// Create POST request
	postReq := "/api/" + DBManagerVer + "/records/" + fileName
	log.Println("Send POST " + postReq + " to " + LoadBalancerAdr)

	req, err := http.NewRequest("POST", LoadBalancerAdr+postReq, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error when creating POST request: ", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send POST request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error when sending POST request: ", err)
		return
	}
	defer resp.Body.Close()

	// Response from the server
	respSer, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error when reading response after POST req, ", err)
		return
	}

	// Print the response from the server
	log.Println("Resp Status: ", resp.Status)
	log.Println("Resp Body: ", string(respSer))
}
