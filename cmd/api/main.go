package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/flashbots/go-utils/cli"
	"github.com/flashbots/go-utils/httplogger"
	"github.com/gorilla/mux"
	"github.com/metachris/eth-was-tx-uncled/txinfo"
)

var (
	version = "dev" // is set during build process

	// Default values
	defaultDebug   = os.Getenv("DEBUG") == "1"
	defaultLogJSON = os.Getenv("LOG_JSON") == "1"
	defaultEthNode = os.Getenv("ETH_NODE_URI")
	defaultAddr    = cli.GetEnv("ADDR", ":8088")

	// Flags
	debugPtr      = flag.Bool("debug", defaultDebug, "print debug output")
	logJSONPtr    = flag.Bool("log-json", defaultLogJSON, "log in JSON")
	addr          = flag.String("addr", defaultAddr, "http service address")
	ethNodeUriPtr = flag.String("eth", defaultEthNode, "URL for eth node (eth node, Infura, etc.)")
)

var client *ethclient.Client

func Perror(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	logFormat := log.TerminalFormat(true)
	if *logJSONPtr {
		logFormat = log.JSONFormat()
	}

	logLevel := log.LvlInfo
	if *debugPtr {
		logLevel = log.LvlDebug
	}

	log.Root().SetHandler(log.LvlFilterHandler(logLevel, log.StreamHandler(os.Stderr, logFormat)))
	log.Info("Starting your-project", "version", version)

	var err error
	flag.Parse()

	client, err = ethclient.Dial(*ethNodeUriPtr)
	Perror(err)

	r := mux.NewRouter()
	r.HandleFunc("/", HandleRoot)
	r.HandleFunc("/{txHash}", HandleTx)
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	loggedRouter := httplogger.LoggingMiddleware(r)

	log.Info("HTTP server running", "addr", *addr)
	err = http.ListenAndServe(*addr, loggedRouter)
	if err != nil {
		log.Crit("webserver failed", "err", err)
	}
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

	log.Info("Check tx", "txHash", txHash)

	status, uncleBlock, err := txinfo.WasTxUncled(client, common.HexToHash(txHash))
	if err != nil {
		log.Info("tx check failed", "err", err)
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
