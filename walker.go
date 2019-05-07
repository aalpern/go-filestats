package filestats

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Walker struct {
	dirs []string
}

type WalkerOpt func(w *Walker)

func WithDirectory(dir string) WalkerOpt {
	return func(w *Walker) {
		w.dirs = append(w.dirs, dir)
	}
}

func NewWalker(opts ...WalkerOpt) *Walker {
	w := &Walker{}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (s *Walker) Go() (*FileStats, error) {
	stats := NewStats()
	for _, dir := range s.dirs {
		if err := s.walk(stats, dir); err != nil {
			return nil, err
		}
	}
	return stats, nil
}

func (s *Walker) walk(stats *FileStats, dir string) error {
	// For each entry in the directory tree
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.WithFields(log.Fields{
				"action": "walk",
				"status": "error",
				"error":  err,
				"path":   path,
			}).Error()
			return err
		}
		stats.Add(info)
		return nil
	})
	return err
}
