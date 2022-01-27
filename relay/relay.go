package relay

import (
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type (
	Entry struct {
		Source      string `json:"src"`
		Destination string `json:"dst"`
	}
	Relay struct {
		logger  *zap.Logger
		watcher *fsnotify.Watcher

		srcDefn map[string]*SourceDefinition
	}
	SourceDefinition struct {
		entry       *Entry
		sourceMtime time.Time
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
		srcDefn: map[string]*SourceDefinition{},
	}, nil
}

func (r *Relay) Run(entries []*Entry) error {
	for _, entry := range entries {
		srcFileInfo, err := os.Stat(entry.Source)
		if err != nil {
			return fmt.Errorf("failed to stat source: %w", err)
		}
		if err := r.watcher.Add(entry.Source); err != nil {
			return fmt.Errorf("failed to watch source: %w", err)
		}

		r.srcDefn[entry.Source] = &SourceDefinition{
			entry:       entry,
			sourceMtime: srcFileInfo.ModTime(),
		}
		r.logger.Info("started to watch", zap.Any("entry", entry))
	}
	notify("Bucketrelay started!")

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
			}
		}
	}
}

func (r *Relay) sync(src string) error {
	defn, ok := r.srcDefn[src]
	if !ok {
		return fmt.Errorf("no such src: %v", src)
	}
	defer func() {
		defn.sourceMtime = time.Now()
	}()

	if dstFileInfo, err := os.Stat(defn.entry.Destination); err != nil {
		r.logger.Warn("failed to stat dst file", zap.Error(err), zap.Any("entry", defn.entry))
	} else if dstFileInfo.ModTime().After(defn.sourceMtime) {
		r.watcher.Remove(defn.entry.Source)
		defer r.watcher.Add(defn.entry.Source)

		if err := copyFile(defn.entry.Destination, defn.entry.Source); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}

		notify(fmt.Sprintf("Synced backword: %v", defn.entry.Source))
		r.logger.Info("synced backword", zap.Any("entry", defn.entry))
		return nil
	}

	if err := copyFile(defn.entry.Source, defn.entry.Destination); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	r.logger.Info("synced", zap.Any("entry", defn.entry))
	return nil
}
