package main

import (
	"github.com/fredex42/smartbackup/netapp"
	"github.com/fredex42/smartbackup/postgres"
)

type BackupTarget struct {
	DatabaseName    string `yaml:"database"`
	NetappName      string `yaml:"netapp"`
	VolumeId        string `yaml:"volumeid"`
	SnapshotsToKeep int    `yaml:"max_snapshots"`
}

type ResolvedBackupTarget struct {
	Database        *postgres.DatabaseConfig
	Netapp          *netapp.NetappConfig
	VolumeId        string
	SnapshotsToKeep int
}
