package main

import (
	"encoding/hex"
	"errors"
	"time"

	rbe "github.com/maruel/mess/third_party/build/bazel/remote/execution/v2"
)

const (
	maxHardTimeout = 7*24*time.Hour + 10*time.Second
	maxExpiration  = 7*24*time.Hour + 10*time.Second
	minHardTimeout = time.Second
	evictionCutOff = 550 * 24 * time.Hour
)

// Digest is a more memory efficient version of rbe.Digest.
type Digest struct {
	Size int64
	Hash [32]byte
}

func (d *Digest) toProto(p *rbe.Digest) {
	p.SizeBytes = d.Size
	p.Hash = hex.EncodeToString(d.Hash[:])
}

func (d *Digest) fromProto(p *rbe.Digest) error {
	d.Size = p.SizeBytes
	if len(p.Hash) != 64 {
		return errors.New("invalid hash")
	}
	// TODO: Manually decode for performance.
	_, err := hex.Decode(d.Hash[:], []byte(p.Hash))
	return err
}

type CIPDPackage struct {
	PkgName string
	Version string
	Path    string
}

type Cache struct {
	Name string
	Path string
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

type StringPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TaskProperties struct {
	Caches       []Cache
	Command      []string
	RelativeWD   string
	CASHost      string
	Input        Digest
	CIPDHost     string
	CIPDPackages []CIPDPackage
	Dimensions   []StringPair
	Env          []StringPair
	EnvPrefixes  []StringPair
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
