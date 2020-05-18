package netapp

import (
	"encoding/json"
	"testing"
	"time"
)

/**
ListsnapshotResponse should parse out of the example JSON provided
*/
func TestListSnapshotResponse_Parse(t *testing.T) {
	exampleRawJson := `{
  "records": [
    {
      "volume": {
        "_links": {
          "self": {
            "href": "/api/resourcelink"
          }
        },
        "name": "volume1",
        "uuid": "028baa66-41bd-11e9-81d5-00a0986138f7"
      },
      "_links": {
        "self": {
          "href": "/api/resourcelink"
        }
      },
      "name": "this_snapshot",
      "create_time": "2019-02-04T19:00:00Z",
      "uuid": "1cd8a442-86d1-11e0-ae1c-123478563412",
      "expiry_time": "2019-02-04T19:00:00Z",
      "state": "valid",
      "snaplock_expiry_time": "2019-02-04T19:00:00Z",
      "comment": "string",
      "svm": {
        "_links": {
          "self": {
            "href": "/api/resourcelink"
          }
        },
        "name": "svm1",
        "uuid": "02c9e252-41be-11e9-81d5-00a0986138f7"
      }
    }
  ],
  "_links": {
    "next": {
      "href": "/api/resourcelink"
    },
    "self": {
      "href": "/api/resourcelink"
    }
  },
  "num_records": 1
}`

	var resp ListSnapshotsResponse

	err := json.Unmarshal([]byte(exampleRawJson), &resp)
	if err != nil {
		t.Errorf("unmarshal returned unexpected error: %s", err)
		t.FailNow()
	}

	if resp.RecordsCount != 1 {
		t.Errorf("got invalid records count, expected 1 got %d", resp.RecordsCount)
	}

	if len(resp.Records) != 1 {
		t.Errorf("expected 1 record got %d", len(resp.Records))
		if len(resp.Records) < 1 {
			t.FailNow()
		}
	}

	rec := resp.Records[0]

	if rec.Name != "this_snapshot" {
		t.Error("got incorrect snapshot name")
	}
	expectedCreateTime, _ := time.Parse(time.RFC3339, "2019-02-04T19:00:00Z")
	if rec.CreateTime != expectedCreateTime {
		t.Error("got incorrect create time")
	}
	if rec.SnapshotId != "1cd8a442-86d1-11e0-ae1c-123478563412" {
		t.Error("got incorrect snapshot ID")
	}
	expectedExpiryTime, _ := time.Parse(time.RFC3339, "2019-02-04T19:00:00Z")
	if rec.ExpiryTime != expectedExpiryTime {
		t.Error("got incorrect expiry time")
	}
	if rec.State != "valid" {
		t.Error("got incorrect state")
	}
	if rec.SnaplockExpiryTime != expectedExpiryTime {
		t.Error("got incorrect snaplock expiry time")
	}
	if rec.Comment != "string" {
		t.Error("got incorrect comment")
	}
}
