package main

import (
	"errors"
	"io"
	"log"
	"regexp"
	"strings"
	"time"
)

const FailureMessage = `The backup for database {database:name} on {database:host} failed at {time} with the error {errorString}`
const FailureSubjectTemplate = `URGENT: {database:name} backup failed`
const SuccessMessage = `The backup for database {database:name} on {database:host} completed at {time}`
const SuccessSubjectTemplate = `{database:name} backed up`

//Note: ordering is important!
var DatabaseSubstitutionTags = []string{"database:name", "database:host", "database:port", "database:dbname"}
var GeneralSubstitutionTags = []string{"time", "errorString"}

type Messenger struct {
	CompiledDatabaseSubstitutions []*regexp.Regexp
	CompiledGeneralSubstitutions  []*regexp.Regexp
}

/**
initialise a new Messenger object, precompiling all regexes
*/
func NewMessenger() (*Messenger, error) {
	m := Messenger{}
	var err error
	m.CompiledDatabaseSubstitutions, err = compileRegexList(&DatabaseSubstitutionTags)
	if err != nil {
		return nil, err
	}

	m.CompiledGeneralSubstitutions, err = compileRegexList(&GeneralSubstitutionTags)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func compileRegexList(stringList *[]string) ([]*regexp.Regexp, error) {
	regexlist := make([]*regexp.Regexp, len(*stringList))

	for i, str := range *stringList {
		var compileErr error
		regexlist[i], compileErr = regexp.Compile(str)
		if compileErr != nil {
			return []*regexp.Regexp{}, nil
		}
	}
	return regexlist, nil
}

func (m *Messenger) GenerateMessage(target *ResolvedBackupTarget, subjectTemplateString string, templateString string, errorString string) (string, io.Reader, error) {
	//these must be in the same order as DatabseSubstitutionTags above!
	databaseSubValues := []string{target.Database.Name, target.Database.Host, string(target.Database.Port), target.Database.DBName}
	if len(databaseSubValues) < len(m.CompiledDatabaseSubstitutions) {
		log.Printf("ERROR: Not enough substitution values. This probably indicates a code bug.")
		return "", nil, errors.New("not enough substitution values")
	}

	var finalString = templateString
	var finalSubjectString = subjectTemplateString
	for ctr, re := range m.CompiledDatabaseSubstitutions {
		finalString = re.ReplaceAllString(finalString, databaseSubValues[ctr])
		finalSubjectString = re.ReplaceAllString(finalSubjectString, databaseSubValues[ctr])
	}

	generalSubValues := []string{time.Now().Format(time.RFC850), errorString}
	for ctr, re := range m.CompiledGeneralSubstitutions {
		finalString = re.ReplaceAllString(finalString, generalSubValues[ctr])
		finalSubjectString = re.ReplaceAllString(finalSubjectString, generalSubValues[ctr])
	}

	rdr := strings.NewReader(finalString)

	return finalSubjectString, rdr, nil
}
