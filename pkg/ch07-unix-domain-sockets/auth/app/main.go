package main

import (
	"flag"
	"fmt"
	"learn-network-programming/pkg/ch07-unix-domain-sockets/auth"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
)

// Func init
func init() {
	// Specify usage callback: print usage to stdcerr
	// (or other if redirected) and flag defaults
	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage:\n\t%s <group names>\n",
			filepath.Base(os.Args[0]),
		)
		flag.PrintDefaults()
	}
}

// Func parseGroupIds, creates map of group ids for given group names
// for each group name
func parseGroupIds(groupNames []string) map[string]struct{} {
	groupIds := make(map[string]struct{})

	for _, groupName := range groupNames {
		// lookup the group by its name
		group, err := user.LookupGroup(groupName)
		if err != nil {
			continue
		}

		// if found, add the id to the map
		groupIds[group.Gid] = struct{}{}
	}

	return groupIds
}

// Func main
func main() {
	// Parse command line options
	flag.Parse()

	// Parse group ids using all given non-flag command line options
	groups := parseGroupIds(flag.Args())
	// create a new socket file "creds.sock" in the temp dir
	socket := filepath.Join(os.TempDir(), "creds.sock")
	// resolve the socket filename into a net.Addr
	addr, err := net.ResolveUnixAddr("unix", socket)
	if err != nil {
		log.Fatal(err)
	}

	// Create a server on the resolved address
	server, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatal(err)
	}

	// Create a channel to listen for interrupts, to do proper cleanup on termination
	// capacity is 1 not to block a signal sender
	c := make(chan os.Signal, 1)
	// Redirect interrupt signal to the channel
	signal.Notify(c, os.Interrupt)
	// Listen asynchronously, close server on notification
	go func() {
		<-c
		log.Print("Interrupt sig received. Shutting down...")
		_ = server.Close()
	}()

	// Report litening start
	log.Printf("Listening on %s ...\n", socket)

	// Accept connections
	_ = acceptConnections(server, groups)
}

// Func accept connections
func acceptConnections(server *net.UnixListener, groups map[string]struct{}) error {
	// Accept new unix domain socket client
	client, err := server.AcceptUnix()
	if err != nil {
		return err
	}

	// Run client session asynchronously
	go clientSession(client, groups)

	// Keep accepting new clients
	return acceptConnections(server, groups)
}

// Func clientSession
func clientSession(client *net.UnixConn, groups map[string]struct{}) {
	// Close client on scope exit
	defer func() {
		_ = client.Close()
	}()

	// Additional wrap to handle errors
	err := doRunClientSession(client, groups)
	if err != nil {
		log.Printf("[%s] error: %v", client.RemoteAddr().String(), err)
	}
}

// Func doRunClientSession
func doRunClientSession(client *net.UnixConn, groups map[string]struct{}) error {
	// Check if the client has required permissions
	allowed, err := auth.Allowed(client, groups)
	if err != nil {
		return err
	}

	// Send welcome or access denied messages depending on the check
	if allowed {
		_, err = client.Write([]byte("Welcome\n"))
	} else {
		_, err = client.Write([]byte("Access denied\n"))
	}

	return err
}
