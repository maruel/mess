package model

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestBotJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "db.json.zst")
	d, err := NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.BotCount(); l != 0 {
		t.Fatal(l)
	}
	want1 := getBot()
	d.BotSet(want1)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}

	if d, err = NewDBJSON(p); err != nil {
		t.Fatal(err)
	}
	if l := d.BotCount(); l != 1 {
		t.Fatal(l)
	}
	got := Bot{}
	d.BotGet("bot1", &got)
	want2 := *want1
	want2.LastSeen = time.Now().UTC()
	d.BotSet(&want2)
	all, _ := d.BotGetSlice("", 100)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want1, &got); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]Bot{want2}, all); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
}

func TestBotSQL(t *testing.T) {
	p := filepath.Join(t.TempDir(), "mess.db")
	d, err := NewDBSqlite3(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.BotCount(); l != 0 {
		t.Fatal(l)
	}
	want1 := getBot()
	d.BotSet(want1)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}

	if d, err = NewDBSqlite3(p); err != nil {
		t.Fatal(err)
	}
	if l := d.BotCount(); l != 1 {
		t.Fatal(l)
	}
	got := Bot{}
	d.BotGet("bot1", &got)
	want2 := *want1
	want2.LastSeen = time.Now().UTC().Round(time.Microsecond)
	d.BotSet(&want2)
	all, _ := d.BotGetSlice("", 100)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want1, &got); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]Bot{want2}, all); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
}

func TestBotNonZero(t *testing.T) {
	r := getBot()
	if err := isNonZero("", reflect.ValueOf(r)); err != nil {
		t.Fatal(err)
	}
	r.Dimensions["a"] = nil
	if err := isNonZero("", reflect.ValueOf(r)); err == nil || err.Error() != "*Bot.Dimensions[a] slice is empty" {
		t.Fatal(err)
	}
	r.Dimensions = nil
	if err := isNonZero("", reflect.ValueOf(r)); err == nil || err.Error() != "*Bot.Dimensions map is empty" {
		t.Fatal(err)
	}
}

func getBot() *Bot {
	return &Bot{
		Key:             "bot1",
		SchemaVersion:   1,
		Created:         time.Date(2020, 3, 13, 10, 9, 8, 7000, time.UTC),
		LastSeen:        time.Date(2020, 4, 13, 10, 9, 8, 7000, time.UTC),
		Version:         "botv1",
		AuthenticatedAs: "gcp1",
		Dimensions:      map[string][]string{"a": {"b", "c"}},
		State:           []byte(`{"python": "2.7"}`),
		ExternalIP:      "1.2.3.4",
		TaskID:          123,
		QuarantinedMsg:  "quarantined for real",
		MaintenanceMsg:  "very busy",
	}
}
