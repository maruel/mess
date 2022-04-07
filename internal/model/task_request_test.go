package model

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTaskRequestJSON(t *testing.T) {
	r := getTaskRequest()
	p := filepath.Join(t.TempDir(), "db.json.zst")
	// TODO(maruel): iterate recursively and check for reflect.Value.IsZero().
	d, err := NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskRequestCount(); l != 0 {
		t.Fatal(l)
	}
	d.TaskRequestSet(r)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskRequestCount(); l != 1 {
		t.Fatal(l)
	}
	got := TaskRequest{}
	d.TaskRequestGet(2, &got)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(r, &got); diff != "" {
		t.Fatal(diff)
	}
}

func getTaskRequest() *TaskRequest {
	return &TaskRequest{
		SchemaVersion: 1,
		Key:           2,
		Created:       time.Date(2020, 3, 13, 10, 9, 8, 7, time.UTC),
		Priority:      3,
		ParentTask:    4,
		Tags: []string{
			"a:b", "c:d",
		},
		Blob: TaskRequestBlob{
			TaskSlices: []TaskSlice{
				{
					Properties: TaskProperties{
						Caches: []Cache{
							{
								Name: "cname",
								Path: "cpath",
							},
						},
						Command:    []string{"echo", "hi"},
						RelativeWD: ".",
						CASHost:    "rbe",
						Input: Digest{
							Size: 10,
							Hash: [32]byte{},
						},
						CIPDHost:     "chrome-package",
						CIPDPackages: []CIPDPackage{},
						Dimensions:   map[string]string{"os": "Windows"},
						Env:          map[string]string{"FOO": "bar"},
						EnvPrefixes:  map[string]string{"PATH": "./foo"},
						HardTimeout:  time.Minute,
						GracePeriod:  time.Second,
						IOTimeout:    2 * time.Minute,
						Idempotent:   true,
						Outputs:      []string{"foo"},
						Containment: Containment{
							LowerPriority:   true,
							ContainmentType: ContainmentJobObject,
						},
					},
					Expiration:      3 * time.Minute,
					WaitForCapacity: true,
				},
			},
			Name:                "name",
			ManualTags:          []string{"manual:tag"},
			Authenticated:       "authuser",
			User:                "user1",
			ServiceAccount:      "serv",
			PubSubTopic:         "pubtop",
			PubSubAuthToken:     "pubauth",
			PubSubUserData:      "pubuser",
			ResultDBUpdateToken: "rdbtok",
			Realm:               "realm1",
			ResultDB:            true,
			BuildToken: BuildToken{
				BuildID:         5,
				Token:           "btok",
				BuildbucketHost: "bhost",
			},
		},
	}
}
