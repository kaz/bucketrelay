package relay

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type (
	Relay struct {
		watcher *fsnotify.Watcher
		mapping map[string]string
	}
	Entry struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	}
)

func New() (*Relay, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	return &Relay{
		watcher: watcher,
		mapping: map[string]string{},
	}, nil
}

func (r *Relay) Run(entries []*Entry) error {
	for _, ent := range entries {
		if err := r.watcher.Add(ent.Src); err != nil {
			return fmt.Errorf("failed to watch file: %w", err)
		}

		r.mapping[ent.Src] = ent.Dst
		log.Printf("\"%v\": WATCHING\n", ent.Src)
	}

	for {
		select {
		case err := <-r.watcher.Errors:
			return fmt.Errorf("an error occured while watching file: %w", err)
		case event := <-r.watcher.Events:
			log.Println(event)
			if event.Op&(fsnotify.Create|fsnotify.Write) > 0 {
				if err := r.sync(event.Name); err != nil {
					return fmt.Errorf("failed to sync file: %w", err)
				}
			}
		}
	}
}

func (r *Relay) sync(src string) error {
	dst, ok := r.mapping[src]
	if !ok {
		return fmt.Errorf("no such src: %v", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open src file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create dst file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	log.Printf("\"%v\" -> \"%v\": SYNCED\n", src, dst)
	return nil
}
