package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	year := mtime.Year()
	if year < 2002 {
		year = 2002
	}
	s.FileCount += 1
	s.TotalSize += size
	s.SizeByYear[year] = s.SizeByYear[year] + size
	yd := s.getYearDayMap(year)
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

// ----------------------------------------------------------------------
// Dumping to D3 datum compatible format for graphing
// ----------------------------------------------------------------------

type D3Year struct {
	Year    int     `json:"year"`
	Bytes   int64   `json:"bytes"`
	KB      float64 `json:"kb"`
	MB      float64 `json:"mb"`
	GB      float64 `json:"gb"`
	GBTotal float64 `json:"gbtotal"`
}

type D3Years []D3Year

func (a D3Years) Len() int           { return len(a) }
func (a D3Years) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a D3Years) Less(i, j int) bool { return a[i].Year < a[j].Year }

type D3YearDump struct {
	Key    string  `json:"key"`
	Values D3Years `json:"values"`
}

func (s *FileStats) Dump() {
	var d3 D3YearDump
	for year, bytes := range s.SizeByYear {
		d3.Values = append(d3.Values, D3Year{
			Year:  year,
			Bytes: bytes,
			KB:    float64(bytes) / 1024,
			MB:    float64(bytes) / (1024 * 1024),
			GB:    float64(bytes) / (1024 * 1024 * 1024),
		})
	}
	sort.Sort(d3.Values)
	for i, _ := range d3.Values {
		d3.Values[i].GBTotal = d3.Values[i].GB
		if i > 0 {
			d3.Values[i].GBTotal += d3.Values[i-1].GBTotal
		}
	}

	if js, err := json.MarshalIndent([]D3YearDump{d3}, "", "  "); err != nil {
		log.WithFields(log.Fields{
			"action": "dump",
			"status": "json_error",
			"error":  err,
		}).Error("Error converting stats to JSON")
	} else {
		fmt.Println(string(js))
	}
}
