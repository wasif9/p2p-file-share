package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dustin/go-humanize"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

type DownloadUI struct {
	node   host.Host
	kadDHT *dht.IpfsDHT // <-- Add this to do real DHT lookups

	searchInput    widget.Editor
	searchButton   widget.Clickable
	results        []types.Manifest
	selectedResult types.Manifest
	resultButtons  []widget.Clickable
	list           widget.List
	dirPath        string // The directory where downloaded files should be saved
	downloadButton widget.Clickable
	loading        bool
}

func (ui *DownloadUI) DownloadLayout(gtx layout.Context, th *material.Theme, prgUI *ProgressUI) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Title label
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Padding
						inset := layout.Inset{Top: 10, Bottom: 10}
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							text := material.Body1(th, "P2P File Lookup")
							text.TextSize = unit.Sp(20)
							return text.Layout(gtx)
						})
					})
				}),
			)
		}),
		// Upper part for text search
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Text bar
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{Top: 10, Right: 20, Left: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ui.searchInput.SingleLine = true
						ui.searchInput.Submit = true
						textInput := material.Editor(th, &ui.searchInput, "Search for files ...")

						// Detect Enter pressed
						if e, ok := ui.searchInput.Update(gtx); ok {
							if _, isSubmit := e.(widget.SubmitEvent); isSubmit {
								go ui.PerformSearch()
							}
						}

						return textInput.Layout(gtx)
					})
				}),
				//Search Button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{Top: 10, Right: 20, Left: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &ui.searchButton, "Search")
						if ui.searchButton.Clicked(gtx) {
							go ui.PerformSearch()
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
		// Divider
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Top: 3, Bottom: 3}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				width := gtx.Constraints.Max.X
				thickness := gtx.Dp(3)
				rect := clip.Rect{Max: image.Point{X: width, Y: thickness}}.Op()
				paint.FillShape(gtx.Ops, color.NRGBA{R: 210, G: 210, B: 210, A: 255}, rect)
				return layout.Dimensions{Size: image.Point{X: width, Y: thickness}}
			})
		}),
		// Lower part for search results
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.loading {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							// Padding
							inset := layout.Inset{Top: 10, Bottom: 10}
							return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								text := material.Body1(th, "⏳...")
								text.TextSize = unit.Sp(20)
								return text.Layout(gtx)
							})
						})
					}),
				)
			} else {
				return ui.LayoutResults(gtx, th, prgUI)
			}
		}),
	)
}

func (ui *DownloadUI) LayoutResults(gtx layout.Context, th *material.Theme, prgUI *ProgressUI) layout.Dimensions {
	// When no result
	if len(ui.results) == 0 {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							text := material.Body1(th, "No Result")
							text.TextSize = unit.Sp(20)
							return text.Layout(gtx)
						})
					}),
				)
			}),
		)
	}

	// Results
	return layout.Inset{Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Display all result in buttons
		return ui.list.List.Layout(gtx, len(ui.results), func(gtx layout.Context, i int) layout.Dimensions {
			btn := &ui.resultButtons[i]
			isSelected := ui.selectedResult == ui.results[i]

			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				// The result button
				layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{Top: 5, Bottom: 5, Left: 20, Right: 20}
					// Apply the padding and layout the button
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						fileSize := uint64(ui.results[i].Size)
						display_str := fmt.Sprintf("%s | %s", ui.results[i].Name, humanize.Bytes(fileSize))
						button := material.Button(th, btn, display_str)

						// Different style for selected items
						if isSelected {
							// Create a highlighted button
							button.Background = th.Palette.ContrastBg
							button.Color = th.Palette.ContrastFg
						} else {
							// Regular button
							button.Background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
							button.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
						}

						// Register click event
						if btn.Clicked(gtx) {
							ui.selectedResult = ui.results[i]
						}

						return button.Layout(gtx)
					})
				}),

				// The download button (only shown for selected item)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if isSelected {
						// Padding
						inset := layout.Inset{Right: 20}
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							// Download buttton property
							btn := material.Button(th, &ui.downloadButton, "Download")
							btn.Background = color.NRGBA{R: 39, G: 215, B: 45, A: 255}

							// When download is clicked
							if ui.downloadButton.Clicked(gtx) {
								log.Println("Selected result:", ui.selectedResult)
								ui.Download(prgUI)
							}
							return btn.Layout(gtx)
						})
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}

