package main

import (
	"context"
	"flag"
	"fmt"
	hw "learn-network-programming/housework/v1"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CLI options
var addr string

// func init
func init() {
	// init CLI options
	flag.StringVar(&addr, "address", "localhost:34443", "server address")

	// print usage
	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			`Usage: %s [flags] [add chore, ...|complete #]
      add       add comma-separated chores
      complete  complete designated chore

      Flags:
      `,
			filepath.Base(os.Args[0]),
		)
		flag.PrintDefaults()
	}
}

// func main
func main() {
	// parse CLI arguments
	flag.Parse()

	// use insecure connection
	creds := insecure.NewCredentials()
	withCreds := grpc.WithTransportCredentials(creds)

	// create new client connecting to the specified address
	client, err := grpc.NewClient(addr, withCreds)
	if err != nil {
		log.Fatal(err)
	}

	// create client-side service representation
	robot := hw.NewRobotMaidClient(client)
	// empty context
	ctx := context.Background()

	// parse cmd
	cmd := strings.ToLower(flag.Arg(0))
	// select action
	switch cmd {
	case "add":
		// add chores
		err = add(ctx, robot, strings.Join(flag.Args()[1:], " "))
	case "complete":
		// complete chore
		err = complete(ctx, robot, flag.Arg(1))
	}

	if err != nil {
		log.Fatal(err)
	}

	// list chores
	err = list(ctx, robot)
	if err != nil {
		log.Fatal(err)
	}
}

// func add
func add(ctx context.Context, robot hw.RobotMaidClient, s string) error {
	// new chores
	chores := new(hw.Chores)

	// iterate over splitted chores
	for _, chore := range strings.Split(s, ",") {
		// add non-empty chores to the list
		if desc := strings.TrimSpace(chore); desc != "" {
			chores.Chores = append(
				chores.Chores,
				&hw.Chore{
					Description: desc,
				},
			)
		}
	}

	// is the list isn't empty, make an RPC call to add the new chores
	if len(chores.Chores) > 0 {
		_, err := robot.Add(ctx, chores)
		if err != nil {
			return err
		}
	}

	return nil
}

// func complete
func complete(ctx context.Context, robot hw.RobotMaidClient, s string) error {
	// convert string to index
	idx, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	// index from zero
	idx -= 1
	// create complete request
	req := hw.CompleteRequest{ChoreNumber: int32(idx)}

	// make an RPC call
	_, err = robot.Complete(ctx, &req)
	if err != nil {
		return err
	}

	return nil
}

// func list
func list(ctx context.Context, robot hw.RobotMaidClient) error {
	// make an RPC call
	chores, err := robot.List(ctx, &hw.Empty{})
	if err != nil {
		return err
	}

	// print header
	fmt.Printf("#\t[X]\tDescription\n")

	// print chores
	for i, chore := range chores.Chores {
		c := " "
		if chore.Complete {
			c = "X"
		}

		fmt.Printf(
			"%d\t[%s]\t%s\n",
			i+1,
			c,
			chore.Description,
		)
	}

	return nil
}
