package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/kaz/bucketrelay/relay"
	"go.uber.org/zap"
)

func run() error {
	config := []*relay.Entry{}
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %v <JSON formatted %v>", os.Args[0], reflect.TypeOf(config))
	}
	if err := json.Unmarshal([]byte(os.Args[1]), &config); err != nil {
		return fmt.Errorf("failed unmarshal json: %w", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	r, err := relay.New(logger)
	if err != nil {
		logger.Error("failed to initialize", zap.Error(err))
		return fmt.Errorf("failed to initialize: %w", err)
	}

	if err := r.Run(config); err != nil {
		logger.Error("failed to run", zap.Error(err))
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
