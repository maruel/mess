package model

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	rbe "github.com/maruel/mess/third_party/build/bazel/remote/execution/v2"
)

// TaskRequest is a single requested task by a client. It is immutable.
type TaskRequest struct {
	Key                 int64       `json:"a,omitempty"`
	SchemaVersion       int64       `json:"b,omitempty"`
	Created             time.Time   `json:"c,omitempty"`
	Priority            int32       `json:"d,omitempty"`
	ParentTask          int64       `json:"e,omitempty"`
	Tags                []string    `json:"f,omitempty"`
	TaskSlices          []TaskSlice `json:"g,omitempty"`
	Name                string      `json:"h,omitempty"`
	ManualTags          []string    `json:"i,omitempty"`
	Authenticated       string      `json:"j,omitempty"`
	User                string      `json:"k,omitempty"`
	ServiceAccount      string      `json:"l,omitempty"`
	PubSubTopic         string      `json:"m,omitempty"`
	PubSubAuthToken     string      `json:"n,omitempty"`
	PubSubUserData      string      `json:"o,omitempty"`
	ResultDBUpdateToken string      `json:"p,omitempty"`
	Realm               string      `json:"q,omitempty"`
	ResultDB            bool        `json:"r,omitempty"`
	BuildToken          BuildToken  `json:"s,omitempty"`
}

// ValidateAndSetDefaults set default values and returns an error if the task
// request is invalid.
func (t *TaskRequest) ValidateAndSetDefaults() error {
	if t.Priority == 0 {
		t.Priority = 200
	}
	if t.Priority < 1 || t.Priority > 255 {
		return fmt.Errorf("invalid priority %d", t.Priority)
	}
	// Add tags from dimensions.
	tags := map[string]struct{}{}
	for _, n := range t.Tags {
		tags[n] = struct{}{}
	}
	for i := range t.TaskSlices {
		for k, v := range t.TaskSlices[i].Properties.Dimensions {
			tags[k+":"+v] = struct{}{}
		}
		// TODO(maruel): Expiration has to be checked here.
		if err := t.TaskSlices[i].ValidateAndSetDefaults(); err != nil {
			return err
		}
	}
	t.Tags = make([]string, 0, len(tags))
	for k := range tags {
		t.Tags = append(t.Tags, k)
	}
	sort.Strings(t.Tags)
	// TODO(maruel): All the other checks.
	return nil
}

// TaskID is a task ID as presented to the user.
type TaskID string

// ToTaskID converts an internal DB key to a external format.
func ToTaskID(key int64) TaskID {
	// Swarming uses the last nibbles:
	// - schema version, used 0 and 1. mess uses 2.
	// - retries, used 0, 1 and 2. mess uses 0
	if key <= 0 {
		return ""
	}
	return TaskID(strconv.FormatInt(key, 10) + "20")
}

// FromTaskID converts an external key to the internal DB format.
func FromTaskID(t TaskID) int64 {
	l := len(t)
	if l < 3 || t[l-2:l] != "20" {
		return 0
	}
	v, _ := strconv.ParseInt(string(t[:l-2]), 10, 64)
	if v < 0 {
		return 0
	}
	return v
}

type taskRequestSQL struct {
	key           int64
	schemaVersion int64
	created       int64
	priority      int32
	parentTask    int64
	tags          string
	blob          []byte
}

func (r *taskRequestSQL) fields() []interface{} {
	return []interface{}{
		&r.key,
		&r.schemaVersion,
		&r.created,
		&r.priority,
		&r.parentTask,
		&r.tags,
		&r.blob,
	}
}

