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

// ToDB converts the API to the model.
func (t *TasksNewRequest) ToDB(now time.Time, m *model.TaskRequest) error {
	// TODO(maruel): Validate!!
	m.Name = t.Name
	m.ParentTask = model.FromTaskID(t.ParentTaskID)
	m.Priority = t.Priority
	m.TaskSlices = make([]model.TaskSlice, len(t.TaskSlices))
	for i := range t.TaskSlices {
		if err := t.TaskSlices[i].ToDB(&m.TaskSlices[i]); err != nil {
			return err
		}
	}
	m.Tags = t.Tags
	m.User = t.User
	m.Created = now
	m.ServiceAccount = t.ServiceAccount
	m.PubSubAuthToken = t.PubSubAuthToken
	m.PubSubAuthToken = t.PubSubAuthToken
	m.PubSubUserData = t.PubSubUserData
	//m.BotPingToleranceSecs
	m.ResultDB = t.ResultDB.Enable
	m.Realm = t.Realm
	return nil
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

// FromDB converts the model to the API.
func (t *CIPDPackage) FromDB(m *model.CIPDPackage) {
	t.PkgName = m.PkgName
	t.Version = m.Version
	t.Path = m.Path
}

// ToDB converts the API to the model.
func (t *CIPDPackage) ToDB(m *model.CIPDPackage) {
	m.PkgName = t.PkgName
	m.Version = t.Version
	m.Path = t.Path
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
	Caches          []Cache          `json:"caches"`
	CIPDInput       CIPDInput        `json:"cipd_input"`
	Command         []string         `json:"command"`
	RelativeWD      string           `json:"relative_cwd"`
	Dimensions      []StringPair     `json:"dimensions"`
	Env             []StringPair     `json:"env"`
	EnvPrefixes     []StringListPair `json:"env_prefixes"`
	HardTimeoutSecs int64            `json:"execution_timeout_secs"`
	GracePeriodSecs int64            `json:"grace_period_secs"`
	Idempotent      bool             `json:"idempotent"`
	CASInputRoot    CASReference     `json:"cas_input_root"`
	IOTimeoutSecs   int64            `json:"io_timeout_secs"`
	Outputs         []string         `json:"outputs"`
	SecretBytes     []byte           `json:"secret_bytes"`
	Containment     Containment      `json:"containment"`
}

// FromDB converts the model to the API.
func (t *TaskProperties) FromDB(m *model.TaskProperties) {
	t.Caches = make([]Cache, len(m.Caches))
	for i := range m.Caches {
		t.Caches[i].Name = m.Caches[i].Name
		t.Caches[i].Path = m.Caches[i].Path
	}
	t.CIPDInput.Server = m.CIPDHost
	t.CIPDInput.ClientPackage.FromDB(&m.CIPDClient)
	t.CIPDInput.Packages = make([]CIPDPackage, len(m.CIPDPackages))
	for i := range m.CIPDPackages {
		t.CIPDInput.Packages[i].FromDB(&m.CIPDPackages[i])
	}
	t.Command = m.Command
	t.RelativeWD = m.RelativeWD
	t.Dimensions = ToStringPairs(m.Dimensions)
	t.Env = ToStringPairs(m.Env)
	t.EnvPrefixes = ToStringListPairs(m.EnvPrefixes)
	t.HardTimeoutSecs = int64(m.HardTimeout / time.Second)
	t.GracePeriodSecs = int64(m.GracePeriod / time.Second)
	t.Idempotent = m.Idempotent
	t.CASInputRoot.Host = m.CASHost
	t.CASInputRoot.Digest.FromDB(&m.Input)
	t.IOTimeoutSecs = int64(m.IOTimeout / time.Second)
	t.Outputs = m.Outputs
	// t.SecretBytes Never read back.
	t.Containment.ContainmentType = m.Containment.ContainmentType
}

// ToDB converts the API to the model.
func (t *TaskProperties) ToDB(m *model.TaskProperties) error {
	m.Caches = make([]model.Cache, len(t.Caches))
	for i := range t.Caches {
		m.Caches[i].Name = t.Caches[i].Name
		m.Caches[i].Path = t.Caches[i].Path
	}
	m.CIPDHost = t.CIPDInput.Server
	t.CIPDInput.ClientPackage.ToDB(&m.CIPDClient)
	m.CIPDPackages = make([]model.CIPDPackage, len(t.CIPDInput.Packages))
	for i := range t.CIPDInput.Packages {
		t.CIPDInput.Packages[i].ToDB(&m.CIPDPackages[i])
	}
	m.Command = t.Command
	m.RelativeWD = t.RelativeWD
	t.Dimensions = ToStringPairs(m.Dimensions)
	m.Env = FromStringPairs(t.Env)
	m.EnvPrefixes = FromStringListPairs(t.EnvPrefixes)
	m.HardTimeout = time.Duration(t.HardTimeoutSecs) * time.Second
	m.GracePeriod = time.Duration(t.GracePeriodSecs) * time.Second
	m.Idempotent = t.Idempotent
	m.CASHost = t.CASInputRoot.Host
	t.CASInputRoot.Digest.ToDB(&m.Input)
	m.IOTimeout = time.Duration(t.IOTimeoutSecs) * time.Second
	m.Outputs = t.Outputs
	m.SecretBytes = t.SecretBytes
	m.Containment.ContainmentType = t.Containment.ContainmentType
	return nil
}

// TaskSlice defines one "option" to run the task.
type TaskSlice struct {
	Properties      TaskProperties `json:"properties"`
	ExpirationSecs  int64          `json:"expiration_secs"`
	WaitForCapacity bool           `json:"wait_for_capacity"`
}

// FromDB converts the model to the API.
func (t *TaskSlice) FromDB(m *model.TaskSlice) {
	t.Properties.FromDB(&m.Properties)
	t.ExpirationSecs = int64(m.Expiration / time.Second)
	t.WaitForCapacity = m.WaitForCapacity
}

// ToDB converts the API to the model.
func (t *TaskSlice) ToDB(m *model.TaskSlice) error {
	t.Properties.ToDB(&m.Properties)
	m.Expiration = m.Expiration * time.Second
	m.WaitForCapacity = t.WaitForCapacity
	return nil
}

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
	t.Name = m.Name
	t.TaskID = model.ToTaskID(m.Key)
	t.ParentTaskID = model.ToTaskID(m.ParentTask)
	t.Priority = m.Priority
	t.Tags = m.Tags
	t.Created = CloudTime(m.Created)
	t.User = m.User
	t.Authenticated = m.Authenticated
	t.TaskSlices = make([]TaskSlice, len(m.TaskSlices))
	for i := range m.TaskSlices {
		t.TaskSlices[i].FromDB(&m.TaskSlices[i])
	}
	t.ServiceAccount = m.ServiceAccount
	t.Realm = m.Realm
	t.ResultDB.Enable = m.ResultDB
	t.PubSubTopic = m.PubSubTopic
	t.PubSubUserData = m.PubSubUserData
	// TODO(maruel):: t.BotPingToleranceSecs = m.BotPingTolerance
}

