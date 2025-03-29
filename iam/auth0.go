package iam

import (
	"bytes"
	"cerpApi/cfg_details"
	"cerpApi/psw_generator"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

var BLOCK_USER_REQ_BODY = bytes.NewReader([]byte(`{
	"blocked": true
}`))

func CreateAuth0User(body map[string]interface{}) string {
	fmt.Printf("Creating a user for %s", body)
	body["password"] = generateRandomPsw()
	url := cfg_details.API_URL + "/users"
	marshalled, err := json.Marshal(body)
	if err != nil {
		return ""
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(marshalled))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg_details.TOKEN)
	if err != nil {
		return ""
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	resBody, err := io.ReadAll(res.Body)
	respMap := make(map[string]interface{})
	err = json.Unmarshal(resBody, &respMap)
	if val, ok := respMap["statusCode"]; err != nil || (ok && val.(float64) > 300) {
		log.Println("User creation failed due to ", respMap["message"])
		return ""
	}
	return respMap["user_id"].(string)
}

func SetUserRoles(uId string, rErr *error, roles []byte) {
	fmt.Printf("Inside setUserRoles %s\n", uId)
	url := cfg_details.API_URL + "/users/" + uId + "/roles"
	fmt.Printf("Inside setUserRoles URL %s %s\n", url, roles)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(roles))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg_details.TOKEN)
	if err != nil {
		log.Println("Error in roles setting before call...")
		*rErr = err
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error in roles setting %v\n", err)
		*rErr = err
		return
	} else if res.StatusCode >= 300 {
		log.Println("Error in roles setting statuscode", res.StatusCode)
		*rErr = errors.New("Roles update failed!!!")
		return
	}
	log.Println("Exiting setUserRoles")
}

func DeactivateUser(uId string, uBy string) error {
	fmt.Printf("Deactivating a user for %s by %s", uId, uBy)
	url := cfg_details.API_URL + "/users/" + uId
	req, err := http.NewRequest(http.MethodDelete, url, BLOCK_USER_REQ_BODY)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg_details.TOKEN)
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode >= 300 {
		log.Println("User deactivation failed due to ", res.StatusCode)
		return err
	}
	fmt.Printf("Deactivated user for %s by %s\n", uId, uBy)
	return nil
}

func DeleteUser(uId string, uBy string) error {
	fmt.Printf("Deactivating a user for %s by %s", uId, uBy)
	url := cfg_details.API_URL + "/users/" + uId
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg_details.TOKEN)
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode >= 300 {
		log.Println("User deactivation failed due to ", res.StatusCode)
		return err
	}
	fmt.Printf("Deactivated user for %s by %s\n", uId, uBy)
	return nil
}

func generateRandomPsw() string {
	return psw_generator.GeneratePsw()
}
