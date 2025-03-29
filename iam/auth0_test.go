package iam

import (
	"cerpApi/cfg_details"
	"encoding/json"
	"testing"
)

func TestSetUserRoles(t *testing.T) {
	roleCounselor := map[string][]string{
		"roles": {cfg_details.COUNSELOR_ROLE},
	}
	rolesTmp, _ := json.Marshal(roleCounselor)
	var err error
	SetUserRoles("auth0|ni|2757128944ce3cd9a281bbdc3fc2e4db", &err, rolesTmp)
}
