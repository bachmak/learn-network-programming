package auth

import (
	"fmt"
	"net"
	"os"
	"os/user"

	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

// Func Allowed, check if a connection peer belongs to one of the allowed groups
func Allowed(conn *net.UnixConn, groups map[string]struct{}) (bool, error) {
	// If no connection of map is absent or empty
	if conn == nil || groups == nil || len(groups) == 0 {
		return false, nil
	}

	// Get file associated with connection (socket)
	file, err := conn.File()
	if err != nil {
		return false, err
	}
	// Close file at the scope exit
	defer func() { _ = file.Close() }()

	// Get peer credentials (from the other side of the socket)
	usrCred, err := getPeerCreds(file)
	if err != nil {
		return false, err
	}

	// Find user by its id
	uid := fmt.Sprintf("%d", usrCred.Uid)
	usr, err := user.LookupId(uid)
	if err != nil {
		return false, err
	}

	// Get all group ids the user belongs to
	usrGroups, err := usr.GroupIds()
	if err != nil {
		return false, err
	}

	// Check if it belongs to one of the allowed groups
	allowed := slices.ContainsFunc(usrGroups, func(gid string) bool {
		_, found := groups[gid]
		return found
	})

	return allowed, nil
}

// Func get peer credentials
func getPeerCreds(file *os.File) (*unix.Ucred, error) {
	// We're going to do a syscall, so no files - only descriptors
	fd := int(file.Fd())
	// Retrieve the creds: for the given file, at the socket level, peer credentials
	usrCred, err := unix.GetsockoptUcred(fd, unix.SOL_SOCKET, unix.SO_PEERCRED)
	// Return if okay
	if err == nil {
		return usrCred, nil
	}

	// Try again if interrupt, return error otherwise
	if isPersistent := err != unix.EINTR; isPersistent {
		return nil, err
	}
	return getPeerCreds(file)
}
