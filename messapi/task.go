package messapi

import (
	"encoding/hex"
	"time"

	"github.com/maruel/mess/internal/model"
)

// TasksCancelRequest is /tasks/cancel (POST).
type TasksCancelRequest struct {
	Tags        []string `json:"tags"`
	Cursor      string   `json:"cursor"`
	Limit       int64    `json:"limit"`
	KillRunning bool     `json:"kill_running"`
	End         float64  `json:"end"`
}

// TasksCancelResponse is /tasks/cancel (POST).
type TasksCancelResponse struct {
	Cursor  string `json:"cursor"`
	Now     Time   `json:"now"`
	Matched int64  `json:"matched"`
}

// TasksCountRequest is /tasks/count (GET).
type TasksCountRequest struct {
	End   time.Time
	Start time.Time
	State TaskStateQuery
	Tags  []string
}

// TasksCountResponse is /tasks/count (GET).
type TasksCountResponse struct {
	Count int32 `json:"count"`
	Now   Time  `json:"now"`
}

// TasksGetStateRequest is /tasks/get_states (GET).
type TasksGetStateRequest struct {
	TaskID []string
}

// TasksGetStateResponse is /tasks/get_states (GET).
type TasksGetStateResponse struct {
	States []TaskState `json:"states"`
}

// TasksListRequest is /tasks/list (GET).
type TasksListRequest struct {
	Limit                   int64
	Cursor                  string
	End                     time.Time
	Start                   time.Time
	State                   TaskStateQuery
	Tags                    []string
	Sort                    TaskSort
	IncludePerformanceStats bool
}

// TasksListResponse is /tasks/list (GET).
type TasksListResponse struct {
	Cursor string       `json:"cursor"`
	Items  []TaskResult `json:"items"`
	Now    Time         `json:"now"`
}

// TasksNewRequest is /tasks/new (POST).
type TasksNewRequest struct {
}

// TasksNewResponse is /tasks/new (POST).
type TasksNewResponse struct {
}

// TasksRequestsRequest is /tasks/requests (GET).
type TasksRequestsRequest struct {
	Limit                   int64
	Cursor                  string
	End                     time.Time
	Start                   time.Time
	State                   TaskStateQuery
	Tags                    []string
	Sort                    TaskSort
	IncludePerformanceStats bool
}

// TasksRequestsResponse is /tasks/requests (GET).
type TasksRequestsResponse struct {
	Cursor string        `json:"cursor"`
	Items  []TaskRequest `json:"items"`
	Now    Time          `json:"now"`
}

// TaskCancelResponse is /task/<id>/cancel (POST).
type TaskCancelResponse struct {
	Ok         bool `json:"ok"`
	WasRunning bool `json:"was_running"`
}

// TaskRequestResponse is /task/<id>/request (GET).
type TaskRequestResponse = TaskRequest

// TaskResultRequest is /task/<id>/result (GET).
type TaskResultRequest struct {
	IncludePerformanceStats bool
}

// TaskResultResponse is /task/<id>/result (GET).
type TaskResultResponse = TaskResult

// TaskStdoutRequest is /task/<id>/stdout (GET).
type TaskStdoutRequest struct {
	Offset int64
	Length int64
}

// TaskStdoutResponse is /task/<id>/stdout (GET).
type TaskStdoutResponse struct {
	Output string    `json:"output"`
	State  TaskState `json:"state"`
}

// TaskQueuesListRequest is /queues/list (GET).
type TaskQueuesListRequest struct {
	Cursor string
	Limit  int64
}

// TaskQueuesListResponse is /queues/list (GET).
type TaskQueuesListResponse struct {
	Cursor string      `json:"cursor"`
	Items  []TaskQueue `json:"items"`
	Now    Time        `json:"now"`
}

//

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
	Priority            int32
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

// TaskStateQuery is TODO. Default is ALL
type TaskStateQuery = string

// TaskSort is TODO. Default is CREATED_TS
type TaskSort = string

// TaskQueue is a task queue.
type TaskQueue struct {
	Dimensions []string `json:"dimensions"`
	ValidUntil Time     `json:"valid_until_ts"`
}
