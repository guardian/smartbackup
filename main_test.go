package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/fredex42/smartbackup/netapp"
	"testing"
	"time"
)

/**
helper function to parse the time string provided or panic, similar to uuid.MustParse
*/
func timeMustParse(layout string, value string) time.Time {
	timeValue, err := time.Parse(layout, value)
	if err != nil {
		panic(err.Error())
	}
	return timeValue
}

func TestFindSnapshotsToDelete(t *testing.T) {
	//source data is deliberately jumbled up to test sorting
	initialSnapshotsList := &netapp.ListSnapshotsResponse{
		Records: []netapp.SnapshotEntry{
			{
				Name:       "snapshot 3",
				CreateTime: timeMustParse(time.RFC3339, "2019-03-04T11:12:13Z"),
				SnapshotId: "BEF03DF4-F85F-45CE-A3CA-E7C212C3B914",
			},
			{
				Name:       "snapshot 2",
				CreateTime: timeMustParse(time.RFC3339, "2019-02-03T11:12:13Z"),
				SnapshotId: "F1B37811-C171-4EA4-9DD0-FE453C687982",
			},
			{
				Name:       "snapshot 4",
				CreateTime: timeMustParse(time.RFC3339, "2019-04-05T11:12:13Z"),
				SnapshotId: "167E6B0B-37A1-4712-8BFC-185E61A076E8",
			},
			{
				Name:       "snapshot 1",
				CreateTime: timeMustParse(time.RFC3339, "2019-01-02T11:12:13Z"),
				SnapshotId: "FC776E0D-F21E-4695-B16A-E05F7454FC4F",
			},
			{
				Name:       "snapshot 5",
				CreateTime: timeMustParse(time.RFC3339, "2019-05-06T11:12:13Z"),
				SnapshotId: "233D7321-2DC3-4049-9504-194280996746",
			},
		},
		RecordsCount: 5,
	}

	backupTarget := &ResolvedBackupTarget{SnapshotsToKeep: 3}
	listToDelete := findSnapshotsToDelete(initialSnapshotsList, backupTarget)

	if len(listToDelete) != 2 {
		spew.Dump(listToDelete)
		t.Error("expected to have 2 items to delete but got ", len(listToDelete))
	}
	if len(listToDelete) < 2 {
		t.FailNow()
	}

	if listToDelete[0].Name != "snapshot 1" {
		t.Error("expected first item to be snapshot 1 but got ", listToDelete[0].Name)
	}
	if listToDelete[1].Name != "snapshot 2" {
		t.Error("expected second item to be snapshot 2 but got ", listToDelete[1].Name)
	}
}
