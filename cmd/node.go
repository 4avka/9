package cmd

func runNode(args []string, tokens Tokens, cmds, all Commands) int {
	setAppDataDir("node")
	if validateWhitelists() != 0 {
		return 1
	}
	if validateProxyListeners() != 0 {
		return 1
	}
	if validatePasswords() != 0 {
		return 1
	}
	if validateRPCCredentials() != 0 {
		return 1
	}
	if validateBlockLimits() != 0 {
		return 1
	}
	if validateUAComments() != 0 {
		return 1
	}
	if validateMiner() != 0 {
		return 1
	}
	if validateCheckpoints() != 0 {
		return 1
	}
	if validateAddresses() != 0 {
		return 1
	}
	// run the node!

	return 0
}
