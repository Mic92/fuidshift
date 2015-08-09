package fuidshift

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func (m IdmapSet) Intersects(i IdmapEntry) bool {
	for _, e := range m.Idmap {
		if i.Intersects(e) {
			return true
		}
	}
	return false
}

func (m IdmapSet) Append(s string) (IdmapSet, error) {
	e := IdmapEntry{}
	err := e.parse(s)
	if err != nil {
		return m, err
	}
	if m.Intersects(e) {
		return m, fmt.Errorf("Conflicting id mapping")
	}
	m.Idmap = Extend(m.Idmap, e)
	return m, nil
}

type IdmapSet struct {
	Idmap []IdmapEntry
}

func (m IdmapSet) Len() int {
	return len(m.Idmap)
}

func (m IdmapSet) doShiftIntoNs(uid int, gid int, how string) (int, int) {
	u := -1
	g := -1
	for _, e := range m.Idmap {
		var err error
		var tmpu, tmpg int
		if e.Isuid && u == -1 {
			switch how {
			case "in":
				tmpu, err = e.shift_into_ns(uid)
			case "out":
				tmpu, err = e.shift_from_ns(uid)
			}
			if err == nil {
				u = tmpu
			}
		}
		if e.Isgid && g == -1 {
			switch how {
			case "in":
				tmpg, err = e.shift_into_ns(gid)
			case "out":
				tmpg, err = e.shift_from_ns(gid)
			}
			if err == nil {
				g = tmpg
			}
		}
	}

	return u, g
}

func (m IdmapSet) ShiftIntoNs(uid int, gid int) (int, int) {
	return m.doShiftIntoNs(uid, gid, "in")
}

func (m IdmapSet) ShiftFromNs(uid int, gid int) (int, int) {
	return m.doShiftIntoNs(uid, gid, "out")
}

func (set *IdmapSet) doUidshiftIntoContainer(dir string, testmode bool, how string) error {
	convert := func(path string, fi os.FileInfo, err error) (e error) {
		uid, gid, err := getOwner(path)
		if err != nil {
			return err
		}
		var newuid, newgid int
		switch how {
		case "in":
			newuid, newgid = set.ShiftIntoNs(uid, gid)
		case "out":
			newuid, newgid = set.ShiftFromNs(uid, gid)
		}
		if testmode {
			fmt.Printf("I would shift %q to %d %d\n", path, newuid, newgid)
		} else {
			err = os.Lchown(path, int(newuid), int(newgid))
			if err == nil {
				m := fi.Mode()
				if m&os.ModeSymlink == 0 {
					err = os.Chmod(path, m)
					if err != nil {
						fmt.Printf("Error resetting mode on %q, continuing\n", path)
					}
				}
			}
		}
		return nil
	}

	if !pathExists(dir) {
		return fmt.Errorf("No such file or directory: %q", dir)
	}
	return filepath.Walk(dir, convert)
}

func (set *IdmapSet) UidshiftIntoContainer(dir string, testmode bool) error {
	return set.doUidshiftIntoContainer(dir, testmode, "in")
}

func (set *IdmapSet) UidshiftFromContainer(dir string, testmode bool) error {
	return set.doUidshiftIntoContainer(dir, testmode, "out")
}

func getOwner(path string) (int, int, error) {
	var stat syscall.Stat_t
	err := syscall.Lstat(path, &stat)
	return int(stat.Uid), int(stat.Gid), err
}

func pathExists(name string) bool {
	_, err := os.Lstat(name)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
