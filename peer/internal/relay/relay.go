package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	peers "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func ConnectRelay() {
	peer, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Printf("Failed to create peer: %v", err)
		return
	}
	knownRelay := "130.245.173.212:45677"
	response, err := http.Get(knownRelay)
	if err != nil {
		log.Printf("Unable to connect relay: %v", err)
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	var relayResponse RelayReponse
	err = json.Unmarshal(body, &relayResponse)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	var receivedAddrInfo peers.AddrInfo

	// Unmarshal the JSON data into the AddrInfo struct
	err = json.Unmarshal(relayResponse.Addr, &receivedAddrInfo)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return
	}
	if err := peer.Connect(context.Background(), receivedAddrInfo); err != nil {
		log.Printf("Failed to connect unreachable1 and relay1: %v", err)
		return
	}
}

var addr []byte

func SetupRelay() {
	addr = createRelayServer()
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
	http.HandleFunc("/relay", handleRelay)
	http.ListenAndServe(":45677", nil)
}

type RelayReponse struct {
	Addr []byte `json:"addr"`
}

func handleRelay(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		response := RelayReponse{
			Addr: addr,
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to encode session data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
		return
	}
}
func createRelayServer() []byte {
	relay1, err := libp2p.New()
	if err != nil {
		log.Printf("Failed to create relay1: %v", err)
		return nil
	}

	_, err = relay.New(relay1)
	if err != nil {
		log.Printf("Failed to instantiate the relay: %v", err)
		return nil
	}

	relay1info := peer.AddrInfo{
		ID:    relay1.ID(),
		Addrs: relay1.Addrs(),
	}
	log.Println(relay1.ID())
	log.Println(relay1.Addrs())
	addrInfoJSON, err := json.Marshal(relay1info)
	if err != nil {
		fmt.Println("Error marshaling AddrInfo:", err)
		return nil
	}
	return addrInfoJSON
}
