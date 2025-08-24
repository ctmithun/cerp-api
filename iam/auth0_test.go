package iam

import (
	"cerpApi/cfg_details"
	"encoding/json"
	"errors"
	"log"
	"testing"
)

func TestSetUserRoles(t *testing.T) {
	roleCounselor := map[string][]string{
		"roles": {cfg_details.FACULTY_ROLE},
	}
	rolesTmp, _ := json.Marshal(roleCounselor)
	var err error
	SetUserRoles("ni", "auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed", &err, rolesTmp)
}

func TestGetUserRoles(t *testing.T) {
	err := errors.New("")
	resp := GetUserRoles("ni", "auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed", &err)
	if resp == nil {
		t.Fail()
	}
	log.Println(resp)
}
