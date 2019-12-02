package main

import (
	"github.com/fredex42/smartbackup/netapp"
	"github.com/fredex42/smartbackup/postgres"
	"time"
)

type BackupTarget struct {
	DatabaseName string `yaml:"database"`
	NetappName   string `yaml:"netapp"`
	VolumeId     string `yaml:"volumeid"`
	KeepFor      string `yaml:"keep_for"`
}

type ResolvedBackupTarget struct {
	Database *postgres.DatabaseConfig
	Netapp   *netapp.NetappConfig
	VolumeId string
	KeepFor  time.Duration
}
