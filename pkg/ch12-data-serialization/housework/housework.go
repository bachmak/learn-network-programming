package housework

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// struct Chore
type Chore struct {
	Complete    bool
	Description string
}

// Chores
type Chores []*Chore

// func Add
func (chores *Chores) Add(chore Chore) {
	*chores = append(*chores, &chore)
}

// func SetComplete
func (chores *Chores) SetComplete(idx int, complete bool) error {
	// check if index is valid
	if idx < 0 || idx >= len(*chores) {
		return fmt.Errorf("chore %d not found", idx)
	}

	// set chore completion
	chore := []*Chore(*chores)[idx]
	chore.Complete = complete

	return nil
}

// load and flush function aliases
type (
	LoadFunc  func(io.Reader) (Chores, error)
	FlushFunc func(io.Writer, Chores) error
)

// func LoadFromFile
func LoadFromFile(filename string, load LoadFunc) (Chores, error) {
	// get file info just to check if a file exists
	_, err := os.Stat(filename)
	// return empty chore list if it doesn't
	if os.IsNotExist(err) {
		return make(Chores, 0), nil
	}

	// open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	// close file at scope exit
	defer func() {
		_ = file.Close()
	}()

	// load chores using the provided function
	chores, err := load(file)
	if err != nil {
		return nil, err
	}

	return chores, nil
}

// func FlushToFile
func FlushToFile(filename string, chores Chores, flush FlushFunc) error {
	// create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	// close file at scope exit
	defer func() {
		_ = file.Close()
	}()

	// flush chores using the provided function
	err = flush(file, chores)
	if err != nil {
		return err
	}

	return nil
}

// func LoadJson
func LoadJson(reader io.Reader) (Chores, error) {
	// create an empty chore list and a decoder
	var chores Chores
	decoder := json.NewDecoder(reader)

	// decode chores
	err := decoder.Decode(&chores)
	if err != nil {
		return nil, err
	}

	return chores, nil
}

// func FlushJson
func FlushJson(writer io.Writer, chores Chores) error {
	// create encoder
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	// encode chores
	err := encoder.Encode(chores)
	if err != nil {
		return err
	}

	return nil
}

// func LoadGob (similar to LoadJson)
func LoadGob(reader io.Reader) (Chores, error) {
	var chores Chores
	decoder := gob.NewDecoder(reader)

	err := decoder.Decode(&chores)
	if err != nil {
		return nil, err
	}

	return chores, nil
}

// func FlushGob (similar to FlushJson)
func FlushGob(writer io.Writer, chores Chores) error {
	encoder := gob.NewEncoder(writer)

	err := encoder.Encode(chores)
	if err != nil {
		return err
	}

	return nil
}