func (r *taskRequestSQL) from(t *TaskRequest) {
	r.key = t.Key
	r.schemaVersion = t.SchemaVersion
	r.created = t.Created.UnixMicro()
	r.priority = t.Priority
	r.parentTask = t.ParentTask
	r.tags = ";" + strings.Join(t.Tags, ";") + ";"
	b := taskRequestSQLBlob{
		TaskSlices:          t.TaskSlices,
		Name:                t.Name,
		ManualTags:          t.ManualTags,
		Authenticated:       t.Authenticated,
		User:                t.User,
		ServiceAccount:      t.ServiceAccount,
		PubSubTopic:         t.PubSubTopic,
		PubSubAuthToken:     t.PubSubAuthToken,
		PubSubUserData:      t.PubSubUserData,
		ResultDBUpdateToken: t.ResultDBUpdateToken,
		Realm:               t.Realm,
		ResultDB:            t.ResultDB,
		BuildToken:          t.BuildToken,
	}
	var err error
	r.blob, err = json.Marshal(&b)
	if err != nil {
		panic("internal error: " + err.Error())
	}
}

func (r *taskRequestSQL) to(t *TaskRequest) {
	t.Key = r.key
	t.SchemaVersion = r.schemaVersion
	t.Created = time.UnixMicro(r.created).UTC()
	t.Priority = r.priority
	t.ParentTask = r.parentTask
	t.Tags = strings.Split(r.tags[1:len(r.tags)-1], ";")
	b := taskRequestSQLBlob{}
	if err := json.Unmarshal(r.blob, &b); err != nil {
		panic("internal error: " + err.Error())
	}
	t.TaskSlices = b.TaskSlices
	t.Name = b.Name
	t.ManualTags = b.ManualTags
	t.Authenticated = b.Authenticated
	t.User = b.User
	t.ServiceAccount = b.ServiceAccount
	t.PubSubTopic = b.PubSubTopic
	t.PubSubAuthToken = b.PubSubAuthToken
	t.PubSubUserData = b.PubSubUserData
	t.ResultDBUpdateToken = b.ResultDBUpdateToken
	t.Realm = b.Realm
	t.ResultDB = b.ResultDB
	t.BuildToken = b.BuildToken
}

// See:
// - https://sqlite.org/lang_createtable.html#rowids_and_the_integer_primary_key
// - https://sqlite.org/datatype3.html
// BLOB
const schemaTaskRequest = `
CREATE TABLE IF NOT EXISTS TaskRequest (
	key           INTEGER NOT NULL,
	schemaVersion INTEGER NOT NULL,
	created       INTEGER NOT NULL,
	priority      INTEGER NOT NULL,
	parentTask    INTEGER,
	tags          TEXT,
	blob          BLOB    NOT NULL,
	PRIMARY KEY(key DESC)
) STRICT;
`

// taskRequestSQLBlob contains the unindexed fields.
type taskRequestSQLBlob struct {
	TaskSlices          []TaskSlice `json:"a,omitempty"`
	Name                string      `json:"b,omitempty"`
	ManualTags          []string    `json:"c,omitempty"`
	Authenticated       string      `json:"d,omitempty"`
	User                string      `json:"e,omitempty"`
	ServiceAccount      string      `json:"f,omitempty"`
	PubSubTopic         string      `json:"g,omitempty"`
	PubSubAuthToken     string      `json:"h,omitempty"`
	PubSubUserData      string      `json:"i,omitempty"`
	ResultDBUpdateToken string      `json:"j,omitempty"`
	Realm               string      `json:"k,omitempty"`
	ResultDB            bool        `json:"l,omitempty"`
	BuildToken          BuildToken  `json:"m,omitempty"`
	//BotPingTolerance time.Duration `json:""`
	//Expiration time.Time          `json:""`
}

// ContainmentType declares the type of process containment the bot shall do.
type ContainmentType int

// Valid ContainmentType.
const (
	ContainmentNotSpecified ContainmentType = iota
	ContainmentNone
	ContainmentAuto
	ContainmentJobObject
)

// Containment declares the type of process containment the bot shall do.
type Containment struct {
	ContainmentType ContainmentType `json:"a,omitempty"`
}

// TaskSlice defines one "option" to run the task.
type TaskSlice struct {
	Properties      TaskProperties `json:"a,omitempty"`
	Expiration      time.Duration  `json:"b,omitempty"`
	WaitForCapacity bool           `json:"c,omitempty"`
}

