package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

type SearchData struct {
	fileHash string
	fileName string
	fileSize string
}

type DownloadUI struct {
	node           host.Host
	searchInput    widget.Editor
	searchButton   widget.Clickable
	results        []SearchData
	selectedResult SearchData
	resultButtons  []widget.Clickable
	list           widget.List
	downloadButton widget.Clickable
}

func (ui *DownloadUI) DownloadLayout(gtx layout.Context, th *material.Theme, prgUI *ProgressUI) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Upper part for text search
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Text bar
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{
						Top:   10,
						Right: 20,
						Left:  20,
					}

					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ui.searchInput.SingleLine = true
						return material.Editor(th, &ui.searchInput, "Search for files ...").Layout(gtx)
					})
				}),

				//Search Button
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{
						Top:   10,
						Right: 20,
						Left:  20,
					}

					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &ui.searchButton, "Search")
						if ui.searchButton.Clicked(gtx) {
							ui.PerformSearch()
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),

		// Lower part for search results
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.LayoutResults(gtx, th, prgUI)
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
	return layout.Inset{
		Top:    8,
		Bottom: 8,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Display all result in buttons
		return ui.list.List.Layout(gtx, len(ui.results), func(gtx layout.Context, i int) layout.Dimensions {
			btn := &ui.resultButtons[i]
			isSelected := ui.selectedResult == ui.results[i]

			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				// The result button
				layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
					// Padding
					inset := layout.Inset{
						Top:    5,
						Bottom: 5,
						Left:   20,
						Right:  20,
					}

					// Apply the padding and layout the button
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						dispaly_str := ui.results[i].fileName + strings.Repeat(" ", 30-len(ui.results[i].fileName)) +
							"| Size: " + strings.Repeat(" ", 8-len(ui.results[i].fileSize)) + ui.results[i].fileSize
						button := material.Button(th, btn, dispaly_str)

						// Different style for selected items
						if isSelected {
							// Create a highlighted button
							// button = material.Button(th, btn, ui.results[i].fileName)
							button.Background = th.Palette.ContrastBg
							button.Color = th.Palette.ContrastFg
						} else {
							// Regular button
							// button = material.Button(th, btn, ui.results[i].fileName)
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
						inset := layout.Inset{
							Right: 20,
						}

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
	query := strings.TrimSpace(ui.searchInput.Text())

	if query == "" {
		ui.results = []SearchData{}
		ui.resultButtons = nil
		return
	}

	// ! Dummy data for the query
	ui.results = []SearchData{
		{"abcdefg", "myfile.txt", "10 KB"},
		// {"fileHash 2", "fileName2", "20 KB"},
		// {"fileHash 3", "fileName3", "30 KB"},
		// {"fileHash 4", "fileName4", "40 KB"},
		// {"fileHash 5", "fileName5", "50 KB"},
		// {"fileHash 6", "fileName6", "10 MB"},
		// {"fileHash 7", "fileName7", "10 GB"},
		// {"fileHash 8", "fileName8", "10 TB"},
		// {"fileHash 9", "fileName9", "1 KB"},
		// {"fileHash 10", "fileName10", "100 KB"},
	}

	// Make buttons to select the result
	ui.resultButtons = make([]widget.Clickable, len(ui.results))
}

func (ui *DownloadUI) Download(prgUI *ProgressUI) {
	go func() {
		fileName := ui.selectedResult.fileName
		// Add file to the progress page
		downloadFile := prgUI.AddDownload(fileName, ui.selectedResult.fileHash)

		// Not process to download since file is downloading
		if downloadFile == nil {
			return
		}

		// Make GET request to the load balancer server
		getReq := "/api/v1/records/" + fileName
		log.Println("Send GET " + getReq + " to " + LoadBalancerAdr)

		// Send GET requet
		resp, err := http.Get(LoadBalancerAdr + getReq)
		if err != nil {
			log.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		// Decode the JSON response
		var manifest types.Manifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			log.Println("Decode Error:", err)
			return
		}

		// Print received data
		log.Println("Name:", manifest.Name)
		log.Println("Hash:", manifest.Hash)

		// ------------------------------------------------------------
		// !P2P Download

		// Update download progress = data received / file size
		downloadFile.Progress = 0

		tabSelected = ProgressTab

		requestFile(ui.node, dhtLookup(ui.node, manifest.Hash), manifest.Name)

		// Set the progress bar to 100% (hardcode for now)
		downloadFile.Progress = 1

		PopupMessage(fileName + " download finish")
	}()
}

func requestFile(h host.Host, providerID peer.ID, fileName string) {
	// Open a new stream to the provider
	ctx := context.Background()
	s, err := h.NewStream(ctx, providerID, protocol)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Send the file request
	fmt.Printf("Requesting file: %s\n", fileName)
	s.Write([]byte(fileName + "\n"))

	// Read the response
	data, err := io.ReadAll(s)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the response indicates an error
	response := string(data)
	if strings.HasPrefix(response, "Error:") || strings.HasPrefix(response, "File not found") {
		fmt.Println("Failed to get file:", response)
		return
	}

	// Write the file data to disk
	err = os.WriteFile("received_"+fileName, data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Received and saved file as received_%s\n", fileName)
}

func dhtLookup(node host.Host, chunkHash string) peer.ID {
	// return the first non-self peer
	for _, p := range node.Peerstore().Peers() {
		if p != node.ID() {
			return p
		}
	}

	log.Fatal("non-self peer not found")
	return ""
}
