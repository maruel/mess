package messapi

import (
	"encoding/hex"
	"time"

	"github.com/maruel/mess/internal/model"
)

// Digest is a CAS reference.
type Digest struct {
	Hash string `json:"hash"`
	Size int64  `json:"size_bytes"`
}

// ToDB converts the Digest to DB's format.
func (d *Digest) ToDB(m *model.Digest) error {
	m.Size = d.Size
	_, err := hex.Decode(m.Hash[:], []byte(d.Hash))
	return err
}

// FromDB converts the Digest from DB's format.
func (d *Digest) FromDB(m *model.Digest) {
	d.Size = m.Size
	d.Hash = hex.EncodeToString(m.Hash[:])
}

// CIPDPackage is a LUCI CIPD package.
type CIPDPackage struct {
	PkgName string `json:"package_name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

// CacheEntry is a named cache entry.
type CacheEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// TaskOutput stores a task's output.
type TaskOutput struct {
	Size int64
}

// TasksCount is /tasks/count.
type TasksCount struct {
	Count int32 `json:"count"`
	Now   Time  `json:"now"`
}

// TasksList is /tasks/list.
type TasksList struct {
	Cursor string       `json:"cursor"`
	Items  []TaskResult `json:"items"`
	Now    Time         `json:"now"`
}

// ContainmentType declares the type of process containment the bot shall do.
type ContainmentType = model.ContainmentType

// Valid ContainmentType.
const (
	ContainmentNone      = model.ContainmentNone
	ContainmentAuto      = model.ContainmentAuto
	ContainmentJobObject = model.ContainmentJobObject
)

// Containment declares the type of process containment the bot shall do.
type Containment struct {
	LowerPriority   bool
	ContainmentType ContainmentType
}

// TaskProperties declares what the task runs.
type TaskProperties struct {
	Caches       []CacheEntry
	Command      []string
	RelativeWD   string
	CASHost      string
	Input        Digest
	CIPDHost     string
	CIPDPackages []CIPDPackage
	Dimensions   map[string]string
	Env          map[string]string
	EnvPrefixes  map[string]string
	HardTimeout  time.Duration
	GracePeriod  time.Duration
	IOTimeout    time.Duration
	Idempotent   bool
	Outputs      []string
	Containment  Containment
}

// TaskSlice defines one "option" to run the task.
type TaskSlice struct {
	Properties      TaskProperties
	Expiration      time.Duration
	WaitForCapacity bool
}

// BuildToken is a LUCI Buildbucket token.
type BuildToken struct {
	BuildID         int64
	Token           string
	BuildbucketHost string
}

// TaskRequest is a single requested task by a client. It is immutable.
type TaskRequest struct {
	Key                 model.TaskID
	Created             time.Time
	TaskSlices          []TaskSlice
	Name                string
	Authenticated       string
	User                string
	ServiceAccount      string
	Priority            int
	Tags                []string
	ManualTags          []string
	ParentTask          int64
	PubSubTopic         string
	PubSubAuthToken     string
	PubSubUserData      string
	ResultDBUpdateToken string
	Realm               string
	ResultDB            bool
	BuildToken          BuildToken
	//BotPingTolerance time.Duration
	//Expiration time.Time
}

// FromDB converts the model to the API.
func (t *TaskRequest) FromDB(m *model.TaskRequest) {
	panic("TODO")
}

// TaskState is the state of the task request.
type TaskState int64

// Valid TaskState.
const (
	Running TaskState = iota
	Pending
	Expired
	Timedout
	BotDied
	Canceled
	Completed
	Killed
	NoResource
)

// ResultDB declares the LUCI ResultDB information.
type ResultDB struct {
	Host       string
	Invocation string
}

// TaskResult is the result of running a TaskRequest.
type TaskResult struct {
	Key            model.TaskID
	BotID          string
	BotVersion     string
	BotDimension   map[string][]string
	BotIdleSince   time.Duration
	ServerVersions []string
	TaskSlice      int64
	DedupedFrom    int64
	PropertiesHash string

	TaskOutput TaskOutput
	ExitCode   int64
	State      TaskState
	Children   []int64
	Output     Digest
	CIPDPins   []CIPDPackage
	ResultDB   ResultDB

	Duration  time.Duration
	Started   time.Time
	Completed time.Time
	Abandoned time.Time
	Modified  time.Time
	Cost      float64

	// Internal flags.
	Killing   bool
	DeadAfter time.Time
	// TODO(maruel): Stats.
}

// FromDB converts the model to the API.
func (t *TaskResult) FromDB(m *model.TaskResult) {
	panic("TODO")
}
