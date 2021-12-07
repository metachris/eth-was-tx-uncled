package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/metachris/eth-was-tx-uncled/txinfo"
)

var addr = flag.String("addr", ":8080", "http service address")
var ethNodeUriPtr = flag.String("eth", os.Getenv("ETH_NODE_URI"), "URL for eth node (eth node, Infura, etc.)")

var client *ethclient.Client

func Perror(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error
	flag.Parse()

	client, err = ethclient.Dial(*ethNodeUriPtr)
	Perror(err)

	r := mux.NewRouter()
	r.HandleFunc("/", HandleRoot)
	r.HandleFunc("/{txHash}", HandleTx)
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	log.Println("HTTP server running at", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}

func HandleRoot(respw http.ResponseWriter, req *http.Request) {
	http.ServeFile(respw, req, "./home.html")
}

func HandleTx(respw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	txHash := vars["txHash"]

	// Allow override through query param
	if _txHash := req.URL.Query().Get("hash"); _txHash != "" {
		txHash = _txHash
	}

	log.Println("Check tx:", txHash)

	status, uncleBlock, err := txinfo.WasTxUncled(client, common.HexToHash(txHash))
	if err != nil {
		log.Println("error:", err)
		respw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(respw, "error: %v\n", err)
		return
	}

	msg := ""
	if status == txinfo.StatusTxUnknown {
		msg = "tx not found"
	} else if status == txinfo.StatusTxNotUncled {
		msg = "tx not uncled"
	} else if status == txinfo.StatusTxWasUncled {
		msg = fmt.Sprintf("tx was uncled in block %s %s\n", uncleBlock.Number(), uncleBlock.Hash().Hex())
	}

	fmt.Fprint(respw, msg)
}
