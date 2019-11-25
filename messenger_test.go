package main

import (
	"github.com/fredex42/smartbackup/netapp"
	"github.com/fredex42/smartbackup/postgres"
	"testing"
)

/**
GenerateMessage should return a subjectline string and a body reader
unchanged if there are no substitutions
*/
func TestMessenger_GenerateMessageNosubs(t *testing.T) {
	fakeTarget := &ResolvedBackupTarget{
		Database: &postgres.DatabaseConfig{},
		Netapp:   &netapp.NetappConfig{},
		VolumeId: "",
	}
	m, err := NewMessenger()
	if err != nil {
		t.Fatalf("Could not set up Messenger: %s", err)
	}

	subject, bodyContent, generateErr := m.GenerateMessage(fakeTarget, "This is a subject", "This is a bodytext template", "This is an error string")

	if subject != "This is a subject" {
		t.Errorf("Got unexpected subject string '%s'", subject)
	}

	if bodyContent != "This is a bodytext template" {
		t.Errorf("Got unexpected bodytext '%s'", bodyContent)
	}

	if generateErr != nil {
		t.Errorf("Generate returned an unexpected error: %s", generateErr)
	}
}

/**
Generate message should add string substitutions to subjectline and body reader
but ignore invalid subs
*/
func TestMessenger_GenerateMessageWithsubs(t *testing.T) {
	fakeTarget := &ResolvedBackupTarget{
		Database: &postgres.DatabaseConfig{
			Name: "MyDatabase",
			Host: "db.company.int",
		},
		Netapp:   &netapp.NetappConfig{},
		VolumeId: "",
	}
	m, err := NewMessenger()
	if err != nil {
		t.Fatalf("Could not set up Messenger: %s", err)
	}

	subject, bodyContent, generateErr := m.GenerateMessage(fakeTarget, "This is backup {for} {database:name}", "This is a bodytext for {database:name} on {database:host} with error {errorString}", "This is an error string")

	if subject != "This is backup {for} MyDatabase" {
		t.Errorf("Got unexpected subject string '%s'", subject)
	}

	if bodyContent != "This is a bodytext for MyDatabase on db.company.int with error This is an error string" {
		t.Errorf("Got unexpected bodytext '%s'", bodyContent)
	}

	if generateErr != nil {
		t.Errorf("Generate returned an unexpected error: %s", generateErr)
	}
}
