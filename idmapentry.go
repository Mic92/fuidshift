package fuidshift

import (
	"fmt"
	"strconv"
	"strings"
)

/*
 * One entry in id mapping set - a single range of either
 * uid or gid mappings.
 */
type IdmapEntry struct {
	Isuid    bool
	Isgid    bool
	Hostid   int // id as seen on the host - i.e. 100000
	Nsid     int // id as seen in the ns - i.e. 0
	Maprange int
}

/*
 * Shift a uid from the host into the container
 * I.e. 0 -> 1000 -> 101000
 */
func (e *IdmapEntry) shift_into_ns(id int) (int, error) {
	if id < e.Nsid || id >= e.Nsid+e.Maprange {
		// this mapping doesn't apply
		return 0, fmt.Errorf("N/A")
	}

	return id - e.Nsid + e.Hostid, nil
}

func is_between(x, low, high int) bool {
	return x >= low && x < high
}

func (e *IdmapEntry) Intersects(i IdmapEntry) bool {
	if (e.Isuid && i.Isuid) || (e.Isgid && i.Isgid) {
		switch {
		case is_between(e.Hostid, i.Hostid, i.Hostid+i.Maprange):
			return true
		case is_between(i.Hostid, e.Hostid, e.Hostid+e.Maprange):
			return true
		case is_between(e.Hostid+e.Maprange, i.Hostid, i.Hostid+i.Maprange):
			return true
		case is_between(i.Hostid+e.Maprange, e.Hostid, e.Hostid+e.Maprange):
			return true
		case is_between(e.Nsid, i.Nsid, i.Nsid+i.Maprange):
			return true
		case is_between(i.Nsid, e.Nsid, e.Nsid+e.Maprange):
			return true
		case is_between(e.Nsid+e.Maprange, i.Nsid, i.Nsid+i.Maprange):
			return true
		case is_between(i.Nsid+e.Maprange, e.Nsid, e.Nsid+e.Maprange):
			return true
		}
	}
	return false
}

/*
 * Shift a uid from the container back to the host
 * I.e. 101000 -> 1000
 */
func (e *IdmapEntry) shift_from_ns(id int) (int, error) {
	if id < e.Hostid || id >= e.Hostid+e.Maprange {
		// this mapping doesn't apply
		return 0, fmt.Errorf("N/A")
	}

	return id - e.Hostid + e.Nsid, nil
}

func (e *IdmapEntry) parse(s string) error {
	split := strings.Split(s, ":")
	var err error
	if len(split) != 4 {
		return fmt.Errorf("Bad idmap: %q", s)
	}
	switch split[0] {
	case "u":
		e.Isuid = true
	case "g":
		e.Isgid = true
	case "b":
		e.Isuid = true
		e.Isgid = true
	default:
		return fmt.Errorf("Bad idmap type in %q", s)
	}
	e.Nsid, err = strconv.Atoi(split[1])
	if err != nil {
		return err
	}
	e.Hostid, err = strconv.Atoi(split[2])
	if err != nil {
		return err
	}
	e.Maprange, err = strconv.Atoi(split[3])
	if err != nil {
		return err
	}

	// wraparound
	if e.Hostid+e.Maprange < e.Hostid || e.Nsid+e.Maprange < e.Nsid {
		return fmt.Errorf("Bad mapping: id wraparound")
	}

	return nil
}

/* taken from http://blog.golang.org/slices (which is under BSD licence) */
func Extend(slice []IdmapEntry, element IdmapEntry) []IdmapEntry {
	n := len(slice)
	if n == cap(slice) {
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]IdmapEntry, len(slice), 2*len(slice)+1)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}
