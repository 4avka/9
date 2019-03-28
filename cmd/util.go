package cmd

import (
	"net"
	"os"
	"path/filepath"
	"strings"

	"git.parallelcoin.io/dev/9/pkg/util"
)

// CleanAndExpandPath expands environment variables and leading ~ in the passed path, cleans the result, and returns it.
func CleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(util.AppDataDir("pod", false))
		path = strings.Replace(path, "~", homeDir, 1)
	}
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, ".") {
		// explicitly prefix is this must be a relative path
		path = "./" + path
	}
	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%, but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// NormalizeAddress returns addr with the passed default port appended if there is not already a port specified.
func NormalizeAddress(addr, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}

// NormalizeAddresses returns a new slice with all the passed peer addresses normalized with the given default port, and all duplicates removed.
func NormalizeAddresses(addrs []string, defaultPort string) []string {
	for i, addr := range addrs {
		addrs[i] = NormalizeAddress(addr, defaultPort)
	}
	return RemoveDuplicateAddresses(addrs)
}

// RemoveDuplicateAddresses returns a new slice with all duplicate entries in addrs removed.
func RemoveDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

func intersection(a, b []string) (out []string) {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				out = append(out, x)
			}
		}
	}
	return
}

func uniq(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}
