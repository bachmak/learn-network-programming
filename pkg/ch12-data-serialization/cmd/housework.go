package main

import (
	"flag"
	"fmt"
	hw "learn-network-programming/pkg/ch12-data-serialization/housework"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// data file name
var dataFile string

// func init
func init() {
	// init data file CLI option
	flag.StringVar(
		&dataFile,
		"file",
		"housework.json",
		"data file",
	)
	// custom usage function
	flag.Usage = func() {
		// show custom usage
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [flags] [add chore, ...|complete #]
      add       add comma-separated chores
      complete  complete deisgnated chore

      Flags:`,
			filepath.Base(os.Args[0]),
		)
		// show default usage
		flag.PrintDefaults()
	}
}

func main() {
	// parse command line arguments
	flag.Parse()

	// select load + flush functions
	load, flush := getLoadAndFlush(dataFile)

	// load chores from file
	chores, err := hw.LoadFromFile(dataFile, load)
	if err != nil {
		log.Fatal(err)
	}

	// parse the command
	cmd := strings.ToLower(flag.Arg(0))
	switch cmd {
	case "add":
		// add chores to the list
		add(chores, parseDescriptions(flag.Args()[1:]))
	case "complete":
		// complete selected chore
		err = complete(chores, flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
	}

	// show all chores
	list(chores)

	// flush chores to file
	err = hw.FlushToFile(dataFile, chores, flush)
	if err != nil {
		log.Fatal(err)
	}
}

// func getLoadAndFlush
func getLoadAndFlush(dataFile string) (hw.LoadFunc, hw.FlushFunc) {
	// select appropriate load and flush function pait depending on file extension
	ext := filepath.Ext(dataFile)
	switch ext {
	case ".json":
		return hw.LoadJson, hw.FlushJson
	case ".dat", ".gob", ".bin":
		return hw.LoadGob, hw.FlushGob
	case ".pb":
		return hw.LoadProto, hw.FlushProto
	}

	return hw.LoadJson, hw.FlushJson
}

// func parseDescriptions
func parseDescriptions(descStrs []string) []string {
	// join description strings
	descsJoined := strings.Join(descStrs, " ")
	// split comma-separated descriptions
	descs := strings.Split(descsJoined, ",")
	var res []string

	// add non-empty descriptions to result
	for _, desc := range descs {
		desc = strings.TrimSpace(desc)
		if desc != "" {
			res = append(res, desc)
		}
	}

	return res
}

// func add
func add(chores *hw.Chores, descs []string) {
	// add new chores for all descriptions (not completed by default)
	for _, desc := range descs {
		chore := &hw.Chore{
			Complete:    false,
			Description: desc,
		}

		chores.Add(chore)
	}
}

// func complete
func complete(chores *hw.Chores, idxStr string) error {
	// convert index string to number
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return err
	}

	// set complete status as true (index from zero)
	return chores.SetComplete(idx-1, true)
}

// func list
func list(chores *hw.Chores) {
	// show header
	fmt.Printf("#\t[X]\tDescription\n")

	// show chores
	for idx, chore := range chores.Chores {
		// helper function to show completion status
		complete := func() string {
			if chore.Complete {
				return "X"
			}

			return " "
		}()

		// index, completion, description
		fmt.Printf(
			"%d\t[%s]\t%s\n",
			idx+1,
			complete,
			chore.Description,
		)
	}
}
