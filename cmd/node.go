package cmd

import "git.parallelcoin.io/dev/9/cmd/node"

func runNode(args []string, tokens Tokens, cmds, all Commands) int {
	setAppDataDir("node")
	if validateWhitelists() != 0 ||
		validateProxyListeners() != 0 ||
		validatePasswords() != 0 ||
		validateRPCCredentials() != 0 ||
		validateBlockLimits() != 0 ||
		validateUAComments() != 0 ||
		validateMiner() != 0 ||
		validateCheckpoints() != 0 ||
		validateAddresses() != 0 ||
		validateDialers() != 0 {
		return 1
	}
	// run the node!
	node.StateCfg = stateconfig
	node.Cfg = config
	node.ActiveNetParams = activenetparams
	if node.Main(nil) != nil {
		return 1
	}
	return 0
}
