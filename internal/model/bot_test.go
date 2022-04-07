package model

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestBotJSON(t *testing.T) {
	b := getBot()
	p := filepath.Join(t.TempDir(), "db.json.zst")
	// TODO(maruel): iterate recursively and check for reflect.Value.IsZero().
	d, err := NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.BotCount(); l != 0 {
		t.Fatal(l)
	}
	d.BotSet(b)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.BotCount(); l != 1 {
		t.Fatal(l)
	}
	got := Bot{}
	d.BotGet("bot1", &got)
	all := d.BotGetAll(nil)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(b, &got); diff != "" {
		t.Fatal(diff)
	}
	if diff := cmp.Diff([]Bot{*b}, all); diff != "" {
		t.Fatal(diff)
	}
}

func getBot() *Bot {
	return &Bot{
		SchemaVersion: 1,
		Key:           "bot1",
		Create:        time.Date(2020, 3, 13, 10, 9, 8, 7, time.UTC),
		LastSeen:      time.Date(2020, 4, 13, 10, 9, 8, 7, time.UTC),
		Version:       "botv1",
		Blob: BotBlob{
			Dimensions: map[string][]string{"a": {"b", "c"}},
			Events: []*BotEvent{
				{
					SchemaVersion: 1,
					Key:           3,
					BotID:         "bot1",
					Time:          time.Date(2020, 5, 13, 10, 9, 8, 7, time.UTC),
					Blob: BotEventBlob{
						Event:           "bot_update",
						Message:         "msg",
						Dimensions:      map[string][]string{"e": {"f", "g"}},
						State:           "{}",
						ExternalIP:      "1.2.3.4",
						AuthenticatedAs: "auth1",
						Version:         "v1",
						QuarantinedMsg:  "quarantined",
						MaintenanceMsg:  "maintenance",
						TaskID:          "task1",
					},
				},
			},
		},
	}
}
