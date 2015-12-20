package store

import (
	"os"
	"testing"
)

func TestSqlitePutGetDelete(t *testing.T) {
	f := tempFilename()
	defer os.Remove(f)

	db := NewSqliteKVStore(f, false)
	db.Open()
	CommonTestPutGetDelete(t, db)
}

func BenchmarkSqlitePutGet(b *testing.B) {
	f := tempFilename()
	defer os.Remove(f)

	db := NewSqliteKVStore(f, false)
	db.Open()
	CommonBenchPutGet(b, db)
}
