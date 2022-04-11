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
	Limit       Int      `json:"limit"`
	KillRunning bool     `json:"kill_running"`
	End         float64  `json:"end"`
}

// TasksCancelResponse is /tasks/cancel (POST).
type TasksCancelResponse struct {
	Cursor  string `json:"cursor,omitempty"`
	Now     Time   `json:"now,omitempty"`
	Matched Int    `json:"matched,omitempty"`
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
	Count int32 `json:"count,omitempty"`
	Now   Time  `json:"now,omitempty"`
}

// TasksGetStateRequest is /tasks/get_states (GET).
type TasksGetStateRequest struct {
	TaskID []string
}

// TasksGetStateResponse is /tasks/get_states (GET).
type TasksGetStateResponse struct {
	States []TaskState `json:"states,omitempty"`
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
	Cursor string       `json:"cursor,omitempty"`
	Items  []TaskResult `json:"items,omitempty"`
	Now    Time         `json:"now,omitempty"`
}

// TasksNewRequest is /tasks/new (POST).
type TasksNewRequest struct {
	//ExpirationSecs Int `json:"expiration_secs"`
	Name         string       `json:"name"`
	ParentTaskID model.TaskID `json:"parent_task_id"`
	Priority     Int          `json:"priority"`
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
	BotPingToleranceSecs Int              `json:"bot_ping_telerance_secs"`
	RequestUUID          string           `json:"request_uuid"`
	ResultDB             ResultDBCfg      `json:"resultdb"`
	Realm                string           `json:"realm"`
}

// ToDB converts the API to the model.
func (t *TasksNewRequest) ToDB(now time.Time, m *model.TaskRequest) error {
	// TODO(maruel): Validate!!
	m.Name = t.Name
	m.ParentTask = model.FromTaskID(t.ParentTaskID)
	m.Priority = t.Priority.Int32()
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
	Request TaskRequest  `json:"request,omitempty"`
	TaskID  model.TaskID `json:"task_id,omitempty"`
	Result  TaskResult   `json:"task_result,omitempty"`
}

// TasksRequestsRequest is /tasks/requests (GET).
type TasksRequestsRequest struct {
	Limit  int64
	Cursor string
	End    time.Time
	Start  time.Time
	State  TaskStateQuery
	Tags   []string
	Sort   TaskSort
}

// TasksRequestsResponse is /tasks/requests (GET).
type TasksRequestsResponse struct {
	Cursor string        `json:"cursor,omitempty"`
	Items  []TaskRequest `json:"items,omitempty"`
	Now    Time          `json:"now,omitempty"`
}

// TaskCancelResponse is /task/<id>/cancel (POST).
type TaskCancelResponse struct {
	Ok         bool `json:"ok,omitempty"`
	WasRunning bool `json:"was_running,omitempty"`
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
	Output string    `json:"output,omitempty"`
	State  TaskState `json:"state,omitempty"`
}

// TaskQueuesListRequest is /queues/list (GET).
type TaskQueuesListRequest struct {
	Cursor string
	Limit  int64
}

// TaskQueuesListResponse is /queues/list (GET).
type TaskQueuesListResponse struct {
	Cursor string      `json:"cursor,omitempty"`
	Items  []TaskQueue `json:"items,omitempty"`
	Now    Time        `json:"now,omitempty"`
}

//

// Digest is a CAS reference.
type Digest struct {
	Hash string `json:"hash,omitempty"`
	Size Int    `json:"size_bytes,omitempty"`
}

// ToDB converts the Digest to DB's format.
func (d *Digest) ToDB(m *model.Digest) error {
	m.Size = d.Size.Int64()
	_, err := hex.Decode(m.Hash[:], []byte(d.Hash))
	return err
}

// FromDB converts the Digest from DB's format.
func (d *Digest) FromDB(m *model.Digest) {
	d.Size.Set64(m.Size)
	if m.Size == 0 {
		d.Hash = ""
	} else {
		d.Hash = hex.EncodeToString(m.Hash[:])
	}
}

// CASReference is a reference to a RBE-CAS host and digest.
type CASReference struct {
	Host   string `json:"cas_instance,omitempty"`
	Digest Digest `json:"digest,omitempty"`
}

