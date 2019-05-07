package filestats

import (
	"os"
)

type FileStats struct {
	FileCount     int64
	DirCount      int64
	TotalSize     int64
	SizeByYear    map[int]int64
	CountByYear   map[int]int64
	SizeByYearDay map[int]map[int]int64
}

func NewStats() *FileStats {
	return &FileStats{
		SizeByYear:    map[int]int64{},
		CountByYear:   map[int]int64{},
		SizeByYearDay: map[int]map[int]int64{},
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
	s.CountByYear[year] = s.CountByYear[year] + 1
	yd := s.getYearDayMap(year)
	yd[mtime.YearDay()] = yd[mtime.YearDay()] + size
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
