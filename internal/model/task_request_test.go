package model

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTaskRequestJSON(t *testing.T) {
	p := filepath.Join(t.TempDir(), "db.json.zst")
	d, err := NewDBJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskRequestCount(); l != 0 {
		t.Fatal(l)
	}
	want := getTaskRequest()
	d.TaskRequestAdd(want)
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
	d.TaskRequestGet(1, &got)
	missing := TaskRequest{}
	d.TaskRequestGet(2, &missing)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, &got); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if diff := cmp.Diff(&TaskRequest{}, &missing); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
}

func TestTaskRequestSQL(t *testing.T) {
	p := filepath.Join(t.TempDir(), "mess.db")
	d, err := NewDBSqlite3(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskRequestCount(); l != 0 {
		t.Fatal(l)
	}
	want := getTaskRequest()
	d.TaskRequestAdd(want)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = NewDBSqlite3(p)
	if err != nil {
		t.Fatal(err)
	}
	if l := d.TaskRequestCount(); l != 1 {
		t.Fatal(l)
	}
	got := TaskRequest{}
	d.TaskRequestGet(1, &got)
	missing := TaskRequest{}
	d.TaskRequestGet(2, &missing)
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, &got); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
	if diff := cmp.Diff(&TaskRequest{}, &missing); diff != "" {
		t.Fatalf("(want +got):\n%s", diff)
	}
}

func TestTaskRequestNonZero(t *testing.T) {
	r := getTaskRequest()
	r.Key = 2
	if err := isNonZero("", reflect.ValueOf(r)); err != nil {
		t.Fatal(err)
	}
	r.Tags = nil
	if err := isNonZero("", reflect.ValueOf(r)); err == nil || err.Error() != "*TaskRequest.Tags slice is empty" {
		t.Fatal(err)
	}
}

func getTaskRequest() *TaskRequest {
	r := &TaskRequest{
		SchemaVersion: 1,
		Key:           0,
		Created:       time.Date(2020, 3, 13, 10, 9, 8, 7000, time.UTC),
		Priority:      3,
		ParentTask:    4,
		Tags: []string{
			"a:b", "c:d",
		},
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
						Hash: [32]byte{1, 2, 3},
					},
					CIPDHost: "chrome-package",
					CIPDClient: CIPDPackage{
						PkgName: "cipdclient",
						Version: "cver",
						Path:    "cpath",
					},
					CIPDPackages: []CIPDPackage{
						{
							PkgName: "pkgname",
							Version: "pkgver",
							Path:    "pkgpath",
						},
					},
					Dimensions:  map[string]string{"os": "Windows"},
					Env:         map[string]string{"FOO": "bar"},
					EnvPrefixes: map[string][]string{"PATH": {"./foo"}},
					HardTimeout: time.Minute,
					GracePeriod: time.Second,
					IOTimeout:   2 * time.Minute,
					SecretBytes: []byte("sekret"),
					Idempotent:  true,
					Outputs:     []string{"foo"},
					Containment: Containment{
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
	}
	return r
}

// isNonZero iterates recursively and check for reflect.Value.IsZero().
func isNonZero(prefix string, v reflect.Value) error {
	//log.Printf("isNonZero(%s, %s)", prefix, v.Type().Name())
	switch v.Kind() {
	case reflect.Struct:
		l := v.NumField()
		for i := 0; i < l; i++ {
			name := v.Type().Field(i).Name
			if !('A' <= name[0] && name[0] <= 'Z') {
				// Not exported.
				continue
			}
			if err := isNonZero(prefix+"."+name, v.Field(i)); err != nil {
				return err
			}
		}
	case reflect.Slice:
		l := v.Len()
		if l == 0 {
			return errors.New(prefix + " slice is empty")
		}
		for i := 0; i < l; i++ {
			if err := isNonZero(fmt.Sprintf("%s[%d]", prefix, i), v.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		l := v.Len()
		if l == 0 {
			return errors.New(prefix + " map is empty")
		}
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			if err := isNonZero(prefix+"["+k.String()+"]", k); err != nil {
				return err
			}
			val := iter.Value()
			if err := isNonZero(prefix+"["+k.String()+"]", val); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if v.IsNil() {
			return errors.New(prefix + "* is nil")
		}
		if err := isNonZero(prefix+"*"+v.Elem().Type().Name(), v.Elem()); err != nil {
			return err
		}
	default:
		if v.IsZero() {
			return errors.New(prefix + " value is zero")
		}
	}
	return nil
}