//

// TaskState is the state of the task request.
type TaskState int64

// FromDBTaskState converts a task state.
func FromDBTaskState(t model.TaskState) TaskState {
	// Values are currently the same but I expect this to change.
	switch t {
	case model.Running:
		return Running
	case model.Pending:
		return Pending
	case model.Expired:
		return Expired
	case model.Timedout:
		return Timedout
	case model.BotDied:
		return BotDied
	case model.Canceled:
		return Canceled
	case model.Completed:
		return Completed
	case model.Killed:
		return Killed
	case model.NoResource:
		return NoResource
	default:
		return BotDied
	}
}

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
	DurationSecs float64 `json:"duration"`
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
	BotDimensions    []StringListPair `json:"bot_dimensions"`
	BotID            string           `json:"bot_id"`
	BotIdleSince     Time             `json:"bot_idle_since_ts"`
	BotVersion       string           `json:"bot_version"`
	Children         []model.TaskID   `json:"children_task_ids"`
	Completed        Time             `json:"completed_ts"`
	CostSavedUSD     float64          `json:"cost_saved_usd"`
	Created          Time             `json:"created_ts"`
	DedupedFrom      model.TaskID     `json:"deduped_from"`
	DurationSecs     float64          `json:"duration"`
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
	CIPDPins         CIPDPins         `json:"cipd_pins"`
	RunID            model.TaskID     `json:"run_id"`
	CurrentTaskSlice int32            `json:"current_task_slice"`
	ResultDB         ResultDB         `json:"resultdb_info"`
}

// FromDB converts the model to the API.
func (t *TaskResult) FromDB(r *model.TaskRequest, m *model.TaskResult) {
	t.Abandoned = CloudTime(m.Abandoned)
	t.BotDimensions = ToStringListPairs(m.BotDimensions)
	t.BotID = m.BotID
	// TODO(maruel): t.BotIdleSince = m.BotIdleSince
	t.BotVersion = m.BotVersion
	t.Children = make([]model.TaskID, len(m.Children))
	for i, c := range m.Children {
		t.Children[i] = model.ToTaskID(c)
	}
	t.Completed = CloudTime(m.Completed)
	t.CostSavedUSD = 0 // TODO(maruel): implement.
	t.Created = CloudTime(r.Created)
	t.DedupedFrom = model.ToTaskID(m.DedupedFrom)
	t.DurationSecs = float64(m.Duration) / float64(time.Second)
	t.ExitCode = m.ExitCode
	t.Failure = m.ExitCode != 0
	t.InternalFailure = m.InternalFailure != ""
	t.Modified = CloudTime(m.Modified)
	// TODO(maruel): Use currentslice?
	t.CASOutput.Host = r.TaskSlices[0].Properties.CASHost
	t.CASOutput.Digest.FromDB(&m.Output)
	t.ServerVersions = m.ServerVersions
	t.Started = CloudTime(m.Started)
	t.State = FromDBTaskState(m.State)
	t.TaskID = model.ToTaskID(m.Key)
	t.TryNumber = 0
	if m.State != model.Pending {
		t.TryNumber = 1
	}
	t.CostsUSD = []float64{}
	t.Name = r.Name
	t.Tags = r.Tags
	t.User = r.User
	//t.Perf
	panic("TODO")
	//t.CIPDPins.Pkgs = m.CIPDPins
	t.RunID = t.TaskID // No difference in mess.
	t.CurrentTaskSlice = t.CurrentTaskSlice
	t.ResultDB.Host = m.ResultDB.Host
	t.ResultDB.Invocation = m.ResultDB.Invocation
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
