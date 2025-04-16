package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net"
	"net/http"
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
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

type DownloadUI struct {
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
								go ui.performSearch()
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
							go ui.performSearch()
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
				return ui.layoutResults(gtx, th, prgUI)
			}
		}),
	)
}

func (ui *DownloadUI) layoutResults(gtx layout.Context, th *material.Theme, prgUI *ProgressUI) layout.Dimensions {
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
		return ui.list.Layout(gtx, len(ui.results), func(gtx layout.Context, i int) layout.Dimensions {
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
							button.Background = th.ContrastBg
							button.Color = th.ContrastFg
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
								ui.download(prgUI)
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

func (ui *DownloadUI) performSearch() {
	ui.loading = true
	query := strings.TrimSpace(ui.searchInput.Text())

	// Make GET request to the load balancer server
	getReq := "/api/" + DBManagerVer + "/manifests?prefix=" + query
	log.Println("Send GET " + getReq + " to " + ReverseProxyAddr)

	// Send GET requet
	reverseProxy := &http.Client{Timeout: time.Second * 2}
	resp, err := reverseProxy.Get("http://" + ReverseProxyAddr + getReq)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			PopupMessage("Cannot search files \ndue to busy servers")
		} else {
			PopupMessage("Cannot search files \ndue to proxy error")
		}
		log.Println("Error when sending search query request", err)
		ui.results = nil
		ui.loading = false
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal("Peer app error when close GET search req", err)
		}
	}()

	// Check for successful response status
	if resp.StatusCode != http.StatusOK {
		PopupMessage("Cannot get data from servers! Try again")
		log.Println("Error receiving non-OK response", resp.Status)
		ui.results = nil
		ui.loading = false
		return
	}

	// Read the response body
	respSer, err := io.ReadAll(resp.Body)
	if err != nil {
		PopupMessage("Error response from database")
		log.Println("Error reading response body:", err)
		ui.results = nil
		ui.loading = false
		return
	}

	// Decode the JSON response
	var manifests []types.Manifest
	if err := json.Unmarshal(respSer, &manifests); err != nil {
		PopupMessage("Error response from database")
		log.Println("Decode Error:", err)
		ui.results = nil
		ui.loading = false
		return
	}

	// Change the GUI search result
	ui.results = manifests
	// Make buttons to select the result
	ui.resultButtons = make([]widget.Clickable, len(ui.results))
	ui.loading = false
}

func (ui *DownloadUI) download(prgUI *ProgressUI) {
	go func() {
		fileName := ui.selectedResult.Name

		// Check file if is downloading
		if prgUI.IsInDownload(fileName, ui.selectedResult.Hash) {
			PopupMessage(fileName + " is downloading!")
			return
		}

		// Add file to the progress page
		TabSelected = ProgressTab
		downloadFile := prgUI.AddDownload(fileName, ui.selectedResult.Hash)

		// Make GET request to the load balancer server
		getReq := "/api/" + DBManagerVer + "/manifests/" + fileName
		log.Println("Send GET " + getReq + " to " + ReverseProxyAddr)

		// Send GET requet
		reverseProxy := &http.Client{Timeout: time.Second * 2}
		resp, err := reverseProxy.Get("http://" + ReverseProxyAddr + getReq)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				PopupMessage("Cannot download file \ndue to busy servers")
			} else {
				PopupMessage("Cannot download file \ndue to proxy error")
			}
			log.Println("Error when sending file query", err)
			downloadFile.Shown = false
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Fatal("Peer app error when close the GET download req", err)
			}
		}()

		// Decode the JSON response
		var manifest types.Manifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			PopupMessage("Error response from servers")
			log.Println("Decode Error:", err)
			downloadFile.Shown = false
			return
		}

		// Print received data
		log.Println("Download Name:", manifest.Name)
		log.Println("Download Hash:", manifest.Hash)

		// ------------------------------------------------------------
		// !P2P Download
		DownloadManifest(manifest.Name, manifest.Hash, filepath.Join(ui.dirPath, ".p2p"))

		downloadFile.Progress = 0

		DownloadFile(manifest.Name, ui.dirPath, downloadFile)

		PopupMessage(fileName + " download finish")
	}()
}
