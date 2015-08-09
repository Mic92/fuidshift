package fuidshift_test

import (
	"fmt"
	"github.com/Mic92/fuidshift"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func ok(t testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		t.FailNow()
	}
}

func assertOwnership(t *testing.T, path string, uid, gid int) {
	var stat syscall.Stat_t
	ok(t, syscall.Lstat(path, &stat))
	gotUid := int(stat.Uid)
	gotGid := int(stat.Gid)
	if gotUid != uid || gotGid != gid {
		t.Errorf("expected '%s' to have uid/gid %d:%d, got: %d:%d", path, uid, gid, gotUid, gotGid)
	}
}

func TestUidshift(t *testing.T) {
	if os.Getuid() != 0 {
		t.Fatal("Tests needs to be run as root")
	}
	idmap := fuidshift.IdmapSet{}
	idmap, err := idmap.Append("b:0:100000:65536")
	ok(t, err)

	tempdir, err := ioutil.TempDir(os.TempDir(), "fuidshift")
	ok(t, err)
	defer os.Remove(tempdir)

	dir := path.Join(tempdir, "dir")
	ok(t, os.Mkdir(dir, 0700))
	ok(t, os.Chown(dir, 1, 1))

	file := path.Join(tempdir, "file")
	ok(t, ioutil.WriteFile(file, []byte("hello\ngo\n"), 0700))
	ok(t, os.Chown(file, 0, 0))

	ok(t, idmap.UidshiftIntoContainer(tempdir, false))
	assertOwnership(t, dir, 100001, 100001)
	assertOwnership(t, file, 100000, 100000)
	ok(t, idmap.UidshiftFromContainer(tempdir, false))
	assertOwnership(t, dir, 1, 1)
	assertOwnership(t, file, 0, 0)
}
