package main

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/wasif9/p2p-file-share/peer-app/discovery"
	"github.com/wasif9/p2p-file-share/peer-app/messaging"
	"github.com/wasif9/p2p-file-share/peer-app/transfer"
)

const (
	LoadBalancerAdr = "http://localhost:8080"
	DBManagerVer    = "v1"
	protocol        = "/file-sharing/1.0.0"
)

const (
	DownloadTab = iota // 0
	UploadTab          // 1
	ProgressTab        // 2
)

var tabSelected = DownloadTab
var node host.Host

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create a new libp2p node (peer)
	node, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	// Print the node's multiaddresses
	fmt.Println("Node ID:", node.ID().String())
	// Enable mDNS-based peer discovery
	discovery.SetupDiscovery(node)
	// Handle incoming messages
	messaging.HandleMessages(node)
	// Handle incoming file requests
	transfer.HandleFileRequests(node)

	node.SetStreamHandler(protocol, func(s network.Stream) {
		handleFileRequest(s)
	})

	go func() {
		w := new(app.Window)
		w.Option(app.Title("P2P File Share"))
		w.Option(app.Size(unit.Dp(800), unit.Dp(600)))

		th := material.NewTheme()
		var ops op.Ops

		tabButtons := make([]widget.Clickable, 3)

		// DownloadUI state instance
		dnUI := &DownloadUI{
			list: widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
		}

		// ProgressUI state instance
		prgUI := &ProgressUI{
			list: widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
		}

		// UploadUI state instance
		upUI := &UploadUI{
			dirPath: ".",
			list: widget.List{
				List: layout.List{Axis: layout.Vertical},
			},
		}
		upUI.LoadFiles()

		// Event
		for {
			e := w.Event()
			switch e := e.(type) {
			case app.DestroyEvent:
				log.Fatal(e.Err)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				// Layout the tabs
				layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					// Tabs on the left
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							tab_Btn(th, &tabButtons[0], "Download", DownloadTab),
							tab_Btn(th, &tabButtons[1], "Upload", UploadTab),
							tab_Btn(th, &tabButtons[2], "Progress", ProgressTab),
						)
					}),

					// Right content of the selected tab
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if tabSelected == DownloadTab {
							return dnUI.DownloadLayout(gtx, th, prgUI)
						} else if tabSelected == UploadTab {
							return upUI.UploadLayout(gtx, th)
						} else {
							return prgUI.ProgressLayout(gtx, th)
						}
					}),
				)
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}

func tab_Btn(th *material.Theme, button *widget.Clickable, title string, tab int) layout.FlexChild {
	return layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
		// Padding
		inset := layout.Inset{
			Top:    10,
			Right:  20,
			Bottom: 10,
			Left:   20,
		}

		// Apply the padding and layout the button
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(th, button, title)

			// Button property
			gtx.Constraints.Min.X = gtx.Dp(120)
			gtx.Constraints.Min.Y = gtx.Dp(80)
			btn.TextSize = unit.Sp(20)
			btn.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
			btn.Background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

			// Click event
			if (*button).Clicked(gtx) {
				tabSelected = tab
			}

			return btn.Layout(gtx)
		})
	})
}

func handleFileRequest(s network.Stream) {
	defer s.Close()

	// Read the request
	reader := bufio.NewReader(s)
	request, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}

	requestedFile := strings.TrimSpace(request)
	fmt.Printf("Received request for file: %s\n", requestedFile)

	// Try to read the file
	data, err := os.ReadFile(requestedFile)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", requestedFile, err)
		s.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
		return
	}

	// Send the file content
	s.Write(data)
	fmt.Printf("Sent file: %s\n", requestedFile)
}
