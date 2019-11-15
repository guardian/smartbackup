package main

import (
	"github.com/fredex42/smartbackup/netapp"
	"github.com/fredex42/smartbackup/postgres"
)

type ConfigData struct {
	Netapp    []netapp.NetappConfig     `yaml:"netapp"`
	Databases []postgres.DatabaseConfig `yaml:"databases"`
	Targets   []BackupTarget            `yaml:"targets"`
}

/**
resolves the names in the provided configuration and returns a list of ResolvedBackupTarget pointers
which contain pointers to descriptor objects for netapp, database, etc.
*/
func (c *ConfigData) ResolveBackupTargets() []*ResolvedBackupTarget {
	netAppMap := make(map[string]*netapp.NetappConfig, len(c.Netapp))
	for _, entry := range c.Netapp {
		netAppMap[entry.Name] = &entry
	}

	dataBaseMap := make(map[string]*postgres.DatabaseConfig, len(c.Databases))
	for _, entry := range c.Databases {
		dataBaseMap[entry.Name] = &entry
	}

	resolvedTargets := make([]*ResolvedBackupTarget, len(c.Targets))
	for i, entry := range c.Targets {
		resolvedTargets[i] = &ResolvedBackupTarget{
			Database: dataBaseMap[entry.DatabaseName],
			Netapp:   netAppMap[entry.NetappName],
			VolumeId: entry.VolumeId,
		}
	}

	return resolvedTargets
}
