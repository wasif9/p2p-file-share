package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	// your local "pkg/models" for Manifest
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

// You had this data structure from your snippet
type SearchData struct {
	fileHash string
	fileName string
	fileSize string
}

// DownloadUI now has kadDHT so we can do real lookups.
type DownloadUI struct {
	node   host.Host
	kadDHT *dht.IpfsDHT // <-- Add this to do real DHT lookups

	searchInput    widget.Editor
	searchButton   widget.Clickable
	results        []SearchData
	selectedResult SearchData
	resultButtons  []widget.Clickable
	list           widget.List
	downloadButton widget.Clickable
}

// The rest of your UI code remains mostly the same...
// DownloadLayout draws your search bar + search results list
func (ui *DownloadUI) DownloadLayout(gtx layout.Context, th *material.Theme, prgUI *ProgressUI) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Right: 20, Left: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ui.searchInput.SingleLine = true
						return material.Editor(th, &ui.searchInput, "Search for files ...").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 10, Right: 20, Left: 20}
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
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.LayoutResults(gtx, th, prgUI)
		}),
	)
}

// LayoutResults lists your search hits with a "Download" button
func (ui *DownloadUI) LayoutResults(gtx layout.Context, th *material.Theme, prgUI *ProgressUI) layout.Dimensions {
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

	return layout.Inset{Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.list.List.Layout(gtx, len(ui.results), func(gtx layout.Context, i int) layout.Dimensions {
			btn := &ui.resultButtons[i]
			isSelected := ui.selectedResult == ui.results[i]

			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
					inset := layout.Inset{Top: 5, Bottom: 5, Left: 20, Right: 20}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						dispalyStr := ui.results[i].fileName + strings.Repeat(" ", 30-len(ui.results[i].fileName)) +
							"| Size: " + strings.Repeat(" ", 8-len(ui.results[i].fileSize)) + ui.results[i].fileSize

						button := material.Button(th, btn, dispalyStr)
						if isSelected {
							button.Background = th.Palette.ContrastBg
							button.Color = th.Palette.ContrastFg
						} else {
							button.Background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
							button.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
						}

						if btn.Clicked(gtx) {
							ui.selectedResult = ui.results[i]
						}
						return button.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if isSelected {
						inset := layout.Inset{Right: 20}
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							downloadBtn := material.Button(th, &ui.downloadButton, "Download")
							downloadBtn.Background = color.NRGBA{R: 39, G: 215, B: 45, A: 255}

							if ui.downloadButton.Clicked(gtx) {
								log.Println("Selected result:", ui.selectedResult)
								ui.Download(prgUI)
							}
							return downloadBtn.Layout(gtx)
						})
					}
					return layout.Dimensions{}
				}),
			)
		})
	})
}

// Simple popup message (unchanged from your snippet)
func PopupMessage(message string) {
	go func() {
		popupWindow := new(app.Window)
		popupWindow.Option(app.Title("P2P File Share"))
		popupWindow.Option(app.Size(unit.Dp(400), unit.Dp(200)))

		thPopup := material.NewTheme()
		var popupOps op.Ops

		for {
			e := popupWindow.Event()
			switch e := e.(type) {
			case app.FrameEvent:
				gtx := app.NewContext(&popupOps, e)
				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									text := material.Body1(thPopup, message)
									text.Color = color.NRGBA{R: 255, A: 255}
									text.TextSize = unit.Sp(20)
									return text.Layout(gtx)
								})
							}),
						)
					}),
				)
				e.Frame(gtx.Ops)
			case app.DestroyEvent:
				return
			}
		}
	}()
}

// PerformSearch is your mock search that populates ui.results
func (ui *DownloadUI) PerformSearch() {
	query := strings.TrimSpace(ui.searchInput.Text())
	if query == "" {
		ui.results = nil
		ui.resultButtons = nil
		return
	}

	// Convert filename to CID-like hash
	fileHashCID, err := cidFromString(query)
	if err != nil {
		log.Println("Error generating CID for query:", err)
		return
	}

	// Find peers storing the file hash
	ctx := context.Background()
	providersChan := ui.kadDHT.FindProvidersAsync(ctx, fileHashCID, 10)

	var foundProviders []peer.AddrInfo
	for p := range providersChan {
		if p.ID == ui.node.ID() {
			continue // Skip self
		}
		foundProviders = append(foundProviders, p)
	}

	// If no providers found, return empty results
	if len(foundProviders) == 0 {
		log.Println("No peers found with file:", query)
		ui.results = nil
		ui.resultButtons = nil
		return
	}

	// Otherwise, populate results based on found peers
	ui.results = []SearchData{}
	for range foundProviders {
		ui.results = append(ui.results, SearchData{
			fileHash: fileHashCID.String(),
			fileName: query,
			fileSize: "Unknown", // No size info in DHT, might need another lookup
		})
	}

	ui.resultButtons = make([]widget.Clickable, len(ui.results))
	log.Println("Search completed. Found peers with file:", query)
}

// Here's where we do the real DHT-based download
func (ui *DownloadUI) Download(prgUI *ProgressUI) {
	go func() {
		fileName := ui.selectedResult.fileName
		// Add file to the progress page
		downloadFile := prgUI.AddDownload(fileName, ui.selectedResult.fileHash)
		if downloadFile == nil {
			return
		}
		// Make HTTP request to your load balancer (if you still want that)
		getReq := "/api/v1/records/" + fileName
		log.Println("Send GET", getReq, "to", LoadBalancerAdr)

		resp, err := http.Get(LoadBalancerAdr + getReq)
		if err != nil {
			log.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		var manifest types.Manifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			log.Println("Decode Error:", err)
			return
		}

		log.Println("Name:", manifest.Name)
		log.Println("Hash:", manifest.Hash)

		// Switch tab
		tabSelected = ProgressTab
		// "Download" from the DHT
		// 1) Convert manifest.Hash to a CID
		c, err := cidFromString(manifest.Hash)
		if err != nil {
			log.Println("Failed to create CID from hash:", err)
			return
		}

		ctx := context.Background()
		// 2) Find providers for that CID
		providersChan := ui.kadDHT.FindProvidersAsync(ctx, c, 10)
		var provider *peer.AddrInfo
		for p := range providersChan {
			if p.ID == ui.node.ID() {
				// skip self
				continue
			}
			provider = &p
			break
		}
		if provider == nil {
			log.Println("No providers found for hash:", manifest.Hash)
			return
		}

		// 3) Connect to that provider
		if err := ui.node.Connect(ctx, *provider); err != nil {
			log.Println("Failed to connect:", err)
			return
		}
		log.Println("Connected to provider:", provider.ID)

		// 4) Open a stream and request the file
		s, err := ui.node.NewStream(ctx, provider.ID, protocol)
		if err != nil {
			log.Println("NewStream error:", err)
			return
		}
		defer s.Close()

		fmt.Fprintf(s, manifest.Name+"\n") // request the file by name

		data, err := ioutil.ReadAll(s)
		if err != nil {
			log.Println("Error reading stream data:", err)
			return
		}
		if len(data) == 0 || strings.HasPrefix(string(data), "Error:") {
			log.Println("Failed to get file:", string(data))
			return
		}

		// 5) Write the file to disk
		err = ioutil.WriteFile("received_"+fileName, data, 0644)
		if err != nil {
			log.Println("Error saving file:", err)
			return
		}

		// Update progress bar to 100%
		downloadFile.Progress = 1

		PopupMessage(fileName + " download finished!")
	}()
}
