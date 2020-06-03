package netapp

import "time"

type NetappConfig struct {
	Host   string `yaml:"host"`
	User   string `yaml:"user"`
	Passwd string `yaml:"password"`
	SVM    string `yaml:"svm"`
	Name   string `yaml:"name"`
}

type NetappEntity struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

type NetappEntityUuid struct {
	UUID string `json:"uuid"`
}

type CreateSnapshotRequest struct {
	Name string       `json:"name"`
	SVM  NetappEntity `json:"svm"`
}

type JobResponse struct {
	UUID string `json:"uuid"`
}

type CreateSnapshotResponse struct {
	Job JobResponse `json:"job"`
}

type DeleteSnapshotResponse struct {
	Job JobResponse `json:"job"`
}

type SnapshotEntry struct {
	Name               string    `json:"name"`
	CreateTime         time.Time `json:"create_time"`
	SnapshotId         string    `json:"uuid"`
	ExpiryTime         time.Time `json:"expiry_time"`
	State              string    `json:"state"`
	SnaplockExpiryTime time.Time `json:"snaplock_expiry_time"`
	Comment            string    `json:"comment"`
}

type ListSnapshotsResponse struct {
	Records      []SnapshotEntry `json:"records"`
	RecordsCount int32           `json:"num_records"`
}

type ErrorArguments struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type Error struct {
	Target    string           `json:"target"` //should be a UUID
	Arguments []ErrorArguments `json:"arguments"`
	Message   string           `json:"message"`
	Code      string           `json:"code"`
}

type ErrorResponse struct {
	Error Error `json:"error"`
}

type NetappJob struct {
	StartTime   string `json:"start_time"`
	UUID        string `json:"uuid"`
	Description string `json:"description"`
	State       string `json:"state"`
	Message     string `json:"message"`
	EndTime     string `json:"end_time"`
	Code        int64  `json:"code"`
}
