package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/aalpern/go-filestats"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := cli.App("filestats", "File System Statistics")

	app.Command("stats", "Produce stats", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			walker := filestats.NewWalker(
				// filestats.WithDirectory("/Volumes/Photos-1/Photos"),
				// filestats.WithDirectory("/Volumes/Photos-1/Photos-Archive"),
				filestats.WithDirectory("/Data/Photos/Photos"),
			)
			stats, err := walker.Go()
			if err != nil {
				log.WithFields(log.Fields{
					"action": "walk",
					"status": "error",
					"error":  err,
				}).Error()
			} else {
				Dump(stats)
			}
		}
	})

	app.Run(os.Args)
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
	Count   int64   `json:"count"`
}

type D3Years []D3Year

func (a D3Years) Len() int           { return len(a) }
func (a D3Years) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a D3Years) Less(i, j int) bool { return a[i].Year < a[j].Year }

type D3YearDump struct {
	Key    string  `json:"key"`
	Values D3Years `json:"values"`
}

func Dump(s *filestats.FileStats) {
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