// ValidateAndSetDefaults set default values and returns an error if the task
// request is invalid.
func (t *TaskSlice) ValidateAndSetDefaults() error {
	if err := t.Properties.ValidateAndSetDefaults(); err != nil {
		return err
	}
	// TODO(maruel): Calculate the sum.
	if t.Expiration == 0 {
		t.Expiration = time.Hour
	}
	if t.Expiration < time.Second || t.Expiration > 3*24*time.Hour+time.Minute {
		return fmt.Errorf("invalid expiration %s", t.Expiration)
	}
	return nil
}

// TaskProperties declares what the task runs.
type TaskProperties struct {
	Caches       []Cache             `json:"a,omitempty"`
	Command      []string            `json:"b,omitempty"`
	RelativeWD   string              `json:"c,omitempty"`
	CASHost      string              `json:"d,omitempty"`
	Input        Digest              `json:"e,omitempty"`
	CIPDHost     string              `json:"f,omitempty"`
	CIPDClient   CIPDPackage         `json:"g,omitempty"`
	CIPDPackages []CIPDPackage       `json:"h,omitempty"`
	Dimensions   map[string]string   `json:"i,omitempty"`
	Env          map[string]string   `json:"j,omitempty"`
	EnvPrefixes  map[string][]string `json:"k,omitempty"`
	HardTimeout  time.Duration       `json:"l,omitempty"`
	GracePeriod  time.Duration       `json:"m,omitempty"`
	IOTimeout    time.Duration       `json:"n,omitempty"`
	SecretBytes  []byte              `json:"o,omitempty"`
	Idempotent   bool                `json:"p,omitempty"`
	Outputs      []string            `json:"q,omitempty"`
	Containment  Containment         `json:"r,omitempty"`
}

// ValidateAndSetDefaults set default values and returns an error if the task
// request is invalid.
func (t *TaskProperties) ValidateAndSetDefaults() error {
	// TODO(maruel): There's a bug in the Swarming bot that the CIPD client is
	// *always* enabled. So we need to set a valid CIPD client here. This should
	// be removed as soon as possible.
	if t.CIPDHost == "" {
		t.CIPDHost = "https://chrome-infra-packages.appspot.com"
	}
	if t.CIPDClient.PkgName == "" {
		t.CIPDClient.PkgName = "infra/tools/cipd/${platform}"
		t.CIPDClient.Version = "git_revision:8e9b0c80860d00dfe951f7ea37d74e210d376c13"
		t.CIPDClient.Path = ""
	}
	return nil
}

// BuildToken is a LUCI Buildbucket token.
type BuildToken struct {
	BuildID         int64  `json:"a,omitempty"`
	Token           string `json:"b,omitempty"`
	BuildbucketHost string `json:"c,omitempty"`
}

// Digest is a more memory efficient version of rbe.Digest.
type Digest struct {
	Size int64    `json:"a,omitempty"`
	Hash [32]byte `json:"b,omitempty"`
}

// ToProto converts to RBE's digest message.
func (d *Digest) ToProto(p *rbe.Digest) {
	p.SizeBytes = d.Size
	p.Hash = hex.EncodeToString(d.Hash[:])
}

// FromProto converts from RBE's digest message.
func (d *Digest) FromProto(p *rbe.Digest) error {
	d.Size = p.SizeBytes
	if len(p.Hash) != 64 {
		return errors.New("invalid hash")
	}
	// TODO: Manually decode for performance.
	_, err := hex.Decode(d.Hash[:], []byte(p.Hash))
	return err
}

// CIPDPackage declares a LUCI CIPD package.
type CIPDPackage struct {
	PkgName string `json:"a,omitempty"`
	Version string `json:"b,omitempty"`
	Path    string `json:"c,omitempty"`
}

// Cache is a named cache that survives across tasks.
type Cache struct {
	Name string `json:"a,omitempty"`
	Path string `json:"b,omitempty"`
}
