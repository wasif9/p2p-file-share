package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/joho/godotenv"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

const (
	DownloadTab = iota
	UploadTab
	ProgressTab
)
const (
	Protocol  = "/file-sharing/1.0.0"
	TmpFolder = ".p2p"
)

var (
	TabSelected      = DownloadTab
	DBManagerVer     string
	DataDir          string
	ReverseProxyAddr string
	PWD              string
	Node             host.Host
	KadDHT           *dht.IpfsDHT
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		log.Fatalln("usage: go run ./... <superconfig-file>")
	}
	superConfigBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	superConfig := types.SuperConfig{}
	err = json.Unmarshal(superConfigBytes, &superConfig)
	if err != nil {
		log.Fatal(err)
	}

	ReverseProxyAddr = superConfig.RpConfig.Address

	if len(superConfig.DbManagerConfigs) < 1 {
		log.Fatalf("There is no db-manager in %s", os.Args[1])
	}
	DBManagerVer = superConfig.DbManagerConfigs[0].Version

	// Get Current work directory
	PWD, err = os.Getwd()
	if err != nil {
		log.Fatal("Error when getting current directory", err)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peer App fail to load .env file", err)
	}
	bootstrapAddr := os.Getenv("BOOTSTRAP_ADDR")
	if bootstrapAddr == "" {
		log.Fatal("BOOTSTRAP_ADDR environment variable not set. Are you running peer-app in the same directory as the .env file?")
	}

	runGUI(bootstrapAddr)
	app.Main()
}

// GUI function
func runGUI(bootstrapAddr string) {
	w := new(app.Window)
	w.Option(app.Title("Peerify"))
	w.Option(app.Size(unit.Dp(800), unit.Dp(600)))
	th := material.NewTheme()
	var ops op.Ops
	set := false

	// Make mainUI tab buttons
	tabButtons := make([]widget.Clickable, 3)

	selectDir_Dn := &SelectUI{
		title:   "Select a Folder for Storing Downloaded Files",
		dirPath: PWD,
		list: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		done: false,
	}
	selectDir_Dn.LoadDirs()

	selectDir_Up := &SelectUI{
		title:   "Select a Folder for Uploading Files",
		dirPath: PWD,
		list: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		done: false,
	}
	selectDir_Up.LoadDirs()

	dnUI := &DownloadUI{
		list: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		loading: false,
	}
	upUI := &UploadUI{
		list: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
	}

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

			if !selectDir_Dn.done {
				selectDir_Dn.SelectLayout(gtx, th)
			} else if !selectDir_Up.done {
				selectDir_Up.SelectLayout(gtx, th)
			} else {
				if !set {
					dnUI.dirPath = selectDir_Dn.dirPath
					upUI.dirPath = selectDir_Up.dirPath
					upUI.LoadFiles()
					DataDir = selectDir_Up.dirPath

					// Create a new libp2p node + Kademlia DHT
					ctx := context.Background()
					setupNode(ctx, bootstrapAddr)

					// Print Node ID & its address
					log.Println("Node ID:", Node.ID())
					log.Println(" -", Node.Addrs()[0], "/p2p/", Node.ID())

					// Handle inbound file requests on our protocol
					Node.SetStreamHandler(Protocol, handleFileRequest)

					set = true
				}
				MainLayout(gtx, th, dnUI, upUI, prUI, tabButtons)
			}

			e.Frame(gtx.Ops)
		}
	}
}

// setupNode creates a libp2p host + Kademlia DHT, optionally connects to a bootstrap node.
func setupNode(ctx context.Context, bootstrapAddr string) {
	priv := loadOrGeneratePrivateKey()
	node, err := libp2p.New(libp2p.Identity(priv))
	if err != nil {
		log.Fatal("Failed to create libp2p host:", err)
	}
	kad, err := dht.New(ctx, node, dht.Mode(dht.ModeServer))
	if err != nil {
		log.Fatal("Failed to create DHT:", err)
	}
	if err := kad.Bootstrap(ctx); err != nil {
		log.Fatal("Failed to bootstrap DHT:", err)
	}

	log.Println("Dialing bootstrap:", bootstrapAddr)
	ma, err := multiaddr.NewMultiaddr(bootstrapAddr)
	if err != nil {
		log.Fatalf("Invalid bootstrap address: %s", err)
	}
	info, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		log.Fatalf("AddrInfoFromP2pAddr failed: %s", err)
	}
	log.Println("Connecting to bootstrap peer:", info.ID)
	if err := node.Connect(ctx, *info); err != nil {
		log.Fatalf("Failed to connect to bootstrap: %s", err)
	}

	announceKey, err := cidFromString("/myapp/peers")
	if err != nil {
		log.Fatalf("Error creating CID: %v", err)
	}
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
	go func() {
		for {
			ctxFind, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			peerChan := kad.FindProvidersAsync(ctxFind, announceKey, 10)
			for p := range peerChan {
				if p.ID != node.ID() {
					//log.Printf("Discovered peer: %s\n", p.ID.String())
					err := node.Connect(ctx, p)
					if err != nil {
						//log.Printf("Failed to connect to %s: %v", p.ID, err)
						continue
					}
					log.Printf("Discovered peer: %s\n", p.ID.String())
					node.Peerstore().AddAddrs(p.ID, p.Addrs, time.Hour)
				}
			}
			cancel()
			time.Sleep(15 * time.Second)
		}
	}()

	Node, KadDHT = node, kad
}

func loadOrGeneratePrivateKey() crypto.PrivKey {
	var keyFilePath = filepath.Join(DataDir, TmpFolder, "peer.key")
	if keyData, err := os.ReadFile(keyFilePath); err == nil {
		key, err := crypto.UnmarshalPrivateKey(keyData)
		if err != nil {
			log.Fatalf("Failed to unmarshal saved private key: %v", err)
		}
		return key
	}
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}
	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyFilePath), 0755); err != nil {
		log.Fatalf("Failed to create key dir: %v", err)
	}
	if err := os.WriteFile(keyFilePath, keyBytes, 0600); err != nil {
		log.Fatalf("Failed to write private key: %v", err)
	}
	return priv
}
func handleFileRequest(s network.Stream) {
	defer func() {
		if err := s.Close(); err != nil {
			log.Fatal("Peer app error when close file request stream", err)
		}
	}()

	// Read requested filename
	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}
	requestedFile := strings.TrimSpace(string(buf[:n]))
	log.Println("Received request for file:", requestedFile)

	// Construct file path based on peer's directory
	filePath := filepath.Join(DataDir, requestedFile)
	log.Println("File path:", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("File not found on provider node:", filePath)
		if _, err := s.Write([]byte("Error: open " + filePath + ": no such file or directory\n")); err != nil {
			log.Println("Error writing to stream:", err)
		}
		return
	}

	// Read file and send it
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("Error reading file:", err)
		return
	}
	if _, err := s.Write(data); err != nil {
		log.Println("Error writing to stream:", err)
		return
	}
	log.Println("Sent file:", requestedFile)
}
