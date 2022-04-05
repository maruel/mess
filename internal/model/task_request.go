package model

import (
	"encoding/hex"
	"errors"
	"time"

	rbe "github.com/maruel/mess/third_party/build/bazel/remote/execution/v2"
)

type TaskRequest struct {
	SchemaVersion int             `json:"a"`
	Key           int64           `json:"b"`
	Created       time.Time       `json:"c"`
	Priority      int             `json:"d"`
	ParentTask    int64           `json:"e"`
	Tags          []string        `json:"f"` // TODO(maruel): repeated values doesn't work in sqlite3?
	Blob          TaskRequestBlob `json:"g"`
}

// TaskRequestBlob contains the unindexed fields.
type TaskRequestBlob struct {
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

type ContainmentType int

const (
	ContainmentNone ContainmentType = iota
	ContainmentAuto
	ContainmentJobObject
)

type Containment struct {
	LowerPriority   bool            `json:"a"`
	ContainmentType ContainmentType `json:"b"`
}

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

type TaskSlice struct {
	Properties      TaskProperties `json:"a"`
	Expiration      time.Duration  `json:"b"`
	WaitForCapacity bool           `json:"c"`
}

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

func (d *Digest) ToProto(p *rbe.Digest) {
	p.SizeBytes = d.Size
	p.Hash = hex.EncodeToString(d.Hash[:])
}

func (d *Digest) FromProto(p *rbe.Digest) error {
	d.Size = p.SizeBytes
	if len(p.Hash) != 64 {
		return errors.New("invalid hash")
	}
	// TODO: Manually decode for performance.
	_, err := hex.Decode(d.Hash[:], []byte(p.Hash))
	return err
}

type CIPDPackage struct {
	PkgName string `json:"a"`
	Version string `json:"b"`
	Path    string `json:"c"`
}

type Cache struct {
	Name string `json:"a"`
	Path string `json:"b"`
}
