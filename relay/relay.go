package relay

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type (
	Relay struct {
		logger  *zap.Logger
		watcher *fsnotify.Watcher
		mapping map[string]string
	}
	Entry struct {
		Source      string `json:"src"`
		Destination string `json:"dst"`
	}
)

func New(logger *zap.Logger) (*Relay, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	return &Relay{
		logger:  logger,
		watcher: watcher,
		mapping: map[string]string{},
	}, nil
}

func (r *Relay) Run(entries []*Entry) error {
	for _, ent := range entries {
		if err := r.watcher.Add(ent.Source); err != nil {
			return fmt.Errorf("failed to watch file: %w", err)
		}

		r.mapping[ent.Source] = ent.Destination
		r.logger.Info("started to watch", zap.Any("entry", ent))
	}

	for {
		select {
		case err := <-r.watcher.Errors:
			return fmt.Errorf("an error occured while watching file: %w", err)

		case event := <-r.watcher.Events:
			r.logger.Info("received event", zap.Any("event", event))

			if event.Op&(fsnotify.Create|fsnotify.Write) > 0 {
				if err := r.sync(event.Name); err != nil {
					return fmt.Errorf("failed to sync file: %w", err)
				}

				r.logger.Info("synced", zap.Any("event", event))
			}
		}
	}
}

func (r *Relay) sync(src string) error {
	dst, ok := r.mapping[src]
	if !ok {
		return fmt.Errorf("no such src: %v", src)
	}

	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}
