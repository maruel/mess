package messapi

import (
	"encoding/hex"
	"time"

	"github.com/maruel/mess/internal/model"
)

type Digest struct {
	Hash string `json:"hash"`
	Size int64  `json:"size_bytes"`
}

func (d *Digest) ToDB(m *model.Digest) error {
	m.Size = d.Size
	_, err := hex.Decode(m.Hash[:], []byte(d.Hash))
	return err
}

func (d *Digest) FromDB(m *model.Digest) {
	d.Size = m.Size
	d.Hash = hex.EncodeToString(m.Hash[:])
}

type CIPDPackage struct {
	PkgName string `json:"package_name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

type CacheEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Store a reference to disk.
type TaskOutput struct {
	Size int64
}

type TasksCount struct {
	Count int32 `json:"count"`
	Now   Time  `json:"now"`
}

type TasksList struct {
	Cursor string       `json:"cursor"`
	Items  []TaskResult `json:"items"`
	Now    Time         `json:"now"`
}

type ContainmentType int

const (
	ContainmentNone ContainmentType = iota
	ContainmentAuto
	ContainmentJobObject
)

type Containment struct {
	LowerPriority   bool
	ContainmentType ContainmentType
}

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

type TaskSlice struct {
	Properties      TaskProperties
	Expiration      time.Duration
	WaitForCapacity bool
}

type BuildToken struct {
	BuildID         int64
	Token           string
	BuildbucketHost string
}

type TaskRequest struct {
	Key                 int64
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

func (t *TaskRequest) FromDB(m *model.TaskRequest) {
	panic("TODO")
}

type TaskState int64

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

type ResultDB struct {
	Host       string
	Invocation string
}

type TaskResult struct {
	Key            int64
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

func (t *TaskResult) FromDB(m *model.TaskResult) {
	panic("TODO")
}
