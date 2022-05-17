package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

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
	defaultAddr    = cli.GetEnv("ADDR", "localhost:8088")

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
	r.HandleFunc("/x", HandleRoot)
	r.HandleFunc("/{txHash}", HandleTx)
	r.HandleFunc("/tx", HandleTx)
	r.HandleFunc("/tx/", HandleTx)
	r.HandleFunc("/tx/{txHash}", HandleTx)

	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	loggedRouter := httplogger.LoggingMiddleware(r)

	log.Info("HTTP server running", "addr", *addr)
	err = http.ListenAndServe(*addr, loggedRouter)
	if err != nil {
		log.Crit("webserver failed", "err", err)
	}
}

func HandleRoot(respw http.ResponseWriter, req *http.Request) {
	log.Info("aaa")
	if req.Method == "OPTIONS" {
		respw.WriteHeader(http.StatusOK)
		return
	}

	// By default, return docs
	http.ServeFile(respw, req, "./public/index.html")
}

type ResultTemplateData struct {
	Err      string
	TxKnown  bool
	TxUncled bool
	TxHash   string

	MinedBlockNumber uint64
	MinedBlockHash   string
	UncleBlockNumber uint64
	UncleBlockHash   string
}

func HandleTx(respw http.ResponseWriter, req *http.Request) {
	if req.Method == "OPTIONS" {
		respw.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(req)
	txHash := vars["txHash"]

	// Allow override through query param
	if _txHash := req.URL.Query().Get("hash"); _txHash != "" {
		txHash = _txHash
	}

	td := ResultTemplateData{
		TxHash: txHash,
	}

	t, err := template.ParseFiles("./public/result.html")
	if err != nil {
		log.Error("parsing template failed:", "err", err)
		respw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(respw, "error: %v\n", err)
		return
	}

	defer func() {
		err = t.Execute(respw, td)
		if err != nil {
			log.Error("parsing template failed:", "err", err)
			respw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(respw, "error: %v\n", err)
			return
		}
	}()

	log.Info("Check tx", "txHash", txHash)
	if len(txHash) != 66 || !strings.HasPrefix(txHash, "0x") {
		td.Err = "invalid tx hash"
		return
	}

	status, minedBlock, uncleBlock, err := txinfo.WasTxUncled(client, common.HexToHash(txHash))
	if err != nil {
		log.Error("tx check failed", "err", err)
		td.Err = fmt.Sprintf("invalid tx hash: %s", err)
		return
	}

	td.TxKnown = status != txinfo.StatusTxUnknown
	td.TxUncled = status == txinfo.StatusTxWasUncled
	td.MinedBlockNumber = minedBlock.NumberU64()
	td.MinedBlockHash = minedBlock.Hash().Hex()
	if td.TxUncled {
		td.UncleBlockNumber = uncleBlock.NumberU64()
		td.UncleBlockHash = uncleBlock.Hash().Hex()
	}
}
