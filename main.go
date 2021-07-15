package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/kaz/bucketrelay/relay"
)

func run() error {
	config := []*relay.Entry{}
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %v <JSON formatted %v>", os.Args[0], reflect.TypeOf(config))
	}
	if err := json.Unmarshal([]byte(os.Args[1]), &config); err != nil {
		return fmt.Errorf("failed unmarshal json: %w", err)
	}

	r, err := relay.New()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	return r.Run(config)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
