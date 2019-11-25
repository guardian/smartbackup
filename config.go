package main

import (
	"github.com/fredex42/smartbackup/mail"
	"github.com/fredex42/smartbackup/netapp"
	"github.com/fredex42/smartbackup/pagerduty"
	"github.com/fredex42/smartbackup/postgres"
)

type ConfigData struct {
	Netapp    []netapp.NetappConfig     `yaml:"netapp"`
	Databases []postgres.DatabaseConfig `yaml:"databases"`
	Targets   []BackupTarget            `yaml:"targets"`
	SMTP      mail.MailConfig           `yaml:"smtp"`
	PagerDuty pagerduty.PagerDutyConfig `yaml:"pagerduty"`
}

/**
resolves the names in the provided configuration and returns a list of ResolvedBackupTarget pointers
which contain pointers to descriptor objects for netapp, database, etc.
returns a list of pointers to ResolvedBackupTarget for targets that resolved and a list of pointers
to BackupTarget for targets that did not resolve
*/
func (c *ConfigData) ResolveBackupTargets() ([]*ResolvedBackupTarget, []*BackupTarget) {
	netAppMap := make(map[string]netapp.NetappConfig, len(c.Netapp))
	for _, entry := range c.Netapp {
		netAppMap[entry.Name] = entry
	}

	dataBaseMap := make(map[string]postgres.DatabaseConfig, len(c.Databases))
	for _, entry := range c.Databases {
		dataBaseMap[entry.Name] = entry
	}

	resolvedTargets := make([]*ResolvedBackupTarget, 0)
	failedTargets := make([]*BackupTarget, 0)

	for i, entry := range c.Targets {
		db, dbOk := dataBaseMap[entry.DatabaseName]
		np, netappOk := netAppMap[entry.NetappName]

		if dbOk == false || netappOk == false {
			//log.Printf("%s failed, incorrect db or netapp name", entry)
			failedTargets = append(failedTargets, &c.Targets[i])
		} else {
			//log.Printf("%s found", entry)
			resolvedTargets = append(resolvedTargets, &ResolvedBackupTarget{
				Database: &db,
				Netapp:   &np,
				VolumeId: entry.VolumeId,
			})
		}
	}

	return resolvedTargets, failedTargets
}
