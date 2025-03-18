package main

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	DownloadTab = iota
	UploadTab
	ProgressTab
)
const (
	LoadBalancerAdr = "http://localhost:8080"
	DBManagerVer    = "v1"
	protocol        = "/file-sharing/1.0.0"
)

var tabSelected = DownloadTab

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ctx := context.Background()

	// Optional: pass a bootstrap multiaddr via ENV
	bootstrapAddr := os.Getenv("BOOTSTRAP_ADDR")

	// 1) Create a new libp2p node + Kademlia DHT
	node, kad := setupNode(ctx, bootstrapAddr)
	go func() {
		for {
			peers := node.Peerstore().Peers()
			fmt.Println("Known peers:", peers)
			time.Sleep(5 * time.Second) // Print every 5 seconds
		}
	}()

	fmt.Println("Node ID:", node.ID())
	for _, addr := range node.Addrs() {
		fmt.Println(" -", addr, "/p2p/", node.ID())
	}

	// 2) Handle inbound file requests on our protocol
	node.SetStreamHandler(protocol, handleFileRequest)

	// 3) Create GUI
	go func() {
		w := new(app.Window)
		w.Option(app.Title("P2P File Share"))
		w.Option(app.Size(unit.Dp(800), unit.Dp(600)))
		th := material.NewTheme()
		var ops op.Ops

		// Make tab buttons
		tabButtons := make([]widget.Clickable, 3)

		// Construct your DownloadUI, now with kadDHT
		dnUI := &DownloadUI{
			node:   node,
			kadDHT: kad,
			list: widget.List{
				List: layout.List{Axis: layout.Vertical},
			},
		}
		upUI := &UploadUI{
			dirPath: "./p2-dir",
			node:    node,
			kadDHT:  kad,
		}
		upUI.LoadFiles()
		//upUI.LoadFilesAgain(node, kad)

		// ProgressUI state instance
		prUI := &ProgressUI{
			list: widget.List{
				List: layout.List{
					Axis: layout.Vertical,
				},
			},
		}

		// The event loop
		for {
			e := w.Event()
			switch e := e.(type) {
			case app.DestroyEvent:
				log.Fatal(e.Err)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				// Layout the tabs
				layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					// Left side: tab buttons
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							tab_Btn(th, gtx, &tabButtons[0], "Download", DownloadTab),
							tab_Btn(th, gtx, &tabButtons[1], "Upload", UploadTab),
							tab_Btn(th, gtx, &tabButtons[2], "Progress", ProgressTab),
						)
					}),
					// Right side: whichever tab is selected
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						switch tabSelected {
						case DownloadTab:
							return dnUI.DownloadLayout(gtx, th, prUI)
						case UploadTab:
							return upUI.UploadLayout(gtx, th)
						case ProgressTab:
							return prUI.ProgressLayout(gtx, th)
						}
						return layout.Dimensions{}
					}),
				)
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}

// setupNode creates a libp2p host + Kademlia DHT, optionally connects to a bootstrap node.
func setupNode(ctx context.Context, bootstrapAddr string) (host.Host, *dht.IpfsDHT) {
	// 1) Create a libp2p host
	node, err := libp2p.New()
	if err != nil {
		log.Fatal("Failed to create libp2p host:", err)
	}

	// 2) Create a DHT
	kad, err := dht.New(ctx, node)
	if err != nil {
		log.Fatal("Failed to create DHT:", err)
	}

	// 3) Bootstrap the DHT
	if err := kad.Bootstrap(ctx); err != nil {
		log.Fatal("Failed to bootstrap DHT:", err)
	}

	// 4) If we have a bootstrap multiaddr, connect to it
	if bootstrapAddr != "" {
		fmt.Println("Dialing bootstrap:", bootstrapAddr)
		ma, err := multiaddr.NewMultiaddr(bootstrapAddr)
		if err != nil {
			log.Fatalf("Invalid bootstrap address: %s", err)
		}
		info, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			log.Fatalf("AddrInfoFromP2pAddr failed: %s", err)
		}

		// Just call node.Connect(...), no need for (*host).Connect
		if err := node.Connect(ctx, *info); err != nil {
			log.Fatalf("Failed to connect to bootstrap: %s", err)
		}
		fmt.Println("Connected to bootstrap!")
		// Explicitly call FindPeer to populate your local routing table
		// ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		// defer cancel()

		// fmt.Println("Populating routing table...")
		// peerInfo, err := kad.FindPeer(ctx, info.ID)
		// if err != nil {
		// 	log.Printf("FindPeer error (might be okay initially): %v\n", err)
		// } else {
		// 	fmt.Printf("Peer found: %v\n", peerInfo)
		// }
		// Announce your node's presence clearly:
		announceKey, err := cidFromString("/myapp/peers")
		if err != nil {
			log.Fatalf("Error creating CID: %v", err)
		}

		// Announce your node periodically
		go func() {
			for {
				ctxProvide, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := kad.Provide(ctxProvide, announceKey, true); err != nil {
					log.Printf("Provide error: %v\n", err)
				}
				cancel()
				time.Sleep(30 * time.Second)
			}
		}()

		// Discover peers periodically
		go func() {
			for {
				ctxFind, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				peerChan := kad.FindProvidersAsync(ctxFind, announceKey, 10)
				for p := range peerChan {
					if p.ID != node.ID() {
						log.Printf("Discovered peer: %s\n", p.ID.String())
						node.Peerstore().AddAddrs(p.ID, p.Addrs, time.Hour)
					}
				}
				cancel()
				time.Sleep(15 * time.Second)
			}
		}()
	}

	return node, kad
}

// A simple stream handler for inbound file requests
func handleFileRequest(s network.Stream) {
	defer s.Close()

	buf := bufio.NewReader(s)
	fileName, err := buf.ReadString('\n')
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}
	fileName = strings.TrimSpace(fileName)
	log.Printf("Received request for file: %s\n", fileName)

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		msg := fmt.Sprintf("Error: %v\n", err)
		s.Write([]byte(msg))
		return
	}
	s.Write(data)
	log.Printf("Sent file: %s\n", fileName)
}

// Tab button layout, same as your snippet
func tab_Btn(th *material.Theme, gtx layout.Context, button *widget.Clickable, title string, tab int) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{Top: 10, Right: 20, Bottom: 10, Left: 20}
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(th, button, title)
			gtx.Constraints.Min.X = gtx.Dp(120)
			gtx.Constraints.Min.Y = gtx.Dp(60)
			btn.TextSize = unit.Sp(16)
			btn.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
			btn.Background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			if button.Clicked(gtx) {
				tabSelected = tab
			}
			return btn.Layout(gtx)
		})
	})
}
