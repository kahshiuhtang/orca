package cli

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"log"

	// "crypto/x509"
	"encoding/json"
	"fmt"
	"io"

	// "log"
	"net"
	"net/http"
	orcaBlockchain "orca-peer/internal/blockchain"
	orcaClient "orca-peer/internal/client"
	"orca-peer/internal/fileshare"
	orcaHash "orca-peer/internal/hash"
	"orca-peer/internal/relay"
	"orca-peer/internal/server"
	orcaServer "orca-peer/internal/server"
	orcaStatus "orca-peer/internal/status"
	orcaStore "orca-peer/internal/store"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
)

var (
	Ip     string
	Port   int64
	Client *orcaClient.Client
)

type Settings struct {
	MarketRPCPort      string `json:"RPC_PORT"`
	MarketDHTPort      string `json:"DHT_PORT"`
	HTTPAPIPort        string `json:"API_PORT"`
	BlockchainPassword string `json:"BLOCKCHAIN_PW"`
}

func loadSetttings() (Settings, error) {
	jsonFile, err := os.Open("config/settings.json")
	if err != nil {
		return Settings{}, err
	}
	defer jsonFile.Close()
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
		return Settings{}, err
	}
	var settings Settings
	err = json.Unmarshal(byteValue, &settings)
	if err != nil {
		return Settings{}, err
	}
	for settings.MarketRPCPort == "" || settings.MarketDHTPort == "" || settings.HTTPAPIPort == "" || settings.BlockchainPassword == "" {
		if settings.MarketRPCPort == "" {
			settings.MarketRPCPort = getPort("Market RPC Server")
		}
		if settings.MarketDHTPort == "" {
			settings.MarketDHTPort = getPort("Market DHT Host")
		}
		if settings.HTTPAPIPort == "" {
			settings.HTTPAPIPort = getPort("HTTP Server")
		}
		if settings.BlockchainPassword == "" {
			settings.BlockchainPassword = getPassKey()
		}
	}
	return settings, nil
}
func StartCLI(bootstrapAddress *string, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey, orcaNetAPIProc *exec.Cmd, startAPIRoutes func(*map[string]fileshare.FileInfo)) {
	settings, err := loadSetttings()
	if err != nil {
		log.Fatal("unable to load settings")
	}
	serverReady := make(chan bool)
	confirming := false
	confirmation := ""
	locationJsonString := orcaStatus.GetLocationData()
	var locationJson map[string]interface{}
	err = json.Unmarshal([]byte(locationJsonString), &locationJson)
	if err != nil {
		fmt.Println("Unable to establish user IP, please try again")
		return
	}
	Ip = locationJson["ip"].(string)
	Port, err = strconv.ParseInt(settings.HTTPAPIPort, 10, 64)
	if err != nil {
		fmt.Println("Error parsing in port: must be a integer.", err)
		return
	}

	//Get libp2p wrapped privKey
	libp2pPrivKey, _, err := libp2pcrypto.KeyPairFromStdKey(privKey)
	if err != nil {
		panic("Could not generate libp2p wrapped key from standard private key.")
	}

	//Construct multiaddr from string and create host to listen on it
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", settings.MarketDHTPort))
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(sourceMultiAddr.String()),
		libp2p.Identity(libp2pPrivKey), //derive id from private key
		libp2p.EnableRelay(),
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	hostMultiAddr := ""
	fmt.Printf("\nlibp2p DHT Host ID: %s\n", host.ID())
	fmt.Println("DHT Market Multiaddr (if in server mode):")
	for _, addr := range host.Addrs() {
		if !strings.Contains(fmt.Sprintf("%s", addr), "127.0.0.1") {
			hostMultiAddr = fmt.Sprintf("%s/p2p/%s", addr, host.ID())
		}
		fmt.Printf("%s/p2p/%s\n", addr, host.ID())
	}

	Client = orcaClient.NewClient("files/names/")
	Client.PrivateKey = privKey
	Client.PublicKey = pubKey
	Client.Host = host
	if detectNAT() {
		// if true {
		addresses := []string{
			"/ip4/194.113.73.99/tcp/44981/p2p/QmZyLQd66AYP9sPxGbdjqZ5Ys76ZBaFFJy5PwzXxosXz74",
			"/ip4/209.151.148.27/tcp/44981/p2p/QmcAhU6MTzDeDvPhJgbk83PpT5dyB5LrZdSYaZW9K7gJm1",
			"/ip4/209.151.155.108/tcp/44981/p2p/QmYGQgBaiukGEUYqsoLAVerqBooERL13btPnLDogshiWi4",
		}
		selected := addresses[0]
		hostMultiAddr = fmt.Sprintf("%s/p2p-circuit/p2p/%s", selected, host.ID())
		//hostMultiAddr, _ = multiaddr.NewMultiaddr(selected + "/p2p-circuit/p2p/" + host.ID())
	}
	go orcaServer.StartServer(orcaServer.Settings(settings), serverReady, &confirming, &confirmation, libp2pPrivKey, Client, startAPIRoutes, host, hostMultiAddr)
	<-serverReady
	orcaBlockchain.InitBlockchainStats(pubKey)
	var cmdLocation = &cobra.Command{
		Use:   "location",
		Short: "Gets current location of THIS peer node",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(orcaStatus.GetLocationData())
		},
	}

	var cmdGet = &cobra.Command{
		Use:   "get [fileHash | fileName]",
		Short: "Get either a hash of a file or an entire file from the DHT network.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			holders, err := server.SetupCheckHolders(args[0])
			if err != nil {
				fmt.Printf("Error finding holders for file: %x", err)
				return
			}
			var bestHolder *fileshare.User
			bestHolder = nil
			for _, holder := range holders.Holders {
				if bestHolder == nil {
					bestHolder = holder
				} else if holder.GetPrice() < bestHolder.GetPrice() {
					bestHolder = holder
				}
			}
			if bestHolder == nil {
				fmt.Println("Unable to find holder for this hash.")
				return
			}
			fmt.Printf("%s - %d OrcaCoin\n", bestHolder.GetIp(), bestHolder.GetPrice())
			pubKeyInterface, err := x509.ParsePKIXPublicKey(bestHolder.Id)
			if err != nil {
				log.Fatal("failed to parse DER encoded public key: ", err)
			}
			rsaPubKey, ok := pubKeyInterface.(*rsa.PublicKey)
			if !ok {
				log.Fatal("not an RSA public key")
			}
			key := orcaServer.ConvertKeyToString(rsaPubKey.N, rsaPubKey.E)
			err = Client.GetFileOnce(bestHolder.GetIp(), bestHolder.GetPort(), args[0], key, fmt.Sprintf("%d", bestHolder.GetPrice()), settings.BlockchainPassword, "")

			if err != nil {
				fmt.Printf("Error getting file %s", err)
			}
		},
	}

	var cmdStore = &cobra.Command{
		Use:   "store",
		Short: "Inform the DHT that a specific file will be stored by the peer node",
		Long: `The DHT will keep track of files that each peer has.
				When a file is requested, the DHT will be able to inform the requester of potential peer nodes that have the file.
				The DHT also keeps track of prices and specific hashes.`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fileName := args[0]
			filePath := "./files/" + fileName
			if _, err := os.Stat(filePath); err == nil {

			} else if os.IsNotExist(err) {
				fmt.Println("file does not exist inside files folder")
				return
			} else {
				fmt.Println("error checking file's existence, please try again")
				return
			}
			costPerMB, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				fmt.Println("Error parsing in cost per MB: must be a int64", err)
				return
			}
			err = server.SetupRegisterFile(filePath, fileName, costPerMB, Ip, int32(Port))
			if err != nil {
				fmt.Printf("Unable to register file on DHT: %s", err)
			} else {
				fmt.Println("Sucessfully registered file on DHT.")
			}
		},
	}
	var cmdNetwork = &cobra.Command{
		Use:   "network",
		Short: "Print out information about the network status of the peer node",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Testing Network Speeds...")
			networkData := orcaStatus.GetNetworkInfo()
			if networkData.Success {
				fmt.Printf("Latency: %fms, Download: %fMbps, Upload: %fMbps\n", networkData.LatencyMs, networkData.DownloadSpeedMbps, networkData.UploadSpeedMbps)
			} else {
				fmt.Println("Unable to test network speeds. Please try again")
			}
		},
	}
	var cmdImport = &cobra.Command{
		Use:   "import [filePath]",
		Short: "Import a file from your local computer into the peer node file system.",
		Long: `Files need to be imported into the file system in order to be sent to other peer nodes.
				You are able to remove files from the file system.
				It is also possible to manually move files into the peer node.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := Client.ImportFile(args[0])
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List all files that are being stored by the peer node",
		Long: `Within the 'files' directory, there is a mini file system that tracks all files that peer node has requested and stored.
				A stored file is a file that the peer node makes accessible to everyone.
				A requested file is a file that has been downloaded by the peer node from another peer node.
				There are also parts of a file, called hashes, stored in the file directory`,
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			files := orcaStore.GetAllLocalFiles()
			fmt.Print("Files found: \n")
			for _, file := range files {
				fmt.Println(file.Name)
			}
		},
	}

	var cmdHash = &cobra.Command{
		Use:   "hash [filePath]",
		Short: "Return the hash of a file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			orcaHash.HashFile(args[0])
		},
	}
	var cmdSend = &cobra.Command{
		Use:   "send [amount] [ip] [port]",
		Short: "Sends some amount of orcaCoin currency to an address, defined by an ip + port",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			cost, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				fmt.Println("Error parsing amount to send")
				return
			}
			orcaClient.SendTransaction(cost, args[1], args[2], pubKey, privKey)
		},
	}
	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the peer node as a backend server",
		Long: `This will run the peer node until SIGINT or SIGTERM is received in the terminal.
				This functionality is provided to allow for the peer node to run as a backend for the GUI.
				You will not be able to interact with the peer node in this state without using API calls.
				`,
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			<-sigs
		},
	}
	var cmdRelay = &cobra.Command{
		Use:   "relay",
		Short: "Run the node as a relay server for other peers",
		Long: `This will run the peer node until SIGINT or SIGTERM is received in the terminal.
				This functionality is provided to allow for the peer node to run as a relay server for any nodes behind a NAT.
				`,
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			go relay.SetupRelay()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			<-sigs
		},
	}

	var rootCmd = &cobra.Command{Use: "orca"}
	rootCmd.AddCommand(cmdLocation, cmdGet, cmdStore, cmdNetwork, cmdImport, cmdList, cmdHash, cmdSend, cmdRun, cmdRelay)
	rootCmd.Execute()
}

// Ask user to enter a port and returns it
func getPort(useCase string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter a port number to start listening to requests for %s: ", useCase)
	for {
		port, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			os.Exit(1)
		}
		port = strings.TrimSpace(port)

		// Validate port
		listener, err := net.Listen("tcp", ":"+port)
		if err == nil {
			defer listener.Close()
			return port
		}

		fmt.Print("Invalid port. Please enter a different port: ")
	}
}

func getPassKey() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your blockchain wallet passkey: ")
	passKey, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}
	passKey = strings.TrimSpace(passKey)
	return passKey
}

// detectNAT simulates the process of detecting whether the node is behind NAT.
func detectNAT() bool {
	ipapiClient := http.Client{}

	ipv4Req, err := http.NewRequest("GET", "http://httpbin.org/ip", nil)
	if err != nil {
		fmt.Println("Error creating IPv4 request:", err)
		os.Exit(1)
	}
	resp, err := ipapiClient.Do(ipv4Req)
	if err != nil {
		fmt.Println("Error retrieving IPv4:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading IPv4 response body:", err)
		os.Exit(1)
	}

	var ipv4JSON struct {
		Origin string `json:"origin"`
	}
	err = json.Unmarshal(body, &ipv4JSON)
	if err != nil {
		fmt.Println("Error unmarshalling IPv4 response body:", err)
		os.Exit(1)
	}

	publicIP := net.ParseIP(ipv4JSON.Origin)

	// Define private IP address ranges.
	privateRanges := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
	}

	// Check if the public IP address is within any of the private IP address ranges.
	for _, pr := range privateRanges {
		if pr.Contains(publicIP) {
			return false
		}
	}
	return true
}
