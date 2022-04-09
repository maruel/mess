package model

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestBotEventJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "db.json.zst")
	d, err := NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	want1 := getBotEvent()
	d.BotEventAdd(want1)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}

	if d, err = NewDBJSON(p); err != nil {
		t.Fatal(err)
	}
	want2 := getBotEvent()
	want2.Message = "message 2"
	d.BotEventAdd(want2)
	f := Filter{Limit: 100}
	all, cursor := d.BotEventGetSlice("bot1", f)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]BotEvent{*want2, *want1}, all); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if cursor != "" {
		t.Fatal(cursor)
	}
}

func TestBotEventSQL(t *testing.T) {
	p := filepath.Join(t.TempDir(), "mess.db")
	d, err := NewDBSqlite3(p)
	if err != nil {
		t.Fatal(err)
	}
	want1 := getBotEvent()
	d.BotEventAdd(want1)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}

	if d, err = NewDBSqlite3(p); err != nil {
		t.Fatal(err)
	}
	want2 := getBotEvent()
	want2.Message = "message 2"
	d.BotEventAdd(want2)
	f := Filter{Limit: 100}
	all, cursor := d.BotEventGetSlice("bot1", f)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]BotEvent{*want2, *want1}, all); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if cursor != "" {
		t.Fatal(cursor)
	}
}

func TestBotEventNonZero(t *testing.T) {
	r := getBotEvent()
	r.Key = 1
	if err := isNonZero("", reflect.ValueOf(r)); err != nil {
		t.Fatal(err)
	}
	r.Dimensions["a"] = nil
	if err := isNonZero("", reflect.ValueOf(r)); err == nil || err.Error() != "*BotEvent.Dimensions[a] slice is empty" {
		t.Fatal(err)
	}
	r.Dimensions = nil
	if err := isNonZero("", reflect.ValueOf(r)); err == nil || err.Error() != "*BotEvent.Dimensions map is empty" {
		t.Fatal(err)
	}
}

func getBotEvent() *BotEvent {
	b := getBot()
	e := &BotEvent{}
	now := time.Date(2020, 3, 13, 10, 9, 8, 7000, time.UTC)
	e.InitFrom(b, now, "event1", "msg1")
	return e
}
