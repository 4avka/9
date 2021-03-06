package walletmain
import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	legacyrpc "git.parallelcoin.io/dev/9/pkg/rpc/legacy"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
	"git.parallelcoin.io/dev/9/pkg/util/interrupt"
	"git.parallelcoin.io/dev/9/pkg/wallet"
	chain "git.parallelcoin.io/dev/9/pkg/wallet/chain"
)
var (
	cfg *nine.Config
)
// Main is a work-around main function that is required since deferred functions (such as log flushing) are not called with calls to os.Exit.
// Instead, main runs this function and checks for a non-nil error, at point any defers have already run, and if the error is non-nil, the program can be exited with an error exit status.
func Main(c *nine.Config, activeNet *nine.Params, path string) error {
	// fmt.Println("wallet Main")
	cfg = c
	ActiveNet = activeNet
	if ActiveNet.Name == "testnet" {
		fork.IsTestnet = true
	}
	if cfg.Profile != nil {
		go func() {
			listenAddr :=
				net.JoinHostPort("127.0.0.1", fmt.Sprint(*cfg.Profile))
			log <- cl.Info{"profile server listening on", listenAddr}
			profileRedirect := http.RedirectHandler("/debug/pprof",
				http.StatusSeeOther)
			http.Handle("/", profileRedirect)
			log <- cl.Error{http.ListenAndServe(listenAddr, nil)}
		}()
	}
	// dbDir := NetworkDir(path, activeNet.Params)
	log <- cl.Debug{"dbDir", path, *cfg.DataDir, *cfg.DataDir, activeNet.Params.Name}
	loader := wallet.NewLoader(activeNet.Params, path, 250)
	// Create and start HTTP server to serve wallet client connections.
	// This will be updated with the wallet and chain server RPC client
	// created below after each is created.
	log <- cl.Trc("startRPCServers loader")
	rpcs, legacyRPCServer, err := startRPCServers(loader)
	if err != nil {
		log <- cl.Error{
			"unable to create RPC servers:", err,
		}
		return err
	}
	// Create and start chain RPC client so it's ready to connect to
	// the wallet when loaded later.
	if !*cfg.NoInitialLoad {
		log <- cl.Trc("starting rpcClientConnectLoop")
		go rpcClientConnectLoop(legacyRPCServer, loader)
	}
	loader.RunAfterLoad(func(w *wallet.Wallet) {
		log <- cl.Trc("starting startWalletRPCServices")
		startWalletRPCServices(w, rpcs, legacyRPCServer)
	})
	if !*cfg.NoInitialLoad {
		log <- cl.Debug{"loading database"}
		// Load the wallet database.  It must have been created already
		// or this will return an appropriate error.
		if cfg.WalletPass != nil {
			_, err = loader.OpenExistingWallet([]byte(*cfg.WalletPass), true)
		} else {
			_, err = loader.OpenExistingWallet([]byte{}, true)
		}
		if err != nil {
			fmt.Println(err)
			log <- cl.Error{err}
			return err
		}
	}
	log <- cl.Trc("adding interrupt handler to unload wallet")
	// Add interrupt handlers to shutdown the various process components
	// before exiting.  Interrupt handlers run in LIFO order, so the wallet
	// (which should be closed last) is added first.
	interrupt.AddHandler(func() {
		err := loader.UnloadWallet()
		if err != nil && err != wallet.ErrNotLoaded {
			log <- cl.Error{
				"failed to close wallet:", err,
			}
		}
	})
	if rpcs != nil {
		log <- cl.Trc("starting rpc server")
		interrupt.AddHandler(func() {
			// TODO: Does this need to wait for the grpc server to
			// finish up any requests?
			log <- cl.Wrn("stopping RPC server...")
			rpcs.Stop()
			log <- cl.Inf("RPC server shutdown")
		})
	}
	if legacyRPCServer != nil {
		interrupt.AddHandler(func() {
			log <- cl.Wrn("stopping legacy RPC server...")
			legacyRPCServer.Stop()
			log <- cl.Inf("legacy RPC server shutdown")
		})
		go func() {
			<-legacyRPCServer.RequestProcessShutdown()
			interrupt.Request()
		}()
	}
	<-interrupt.HandlersDone
	log <- cl.Inf("shutdown complete")
	return nil
}
func readCAFile() []byte {
	// Read certificate file if TLS is not disabled.
	var certs []byte
	if !*cfg.NoTLS {
		var err error
		certs, err = ioutil.ReadFile(*cfg.CAFile)
		if err != nil {
			log <- cl.Warn{
				"cannot open CA file:", err,
			}
			// If there's an error reading the CA file, continue
			// with nil certs and without the client connection.
			certs = nil
		}
	} else {
		log <- cl.Inf("chain server RPC TLS is disabled")
	}
	return certs
}
// rpcClientConnectLoop continuously attempts a connection to the consensus RPC server.
// When a connection is established, the client is used to sync the loaded wallet, either
// immediately or when loaded at a later time.
//
// The legacy RPC is optional.  If set, the connected RPC client will be
// associated with the server for RPC passthrough and to enable additional
// methods.
func rpcClientConnectLoop(legacyRPCServer *legacyrpc.Server, loader *wallet.Loader) {
	var certs []byte
	// if !cfg.UseSPV {
	certs = readCAFile()
	// }
	for {
		var (
			chainClient chain.Interface
			err         error
		)
		// if cfg.UseSPV {
		// 	var (
		// 		chainService *neutrino.ChainService
		// 		spvdb        walletdb.DB
		// 	)
		// 	netDir := networkDir(cfg.AppDataDir.Value, ActiveNet.Params)
		// 	spvdb, err = walletdb.Create("bdb",
		// 		filepath.Join(netDir, "neutrino.db"))
		// 	defer spvdb.Close()
		// 	if err != nil {
		// 		log<-cl.Errorf{"unable to create Neutrino DB: %s", err)
		// 		continue
		// 	}
		// 	chainService, err = neutrino.NewChainService(
		// 		neutrino.Config{
		// 			DataDir:      netDir,
		// 			Database:     spvdb,
		// 			ChainParams:  *ActiveNet.Params,
		// 			ConnectPeers: cfg.ConnectPeers,
		// 			AddPeers:     cfg.AddPeers,
		// 		})
		// 	if err != nil {
		// 		log<-cl.Errorf{"couldn't create Neutrino ChainService: %s", err)
		// 		continue
		// 	}
		// 	chainClient = chain.NewNeutrinoClient(ActiveNet.Params, chainService)
		// 	err = chainClient.Start()
		// 	if err != nil {
		// 		log<-cl.Errorf{"couldn't start Neutrino client: %s", err)
		// 	}
		// } else {
		chainClient, err = startChainRPC(certs)
		if err != nil {
			log <- cl.Error{
				"unable to open connection to consensus RPC server:", err}
			continue
		}
		// }
		// Rather than inlining this logic directly into the loader
		// callback, a function variable is used to avoid running any of
		// this after the client disconnects by setting it to nil.  This
		// prevents the callback from associating a wallet loaded at a
		// later time with a client that has already disconnected.  A
		// mutex is used to make this concurrent safe.
		associateRPCClient := func(w *wallet.Wallet) {
			w.SynchronizeRPC(chainClient)
			if legacyRPCServer != nil {
				legacyRPCServer.SetChainServer(chainClient)
			}
		}
		mu := new(sync.Mutex)
		loader.RunAfterLoad(func(w *wallet.Wallet) {
			mu.Lock()
			associate := associateRPCClient
			mu.Unlock()
			if associate != nil {
				associate(w)
			}
		})
		chainClient.WaitForShutdown()
		mu.Lock()
		associateRPCClient = nil
		mu.Unlock()
		loadedWallet, ok := loader.LoadedWallet()
		if ok {
			// Do not attempt a reconnect when the wallet was explicitly stopped.
			if loadedWallet.ShuttingDown() {
				return
			}
			loadedWallet.SetChainSynced(false)
			// TODO: Rework the wallet so changing the RPC client does not require stopping and restarting everything.
			loadedWallet.Stop()
			loadedWallet.WaitForShutdown()
			loadedWallet.Start()
		}
	}
}
// startChainRPC opens a RPC client connection to a pod server for blockchain
// services.  This function uses the RPC options from the global config and
// there is no recovery in case the server is not available or if there is an
// authentication error.  Instead, all requests to the client will simply error.
func startChainRPC(certs []byte) (*chain.RPCClient, error) {
	// log <- cl.Infof{
	// 	"attempting RPC client connection to %v, TLS: %s",
	// 	*cfg.RPCConnect, fmt.Sprint(!*cfg.NoTLS),
	// }
	// spew.Dump(cfg)
	rpcc, err := chain.NewRPCClient(
		ActiveNet.Params,
		*cfg.RPCConnect,
		*cfg.Username,
		*cfg.Password,
		certs,
		!*cfg.NoTLS, 0)
	if err != nil {
		return nil, err
	}
	err = rpcc.Start()
	return rpcc, err
}
