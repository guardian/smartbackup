package main

import (
	"github.com/fredex42/smartbackup/netapp"
	"github.com/fredex42/smartbackup/postgres"
	"testing"
)

/**
resolveBackupTargets should return seperate lists of resolved and unresolved targs
*/
func TestResolveBackupTargets(t *testing.T) {
	testConfig := ConfigData{
		Netapp: []netapp.NetappConfig{
			{Name: "test1", SVM: "svm1", Host: "hostname1", User: "username1", Passwd: "password1"},
			{Name: "test2", SVM: "svm2", Host: "hostname2", User: "username2", Passwd: "password2"},
			{Name: "test3", SVM: "svm3", Host: "hostname3", User: "username3", Passwd: "password3"},
		},
		Databases: []postgres.DatabaseConfig{
			{Name: "database1", Host: "dbhost1", Port: 5432, DBName: "somedb", User: "user1", Password: "password1"},
			{Name: "database2", Host: "dbhost2", Port: 5432, DBName: "somedb", User: "user2", Password: "password2"},
		},
		Targets: []BackupTarget{
			{DatabaseName: "database2", NetappName: "test3", VolumeId: "vol1"},
			{DatabaseName: "database95", NetappName: "test1", VolumeId: "vol2"},
			{DatabaseName: "database1", NetappName: "test1", VolumeId: "vol3"},
			{DatabaseName: "database1", NetappName: "test99", VolumeId: "vol4"},
		},
	}

	resolved, unresolved := testConfig.ResolveBackupTargets()

	if len(resolved) != 2 {
		t.Errorf("Expected 2 resolved targets, got %d", len(resolved))
	}
	if len(unresolved) != 2 {
		t.Errorf("Expected 2 unresolved targets, got %d", len(unresolved))
	}

	if resolved[0].Netapp.Name != "test3" {
		t.Errorf("Resolved target 1 had incorrect netapp pointer %s", resolved[0].Netapp)
	}
	if resolved[1].Netapp.Name != "test1" {
		t.Errorf("Resolved target 2 had incorrect netapp pointer %s", resolved[1].Netapp)
	}
	if unresolved[0].DatabaseName != "database95" {
		t.Errorf("Unresolved target 1 had incorrect name %s, was expecting database95", unresolved[0].DatabaseName)
	}
	if unresolved[1].NetappName != "test99" {
		t.Errorf("Unresolved target 2 had incorred netapp %s, was expecting test99", unresolved[1].NetappName)
	}
}