// CIPDInput is a LUCI CIPD input reference.
type CIPDInput struct {
	Server        string        `json:"server,omitempty"`
	ClientPackage CIPDPackage   `json:"client_package,omitempty"`
	Packages      []CIPDPackage `json:"packages,omitempty"`
}

// CIPDPins is a LUCI CIPD resolved reference.
type CIPDPins struct {
	ClientPkg CIPDPackage   `json:"client_package,omitempty"`
	Pkgs      []CIPDPackage `json:"packages,omitempty"`
}

// CIPDPackage is a LUCI CIPD package.
type CIPDPackage struct {
	PkgName string `json:"package_name,omitempty"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
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
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

// ContainmentType declares the type of process containment the bot shall do.
type ContainmentType string

// Valid ContainmentType.
const (
	ContainmentNone      = model.ContainmentNone
	ContainmentAuto      = model.ContainmentAuto
	ContainmentJobObject = model.ContainmentJobObject
)

func (c ContainmentType) ToDB() model.ContainmentType {
	switch c {
	case "NONE":
		return model.ContainmentNone
	case "AUTO":
		return model.ContainmentAuto
	case "JOB_OBJECT":
		return model.ContainmentJobObject
	default:
		return model.ContainmentNotSpecified
	}
}

// FromDB converts the model to the API.
func (c *ContainmentType) FromDB(m model.ContainmentType) {
	switch m {
	case model.ContainmentNone:
		*c = "NONE"
	case model.ContainmentAuto:
		*c = "AUTO"
	case model.ContainmentJobObject:
		*c = "JOB_OBJECT"
	default:
		*c = "NOT_SPECIFIED"
	}
}

// Containment declares the type of process containment the bot shall do.
type Containment struct {
	ContainmentType ContainmentType `json:"containment_type,omitempty"`
}

// TaskProperties declares what the task runs.
type TaskProperties struct {
	Caches          []Cache          `json:"caches,omitempty"`
	CIPDInput       CIPDInput        `json:"cipd_input,omitempty"`
	Command         []string         `json:"command,omitempty"`
	RelativeWD      string           `json:"relative_cwd,omitempty"`
	Dimensions      []StringPair     `json:"dimensions,omitempty"`
	Env             []StringPair     `json:"env,omitempty"`
	EnvPrefixes     []StringListPair `json:"env_prefixes,omitempty"`
	HardTimeoutSecs Int              `json:"execution_timeout_secs,omitempty"`
	GracePeriodSecs Int              `json:"grace_period_secs,omitempty"`
	Idempotent      bool             `json:"idempotent,omitempty"`
	CASInputRoot    CASReference     `json:"cas_input_root,omitempty"`
	IOTimeoutSecs   Int              `json:"io_timeout_secs,omitempty"`
	Outputs         []string         `json:"outputs,omitempty"`
	SecretBytes     []byte           `json:"secret_bytes,omitempty"`
	Containment     Containment      `json:"containment,omitempty"`
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
	t.HardTimeoutSecs.Set64(int64(m.HardTimeout / time.Second))
	t.GracePeriodSecs.Set64(int64(m.GracePeriod / time.Second))
	t.Idempotent = m.Idempotent
	t.CASInputRoot.Host = m.CASHost
	t.CASInputRoot.Digest.FromDB(&m.Input)
	t.IOTimeoutSecs.Set64(int64(m.IOTimeout / time.Second))
	t.Outputs = m.Outputs
	// t.SecretBytes Never read back.
	t.Containment.ContainmentType.FromDB(m.Containment.ContainmentType)
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
	m.Dimensions = FromStringPairs(t.Dimensions)
	m.Env = FromStringPairs(t.Env)
	m.EnvPrefixes = FromStringListPairs(t.EnvPrefixes)
	m.HardTimeout = time.Duration(t.HardTimeoutSecs.Int64()) * time.Second
	m.GracePeriod = time.Duration(t.GracePeriodSecs.Int64()) * time.Second
	m.Idempotent = t.Idempotent
	m.CASHost = t.CASInputRoot.Host
	t.CASInputRoot.Digest.ToDB(&m.Input)
	m.IOTimeout = time.Duration(t.IOTimeoutSecs.Int64()) * time.Second
	m.Outputs = t.Outputs
	m.SecretBytes = t.SecretBytes
	m.Containment.ContainmentType = t.Containment.ContainmentType.ToDB()
	return nil
}

// TaskSlice defines one "option" to run the task.
type TaskSlice struct {
	Properties      TaskProperties `json:"properties,omitempty"`
	ExpirationSecs  Int            `json:"expiration_secs,omitempty"`
	WaitForCapacity bool           `json:"wait_for_capacity,omitempty"`
}

// FromDB converts the model to the API.
func (t *TaskSlice) FromDB(m *model.TaskSlice) {
	t.Properties.FromDB(&m.Properties)
	t.ExpirationSecs.Set64(int64(m.Expiration / time.Second))
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
	Enable bool `json:"enable,omitempty"`
}

// TaskRequest is a single requested task by a client. It is immutable.
type TaskRequest struct {
	//ExpirationSecs int64 `json:"expiration_secs"`
	Name         string       `json:"name,omitempty"`
	TaskID       model.TaskID `json:"task_id,omitempty"`
	ParentTaskID model.TaskID `json:"parent_task_id,omitempty"`
	Priority     Int          `json:"priority,omitempty"`
	//Properties TaskProperties `json:"properties,omitempty"`
	Tags                 []string    `json:"tags,omitempty"`
	Created              Time        `json:"created_ts,omitempty"`
	User                 string      `json:"user,omitempty"`
	Authenticated        string      `json:"authenticated,omitempty"`
	TaskSlices           []TaskSlice `json:"task_slices,omitempty"`
	ServiceAccount       string      `json:"service_account,omitempty"`
	Realm                string      `json:"realm,omitempty"`
	ResultDB             ResultDBCfg `json:"resultdb,omitempty"`
	PubSubTopic          string      `json:"pubsub_topic,omitempty"`
	PubSubUserData       string      `json:"pubsub_userdata,omitempty"`
	BotPingToleranceSecs Int         `json:"bot_ping_telerance_secs,omitempty"`
}

// FromDB converts the model to the API.
func (t *TaskRequest) FromDB(m *model.TaskRequest) {
	t.Name = m.Name
	t.TaskID = model.ToTaskID(m.Key)
	t.ParentTaskID = model.ToTaskID(m.ParentTask)
	t.Priority.Set32(m.Priority)
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
type TaskState string

// FromDB converts the model to the API.
func (t *TaskState) FromDB(m model.TaskState) {
	switch m {
	case model.Running:
		*t = "RUNNING"
	case model.Pending:
		*t = "PENDING"
	case model.Expired:
		*t = "EXPIRED"
	case model.Timedout:
		*t = "TIMED_OUT"
	case model.BotDied:
		*t = "BOT_DIED"
	case model.Canceled:
		*t = "CANCELED"
	case model.Completed:
		*t = "COMPLETED"
	case model.Killed:
		*t = "KILLED"
	case model.NoResource:
		*t = "NO_RESOURCE"
	default:
		*t = "BOT_DIED"
	}
}

// ResultDB declares the LUCI ResultDB information.
type ResultDB struct {
	Host       string `json:"hostname,omitempty"`
	Invocation string `json:"invocation,omitempty"`
}

// OperationStats is the statistic for one operation.
type OperationStats struct {
	DurationSecs float64 `json:"duration,omitempty"`
}

// CASOperationStats is RBE-CAS operation.
type CASOperationStats struct {
	Duration        float64 `json:"duration,omitempty"`
	InitialNumItems Int     `json:"initial_number_items,omitempty"`
	InitialSize     Int     `json:"initial_size,omitempty"`
	ItemsCold       []byte  `json:"items_cold,omitempty"` // zlib deflated varints.
	ItemsHot        []byte  `json:"items_hot,omitempty"`
	NumItemsCold    int     `json:"num_items_cold,omitempty"`
	NumItemsHot     int     `json:"num_items_hot,omitempty"`
}

// TaskPerfStats is the performance stats for a task.
type TaskPerfStats struct {
	BotOverheadSecs      float64           `json:"bot_overhead,omitempty"`
	CASDownload          CASOperationStats `json:"isolated_download,omitempty"`
	CASUpload            CASOperationStats `json:"isolated_upload,omitempty"`
	PkgInstallation      OperationStats    `json:"package_installation,omitempty"`
	CacheTrim            OperationStats    `json:"cache_trim,omitempty"`
	NamnedCachesInstall  OperationStats    `json:"named_caches_install,omitempty"`
	NamedCachesUninstall OperationStats    `json:"named_caches_uninstall,omitempty"`
	Cleanup              OperationStats    `json:"cleanup,omitempty"`
}

// FromDB converts the model to the API.
func (t *TaskPerfStats) FromDB() {
	// TODO(maruel): yo.
}

// TaskResult is the result of running a TaskRequest.
type TaskResult struct {
	Abandoned        Time             `json:"abandoned_ts,omitempty"`
	BotDimensions    []StringListPair `json:"bot_dimensions,omitempty"`
	BotID            string           `json:"bot_id,omitempty"`
	BotIdleSince     Time             `json:"bot_idle_since_ts,omitempty"`
	BotVersion       string           `json:"bot_version,omitempty"`
	Children         []model.TaskID   `json:"children_task_ids,omitempty"`
	Completed        Time             `json:"completed_ts,omitempty"`
	CostSavedUSD     float64          `json:"cost_saved_usd,omitempty"`
	Created          Time             `json:"created_ts,omitempty"`
	DedupedFrom      model.TaskID     `json:"deduped_from,omitempty"`
	DurationSecs     float64          `json:"duration,omitempty"`
	ExitCode         Int              `json:"exit_code,omitempty"`
	Failure          bool             `json:"failure,omitempty"`
	InternalFailure  bool             `json:"internal_failure,omitempty"`
	Modified         Time             `json:"modified_ts,omitempty"`
	CASOutput        CASReference     `json:"cas_output_root,omitempty"`
	ServerVersions   []string         `json:"server_versions,omitempty"`
	Started          Time             `json:"started_ts,omitempty"`
	State            TaskState        `json:"state,omitempty"`
	TaskID           model.TaskID     `json:"task_id,omitempty"`
	TryNumber        Int              `json:"try_number,omitempty"`
	CostsUSD         []float64        `json:"costs_usd,omitempty"`
	Name             string           `json:"name,omitempty"`
	Tags             []string         `json:"tags,omitempty"`
	User             string           `json:"user,omitempty"`
	Perf             *TaskPerfStats   `json:"performance_stats,omitempty"`
	CIPDPins         CIPDPins         `json:"cipd_pins,omitempty"`
	RunID            model.TaskID     `json:"run_id,omitempty"`
	CurrentTaskSlice Int              `json:"current_task_slice,omitempty"`
	ResultDB         ResultDB         `json:"resultdb_info,omitempty"`
}

// FromDB converts the model to the API.
func (t *TaskResult) FromDB(r *model.TaskRequest, m *model.TaskResult, includePerf bool) {
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
	t.ExitCode.Set32(m.ExitCode)
	t.Failure = m.ExitCode != 0
	t.InternalFailure = m.InternalFailure != ""
	t.Modified = CloudTime(m.Modified)
	// TODO(maruel): Use currentslice?
	t.CASOutput.Host = r.TaskSlices[0].Properties.CASHost
	t.CASOutput.Digest.FromDB(&m.Output)
	t.ServerVersions = m.ServerVersions
	t.Started = CloudTime(m.Started)
	t.State.FromDB(m.State)
	t.TaskID = model.ToTaskID(m.Key)
	t.TryNumber.Set32(0)
	if m.State != model.Pending {
		t.TryNumber.Set32(1)
	}
	t.CostsUSD = []float64{}
	t.Name = r.Name
	t.Tags = r.Tags
	t.User = r.User
	if includePerf {
		t.Perf = &TaskPerfStats{}
		t.Perf.FromDB()
	}
	t.CIPDPins.ClientPkg.FromDB(&m.CIPDClientUsed)
	t.CIPDPins.Pkgs = make([]CIPDPackage, len(m.CIPDPins))
	for i := range m.CIPDPins {
		t.CIPDPins.Pkgs[i].FromDB(&m.CIPDPins[i])
	}
	t.RunID = t.TaskID // No difference in mess.
	t.CurrentTaskSlice.Set32(m.CurrentTaskSlice)
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
	Dimensions []string `json:"dimensions,omitempty"`
	ValidUntil Time     `json:"valid_until_ts,omitempty"`
}
