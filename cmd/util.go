package cmd

import (
	"net"
	"os"
	"path/filepath"
	"strings"

	"git.parallelcoin.io/dev/pod/cmd/node"
	"git.parallelcoin.io/dev/pod/pkg/util"
)

// CleanAndExpandPath expands environment variables and leading ~ in the passed path, cleans the result, and returns it.
func CleanAndExpandPath(path string) string {

	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {

		homeDir := filepath.Dir(util.AppDataDir("pod", false))
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%, but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// NormalizeAddresses reads and collects a space separated list of addresses contained in a string
func NormalizeAddresses(addrs string, defaultPort string, out *[]string) {

	O := new([]string)
	addrS := strings.Split(addrs, " ")

	for i := range addrS {

		a := addrS[i]

		// o := ""
		NormalizeAddress(a, defaultPort, &a)

		if a != "" {

			*O = append(*O, a)
		}

	}

	// atomically switch out if there was valid addresses
	if len(*O) > 0 {

		*out = *O
	}

}

// NormalizeAddress reads and corrects an address if it is missing pieces
func NormalizeAddress(addr, defaultPort string, out *string) {

	o := node.NormalizeAddress(addr, defaultPort)
	_, _, err := net.ParseCIDR(o)

	if err != nil {

		ip := net.ParseIP(addr)

		if ip != nil {

			*out = o
		}

	} else {

		*out = o
	}

}
