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
	//ExpirationSecs int64 `json:"expiration_secs"`
	Name         string       `json:"name"`
	ParentTaskID model.TaskID `json:"parent_task_id"`
	Priority     int32        `json:"priority"`
	//Properties TaskProperties `json:"properties"`
	TaskSlices           []TaskSlice      `json:"task_slices"`
	Tags                 []string         `json:"tags"`
	User                 string           `json:"user"`
	ServiceAccount       string           `json:"service_account"`
	PubSubTopic          string           `json:"pubsub_topic"`
	PubSubAuthToken      string           `json:"pubsub_auth_token"`
	PubSubUserData       string           `json:"pubsub_userdata"`
	EvaluateOnly         bool             `json:"evaluate_only"`
	PoolTaskTemplate     PoolTaskTemplate `json:"pool_task_template"`
	BotPingToleranceSecs int64            `json:"bot_ping_telerance_secs"`
	RequestUUID          string           `json:"request_uuid"`
	ResultDB             ResultDBCfg      `json:"resultdb"`
	Realm                string           `json":realm"`
}

// TasksNewResponse is /tasks/new (POST).
type TasksNewResponse struct {
	Request TaskRequest  `json:"request"`
	TaskID  model.TaskID `json:"task_id"`
	Result  TaskResult   `json:"task_result"`
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

// CASReference is a reference to a RBE-CAS host and digest.
type CASReference struct {
	Host   string `json:"cas_instance"`
	Digest Digest `json:"digest"`
}

// CIPDInput is a LUCI CIPD input reference.
type CIPDInput struct {
	Server        string        `json:"server"`
	ClientPackage CIPDPackage   `json:"client_package"`
	Packages      []CIPDPackage `json:"packages"`
}

// CIPDPins is a LUCI CIPD resolved reference.
type CIPDPins struct {
	ClientPkg CIPDPackage `json:"client_package"`
	Pkgs      CIPDPackage `json:"packages"`
}

// CIPDPackage is a LUCI CIPD package.
type CIPDPackage struct {
	PkgName string `json:"package_name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

// Cache is a named cache entry.
type Cache struct {
	Name string `json:"name"`
	Path string `json:"path"`
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
	ContainmentType ContainmentType `json:"containment_type"`
}

// TaskProperties declares what the task runs.
type TaskProperties struct {
	Caches       []Cache          `json:"caches"`
	CIPDInput    CIPDInput        `json:"cipd_input"`
	Command      []string         `json:"command"`
	RelativeWD   string           `json:"relative_cwd"`
	Dimensions   []StringPair     `json:"dimensions"`
	Env          []StringPair     `json:"env"`
	EnvPrefixes  []StringListPair `json:"env_prefixes"`
	HardTimeout  int64            `json:"execution_timeout_secs"`
	GracePeriod  int64            `json:"grace_period_secs"`
	Idempotent   bool             `json:"idempotent"`
	CASInputRoot CASReference     `json:"cas_input_root"`
	IOTimeout    int64            `json:"io_timeout_secs"`
	Outputs      []string         `json:"outputs"`
	SecretBytes  []byte           `json:"secret_bytes"`
	Containment  Containment      `json:"containment"`
}

// TaskSlice defines one "option" to run the task.
type TaskSlice struct {
	Properties      TaskProperties `json:"properties"`
	ExpirationSecs  int64          `json:"expiration_secs"`
	WaitForCapacity bool           `json:"wait_for_capacity"`
}

/* REMOVE
// BuildToken is a LUCI Buildbucket token.
type BuildToken struct {
	BuildID         int64
	Token           string
	BuildbucketHost string
}
*/

// PoolTaskTemplate determines the kind of template to use.
type PoolTaskTemplate int32

// Valid PoolTaskTemplate.
const (
	PoolTaskTemplateAuto PoolTaskTemplate = iota
	PoolTaskTemplateCanaryPrefer
	PoolTaskTemplateCanaryNever
	PoolTaskTemplateSkip
)

// ResultDBCfg is used in TasksNewRequest.
type ResultDBCfg struct {
	Enable bool `json:"enable"`
}

// TaskRequest is a single requested task by a client. It is immutable.
type TaskRequest struct {
	//ExpirationSecs int64 `json:"expiration_secs"`
	Name         string       `json:"name"`
	TaskID       model.TaskID `json:"task_id"`
	ParentTaskID model.TaskID `json:"parent_task_id"`
	Priority     int32        `json:"priority"`
	//Properties TaskProperties `json:"properties"`
	Tags                 []string    `json:"tags"`
	Created              Time        `json:"created_ts"`
	User                 string      `json:"user"`
	Authenticated        string      `json:"authenticated"`
	TaskSlices           []TaskSlice `json:"task_slices"`
	ServiceAccount       string      `json:"service_account"`
	Realm                string      `json":realm"`
	ResultDB             ResultDBCfg `json:"resultdb"`
	PubSubTopic          string      `json:"pubsub_topic"`
	PubSubUserData       string      `json:"pubsub_userdata"`
	BotPingToleranceSecs int64       `json:"bot_ping_telerance_secs"`
}

// FromDB converts the model to the API.
func (t *TaskRequest) FromDB(m *model.TaskRequest) {
	panic("TODO")
}

//

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
	Host       string `json:"hostname"`
	Invocation string `json:"invocation"`
}

// OperationStats is the statistic for one operation.
type OperationStats struct {
	Duration float64 `json:"duration"`
}

// CASOperationStats is RBE-CAS operation.
type CASOperationStats struct {
	Duration        float64 `json:"duration"`
	InitialNumItems int64   `json:"initial_number_items"`
	InitialSize     int64   `json:"initial_size"`
	ItemsCold       []byte  `json:"items_cold"` // zlib deflated varints.
	ItemsHot        []byte  `json:"items_hot"`
	NumItemsCold    int     `json:"num_items_cold"`
	NumItemsHot     int     `json:"num_items_hot"`
}

// TaskPerfStats is the performance stats for a task.
type TaskPerfStats struct {
	BotOverhead          OperationStats    `json:"bot_overhead"`
	CASDownload          CASOperationStats `json:"isolated_download"`
	CASUpload            CASOperationStats `json:"isolated_upload"`
	PkgInstallation      OperationStats    `json:"package_installation"`
	CacheTrim            OperationStats    `json:"cache_trim"`
	NamnedCachesInstall  OperationStats    `json:"namned_caches_install"`
	NamedCachesUninstall OperationStats    `json:"named_caches_uninstall"`
	Cleanup              OperationStats    `json:"cleanup"`
}

// TaskResult is the result of running a TaskRequest.
type TaskResult struct {
	Abandoned        Time             `json:"abandoned_ts"`
	BotDimension     []StringListPair `json:"bot_dimensions"`
	BotID            string           `json:"bot_id"`
	BotIdleSince     Time             `json:"bot_idle_since_ts"`
	BotVersion       string           `json:"bot_version"`
	Children         []model.TaskID   `json:"children_task_ids"`
	Completed        Time             `json:"completed_ts"`
	CostSavedUSD     float64          `json:"cost_saved_usd"`
	Created          Time             `json:"created_ts"`
	DedupedFrom      model.TaskID     `json:"deduped_from"`
	Duration         float64          `json:"duration"`
	ExitCode         int32            `json:"exit_code"`
	Failure          bool             `json:"failure"`
	InternalFailure  bool             `json:"internal_failure"`
	Modified         Time             `json:"modified_ts"`
	CASOutput        CASReference     `json:"cas_output_root"`
	ServerVersions   []string         `json:"server_verions"`
	Started          Time             `json:"started_ts"`
	State            TaskState        `json:"state"`
	TaskID           model.TaskID     `json:"task_id"`
	TryNumber        int32            `json:"try_number"`
	CostsUSD         []float64        `json:"costs_usd"`
	Name             string           `json:"name"`
	Tags             []string         `json:"tags"`
	User             string           `json:"user"`
	Perf             TaskPerfStats    `json:"performance_stats"`
	CIPDPIns         CIPDPins         `json:"cipd_pins"`
	RunID            model.TaskID     `json:"run_id"`
	CurrentTaskSlice int32            `json:"current_task_slice"`
	ResultDB         ResultDB         `json:"resultdb_info"`
}

// FromDB converts the model to the API.
func (t *TaskResult) FromDB(m *model.TaskResult) {
	panic("TODO")
}

//

// TaskStateQuery is TODO. Default is ALL
type TaskStateQuery = string

// TaskSort is TODO. Default is CREATED_TS
type TaskSort = string

// TaskQueue is a task queue.
type TaskQueue struct {
	Dimensions []string `json:"dimensions"`
	ValidUntil Time     `json:"valid_until_ts"`
}
