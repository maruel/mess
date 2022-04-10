package model

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTaskResultJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "db.json.zst")
	// TODO(maruel): iterate recursively and check for reflect.Value.IsZero().
	d, err := NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskResultCount(); l != 0 {
		t.Fatal(l)
	}
	want1 := getTaskResult()
	d.TaskResultSet(want1)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}

	if d, err = NewDBJSON(p); err != nil {
		t.Fatal(err)
	}
	if l := d.TaskResultCount(); l != 1 {
		t.Fatal(l)
	}
	got1 := TaskResult{}
	d.TaskResultGet(2, &got1)
	want2 := *want1
	want2.BotIdleSince += time.Minute
	d.TaskResultSet(&want2)
	got2 := TaskResult{}
	d.TaskResultGet(2, &got2)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want1, &got1); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want2, got2); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
}

func TestTaskResultSQL(t *testing.T) {
	p := filepath.Join(t.TempDir(), "mess.db")
	d, err := NewDBSqlite3(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskResultCount(); l != 0 {
		t.Fatal(l)
	}
	want1 := getTaskResult()
	d.TaskResultSet(want1)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}

	if d, err = NewDBSqlite3(p); err != nil {
		t.Fatal(err)
	}
	if l := d.TaskResultCount(); l != 1 {
		t.Fatal(l)
	}
	got1 := TaskResult{}
	d.TaskResultGet(2, &got1)
	want2 := *want1
	want2.BotIdleSince += time.Minute
	d.TaskResultSet(&want2)
	got2 := TaskResult{}
	d.TaskResultGet(2, &got2)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want1, &got1); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want2, got2); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
}

func TestTaskResultNonZero(t *testing.T) {
	r := getTaskResult()
	if err := isNonZero("", reflect.ValueOf(r)); err != nil {
		t.Fatal(err)
	}
	r.Output.Hash[0] = 0
	if err := isNonZero("", reflect.ValueOf(r)); err == nil || err.Error() != "*TaskResult.Output.Hash value is zero" {
		t.Fatal(err)
	}
}

func getTaskResult() *TaskResult {
	return &TaskResult{
		SchemaVersion:    1,
		Key:              2,
		BotID:            "bot1",
		BotVersion:       "version1",
		BotDimensions:    map[string][]string{"a": {"b", "c"}},
		BotIdleSince:     time.Hour,
		ServerVersions:   []string{"v1"},
		CurrentTaskSlice: 1,
		DedupedFrom:      2134,
		PropertiesHash:   "abc",
		TaskOutput: TaskOutput{
			Size: 1000,
		},
		ExitCode:        128,
		InternalFailure: "blew up",
		State:           Killed,
		Children:        []int64{3, 4},
		Output: Digest{
			Size: 101,
			Hash: [32]byte{1},
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
		Started:   time.Date(2020, 1, 13, 10, 9, 8, 7000, time.UTC),
		Completed: time.Date(2020, 2, 13, 10, 9, 8, 7000, time.UTC),
		Abandoned: time.Date(2020, 3, 13, 10, 9, 8, 7000, time.UTC),
		Modified:  time.Date(2020, 4, 13, 10, 9, 8, 7000, time.UTC),
		Cost:      100.2,
		Killing:   true,
		DeadAfter: time.Date(2020, 5, 13, 10, 9, 8, 7000, time.UTC),
	}
}
