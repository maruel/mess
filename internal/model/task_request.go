package model

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	rbe "github.com/maruel/mess/third_party/build/bazel/remote/execution/v2"
)

// TaskRequest is a single requested task by a client. It is immutable.
type TaskRequest struct {
	Key                 int64       `json:"a"`
	SchemaVersion       int         `json:"b"`
	Created             time.Time   `json:"c"`
	Priority            int         `json:"d"`
	ParentTask          int64       `json:"e"`
	Tags                []string    `json:"f"`
	TaskSlices          []TaskSlice `json:"g"`
	Name                string      `json:"h"`
	ManualTags          []string    `json:"i"`
	Authenticated       string      `json:"j"`
	User                string      `json:"k"`
	ServiceAccount      string      `json:"l"`
	PubSubTopic         string      `json:"m"`
	PubSubAuthToken     string      `json:"n"`
	PubSubUserData      string      `json:"o"`
	ResultDBUpdateToken string      `json:"p"`
	Realm               string      `json:"q"`
	ResultDB            bool        `json:"r"`
	BuildToken          BuildToken  `json:"s"`
}

// TaskID is a task ID as presented to the user.
type TaskID string

// ToTaskID converts an internal DB key to a external format.
func ToTaskID(key int64) TaskID {
	// Swarming uses the last nibbles:
	// - schema version, used 0 and 1. mess uses 2.
	// - retries, used 0, 1 and 2. mess uses 0
	if key < 0 {
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
	schemaVersion int
	created       int64
	priority      int
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
	TaskSlices          []TaskSlice `json:"a"`
	Name                string      `json:"b"`
	ManualTags          []string    `json:"c"`
	Authenticated       string      `json:"d"`
	User                string      `json:"e"`
	ServiceAccount      string      `json:"f"`
	PubSubTopic         string      `json:"g"`
	PubSubAuthToken     string      `json:"h"`
	PubSubUserData      string      `json:"i"`
	ResultDBUpdateToken string      `json:"j"`
	Realm               string      `json:"k"`
	ResultDB            bool        `json:"l"`
	BuildToken          BuildToken  `json:"m"`
	//BotPingTolerance time.Duration `json:""`
	//Expiration time.Time          `json:""`
}

// ContainmentType declares the type of process containment the bot shall do.
type ContainmentType int

// Valid ContainmentType.
const (
	ContainmentNone ContainmentType = iota
	ContainmentAuto
	ContainmentJobObject
)

// Containment declares the type of process containment the bot shall do.
type Containment struct {
	LowerPriority   bool            `json:"a"`
	ContainmentType ContainmentType `json:"b"`
}

// TaskProperties declares what the task runs.
type TaskProperties struct {
	Caches       []Cache           `json:"a"`
	Command      []string          `json:"b"`
	RelativeWD   string            `json:"c"`
	CASHost      string            `json:"d"`
	Input        Digest            `json:"e"`
	CIPDHost     string            `json:"f"`
	CIPDPackages []CIPDPackage     `json:"g"`
	Dimensions   map[string]string `json:"h"`
	Env          map[string]string `json:"i"`
	EnvPrefixes  map[string]string `json:"j"`
	HardTimeout  time.Duration     `json:"k"`
	GracePeriod  time.Duration     `json:"l"`
	IOTimeout    time.Duration     `json:"m"`
	Idempotent   bool              `json:"n"`
	Outputs      []string          `json:"o"`
	Containment  Containment       `json:"p"`
}

// TaskSlice defines one "option" to run the task.
type TaskSlice struct {
	Properties      TaskProperties `json:"a"`
	Expiration      time.Duration  `json:"b"`
	WaitForCapacity bool           `json:"c"`
}

// BuildToken is a LUCI Buildbucket token.
type BuildToken struct {
	BuildID         int64  `json:"a"`
	Token           string `json:"b"`
	BuildbucketHost string `json:"c"`
}

// Digest is a more memory efficient version of rbe.Digest.
type Digest struct {
	Size int64    `json:"a"`
	Hash [32]byte `json:"b"`
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
	PkgName string `json:"a"`
	Version string `json:"b"`
	Path    string `json:"c"`
}

// Cache is a named cache that survives across tasks.
type Cache struct {
	Name string `json:"a"`
	Path string `json:"b"`
}
