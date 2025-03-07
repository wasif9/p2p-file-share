package main

import (
	"fmt"
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/wasif9/p2p-file-share/discovery"
	"github.com/wasif9/p2p-file-share/messaging"
	"github.com/wasif9/p2p-file-share/transfer"
)

const (
	LoadBalancerAdr = "http://localhost:8080"
)

const (
	DownloadTab = iota // 0
	UploadTab          // 1
	ProgressTab        // 2
)

var tabSelected = DownloadTab

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

	go func() {
		w := new(app.Window)
		w.Option(app.Title("P2P File Share"))
		w.Option(app.Size(unit.Dp(800), unit.Dp(600)))

		th := material.NewTheme()
		var ops op.Ops

		tabButtons := make([]widget.Clickable, 3)

		// DownloadUI instance
		dnUI := &DownloadUI{
			list: widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
			node: node,
		}

		// ProgressUI instance
		prgUI := &ProgressUI{
			list: widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
		}

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
							tab_Btn(th, gtx, &tabButtons[0], "Download", DownloadTab),
							tab_Btn(th, gtx, &tabButtons[1], "Upload", UploadTab),
							tab_Btn(th, gtx, &tabButtons[2], "Progress", ProgressTab),
						)
					}),

					// Right content of the selected tab
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if tabSelected == DownloadTab {
							return dnUI.DownloadLayout(gtx, th, prgUI)
						} else if tabSelected == UploadTab {
							return material.Body1(th, "Upload").Layout(gtx)
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

func tab_Btn(th *material.Theme, gtx layout.Context, button *widget.Clickable, title string, tab int) layout.FlexChild {
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
