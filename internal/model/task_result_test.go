package model

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTaskResultJSON(t *testing.T) {
	r := getTaskResult()
	p := filepath.Join(t.TempDir(), "db.json.zst")
	// TODO(maruel): iterate recursively and check for reflect.Value.IsZero().
	d, err := NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskResultCount(); l != 0 {
		t.Fatal(l)
	}
	d.TaskResultSet(r)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskResultCount(); l != 1 {
		t.Fatal(l)
	}
	got := TaskResult{}
	d.TaskResultGet(2, &got)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(r, &got); diff != "" {
		t.Fatal(diff)
	}
}

func getTaskResult() *TaskResult {
	return &TaskResult{
		SchemaVersion: 1,
		Key:           2,
		BotID:         "bot1",
		Blob: TaskResultBlob{
			BotVersion:     "version1",
			BotDimension:   map[string][]string{"a": {"b", "c"}},
			BotIdleSince:   time.Hour,
			ServerVersions: []string{"v1"},
			TaskSlice:      1,
			DedupedFrom:    2134,
			PropertiesHash: "abc",
			TaskOutput: TaskOutput{
				Size: 1000,
			},
			ExitCode: 128,
			State:    Killed,
			Children: []int64{3, 4},
			Output: Digest{
				Size: 101,
				Hash: [32]byte{},
			},
			CIPDPins: []CIPDPackage{
				{
					PkgName: "pkg1",
					Version: "pkgv",
					Path:    "pkgp",
				},
			},
			ResultDB: ResultDB{
				Host:       "rdbh",
				Invocation: "rdbi",
			},
			Duration:  time.Millisecond,
			Started:   time.Date(2020, 1, 13, 10, 9, 8, 7, time.UTC),
			Completed: time.Date(2020, 2, 13, 10, 9, 8, 7, time.UTC),
			Abandoned: time.Date(2020, 3, 13, 10, 9, 8, 7, time.UTC),
			Modified:  time.Date(2020, 4, 13, 10, 9, 8, 7, time.UTC),
			Cost:      100.2,
			Killing:   true,
			DeadAfter: time.Date(2020, 5, 13, 10, 9, 8, 7, time.UTC),
		},
	}
}
