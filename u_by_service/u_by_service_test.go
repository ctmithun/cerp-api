package u_by_service

import (
	"cerpApi/cfg_details"
	"log"
	"testing"
)

func TestGetItems(t *testing.T) {
	uByMap := map[string]string{
		"auth0|ni|5c4ff8dd9ed0a67c2772c33e65f8ce22": cfg_details.NA,
		"auth0|ni|161ccfd2666790ef960b3a64fcc2c168": cfg_details.NA,
		"auth0|ni|790ef960b3a64fcc2c168":            cfg_details.NA,
	}
	err := getUByItems("ni_faculty", "ni_faculty", uByMap)
	if err != nil {
		t.Fail()
	}
	log.Println(uByMap)
}
