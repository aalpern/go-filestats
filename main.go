package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := cli.App("filestats", "File System Statistics")

	app.Command("stats", "Produce stats", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			log.Info("Running stats")
			stats := NewStats()
			stats.Walk("/Volumes/Photos-1/Photos")
			stats.Walk("/Volumes/Photos-1/Photos-Archive")
			stats.Dump()
		}
	})

	app.Run(os.Args)
}

type FileStats struct {
	FileCount     int64
	DirCount      int64
	TotalSize     int64
	SizeByYear    map[int]int64
	SizeByYearDay map[int]map[int]int64
}

func NewStats() *FileStats {
	return &FileStats{
		SizeByYear:    map[int]int64{},
		SizeByYearDay: map[int]map[int]int64{},
	}
}

func (s *FileStats) getYearDayMap(year int) map[int]int64 {
	if m, ok := s.SizeByYearDay[year]; ok {
		return m
	} else {
		m = map[int]int64{}
		s.SizeByYearDay[year] = m
		return m
	}
}

func (s *FileStats) Add(info os.FileInfo) {
	if info.IsDir() {
		s.DirCount += 1
		return
	}
	size := info.Size()
	mtime := info.ModTime()
	s.FileCount += 1
	s.TotalSize += size
	s.SizeByYear[mtime.Year()] = s.SizeByYear[mtime.Year()] + size
	yd := s.getYearDayMap(mtime.Year())
	yd[mtime.YearDay()] = yd[mtime.YearDay()] + size
}

func (s *FileStats) Walk(dir string) error {
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
		s.Add(info)
		return nil
	})
	if err != nil {
		log.WithFields(log.Fields{
			"action": "Get",
			"status": "error",
			"error":  err,
		}).Error()
	}
	return err
}

func (s *FileStats) Dump() {
	log.WithFields(log.Fields{
		"action":           "Dump",
		"file_count":       s.FileCount,
		"dir_count":        s.DirCount,
		"total_size_bytes": s.TotalSize,
		"year_stats":       s.SizeByYear,
	}).Info()

	if js, err := json.Marshal(s); err != nil {
		log.WithFields(log.Fields{
			"action": "dump",
			"status": "json_error",
			"error":  err,
		}).Error("Error converting stats to JSON")
	} else {
		fmt.Println(string(js))
	}
}