func (ui *DownloadUI) PerformSearch() {
	ui.loading = true
	query := strings.TrimSpace(ui.searchInput.Text())

	// Make GET request to the load balancer server
	getReq := "/api/" + DBManagerVer + "/manifests?prefix=" + query
	log.Println("Send GET " + getReq + " to " + reverseProxyAddr)

	// Send GET requet
	reverseProxy := &http.Client{Timeout: time.Second * 2}
	resp, err := reverseProxy.Get("http://" + reverseProxyAddr + getReq)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			PopupMessage("Cannot search files \ndue to busy servers")
		} else {
			PopupMessage("Cannot search files \ndue to proxy error")
		}
		log.Println("Error when sending search query request", err)
		ui.results = nil
		return
	}
	defer resp.Body.Close()

	// Check for successful response status
	if resp.StatusCode != http.StatusOK {
		PopupMessage("Cannot get data from servers! Try again")
		log.Println("Error receiving non-OK response", resp.Status)
		ui.results = nil
		return
	}

	// Read the response body
	respSer, err := io.ReadAll(resp.Body)
	if err != nil {
		PopupMessage("Error response from database")
		log.Println("Error reading response body:", err)
		ui.results = nil
		return
	}

	// Decode the JSON response
	var manifests []types.Manifest
	if err := json.Unmarshal(respSer, &manifests); err != nil {
		PopupMessage("Error response from database")
		log.Println("Decode Error:", err)
		ui.results = nil
		return
	}

	// Change the GUI search result
	ui.results = manifests
	// Make buttons to select the result
	ui.resultButtons = make([]widget.Clickable, len(ui.results))
	ui.loading = false
	log.Println("-------------------------------------")
}

func (ui *DownloadUI) Download(prgUI *ProgressUI) {
	go func() {
		fileName := ui.selectedResult.Name

		// Check file if is downloading
		if prgUI.IsInDownload(fileName, ui.selectedResult.Hash) {
			PopupMessage(fileName + " is downloading!")
			return
		}

		// Make GET request to the load balancer server
		getReq := "/api/" + DBManagerVer + "/manifests/" + fileName
		log.Println("Send GET " + getReq + " to " + reverseProxyAddr)

		// Send GET requet
		reverseProxy := &http.Client{Timeout: time.Second * 2}
		resp, err := reverseProxy.Get("http://" + reverseProxyAddr + getReq)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				PopupMessage("Cannot download file \ndue to busy servers")
			} else {
				PopupMessage("Cannot download file \ndue to proxy error")
			}
			log.Println("Error when sending single file query", err)
			return
		}
		defer resp.Body.Close()

		// Decode the JSON response
		var manifest types.Manifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			PopupMessage("Error response from servers")
			log.Println("Decode Error:", err)
			return
		}

		// Print received data
		log.Println("Download Name:", manifest.Name)
		log.Println("Download Hash:", manifest.Hash)

		// ------------------------------------------------------------
		// !P2P Download
		tabSelected = ProgressTab

		peerID, err := dhtLookup(ui, manifest.Hash)
		if err != nil {
			PopupMessage("Cannot download file \ndue to no providers")
			log.Println("Error finding provider:", err)
			return
		}

		// Add file to the progress page
		downloadFile := prgUI.AddDownload(fileName, ui.selectedResult.Hash)

		// TODO Update download progress = data received / file size
		downloadFile.Progress = 0

		// TODO Keep checking next provider or trying until some providers is online
		if err := requestFile(ui.node, peerID, manifest.Name, ui.dirPath, manifest.Hash); err != nil {
			if strings.HasPrefix(err.Error(), "File Not Found") {
				PopupMessage("File Not Found")
			} else if strings.HasPrefix(err.Error(), "Integrity check failed") {
				PopupMessage("Integrity check failed")
			} else {
				PopupMessage("Fail to get file from peers")
			}
			return
		}

		// Set the progress bar to 100% (hardcode for now)
		downloadFile.Progress = 1

		PopupMessage(fileName + " download finish")
	}()
}

func requestFile(node host.Host, providerID peer.ID, fileName string, dirPath string, expectedCID string) error {
	// Open a new stream to the provider
	ctx := context.Background()

	// Log which peer we're contacting
	log.Println("Attempting to request file from provider:", providerID.String())

	s, err := node.NewStream(ctx, providerID, protocol)
	if err != nil {
		return err
	}
	defer s.Close()

	// Send the file request
	log.Printf("Requesting file: %s\n", fileName)
	if _, err := s.Write([]byte(fileName + "\n")); err != nil {
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
		log.Printf("Failed to get file %s from the provider %s\n", fileName, providerID.String())
		return fmt.Errorf("File Not Found from the provider")
	}

	// Construct the correct save path in the peer's directory
	savePath := filepath.Join(dirPath, fileName)

	// Write the file data to correct directory
	err = os.WriteFile(savePath, data, 0644)
	if err != nil {
		return err
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
		return fmt.Errorf("Integrity check failed")
	}

	return nil
}

func dhtLookup(ui *DownloadUI, fileCID string) (peer.ID, error) {
	ctx := context.Background()

	// Decode the CID from the string
	c, err := cid.Decode(fileCID)
	if err != nil {
		return "", fmt.Errorf("failed to convert CID: %v", err)
	}

	providerChan := ui.kadDHT.FindProvidersAsync(ctx, c, 10)

	// Log all found providers
	var foundProvider peer.ID
	for p := range providerChan {
		log.Println("Discovered provider:", p.ID.String())
		if p.ID != ui.node.ID() {
			foundProvider = p.ID
			break
		}
	}

	if foundProvider == "" {
		return "", fmt.Errorf("no providers found for CID: %s", fileCID)
	}

	log.Println("Using provider:", foundProvider.String())
	return foundProvider, nil
}
