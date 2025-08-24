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
	"net/url"
)

var BLOCK_USER_REQ_BODY = bytes.NewReader([]byte(`{
	"blocked": true
}`))

func CreateAuth0User(body map[string]interface{}, colId string) string {
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
	tkn, err := cfg_details.GetSecretCfg("token", colId, "ap-south-1")
	if err != nil {
		log.Println("User creation failed due to token unavailability")
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+tkn)
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

func SetUserRoles(colId string, uId string, rErr *error, roles []byte) {
	fmt.Printf("Inside setUserRoles %s\n", uId)
	enUId := url.QueryEscape(uId)
	url := cfg_details.API_URL + "/users/" + enUId + "/roles"
	fmt.Printf("Inside setUserRoles URL %s %s\n", url, roles)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(roles))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	tkn, err := cfg_details.GetSecretCfg("token", colId, "ap-south-1")
	if err != nil {
		log.Println("User roles failed due to token unavailability")
		return
	}
	req.Header.Set("Authorization", "Bearer "+tkn)
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

func DeleteUserRoles(colId string, uId string, rErr *error, roles []byte) {
	fmt.Printf("Inside deleteUserRoles %s\n", uId)
	enUId := url.QueryEscape(uId)
	url := cfg_details.API_URL + "/users/" + enUId + "/roles"
	fmt.Printf("Inside deleteUserRoles URL %s %s\n", url, roles)
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(roles))
	if err != nil {
		*rErr = err
		log.Printf("Error while creating delete req")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	tkn, err := cfg_details.GetSecretCfg("token", colId, cfg_details.CFG.Region)
	if err != nil {
		log.Println("User roles failed due to token unavailability")
		return
	}
	req.Header.Set("Authorization", "Bearer "+tkn)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error in roles setting %v\n", err)
		*rErr = err
		return
	} else if res.StatusCode >= 300 {
		log.Println("Error in roles setting statuscode", res.StatusCode)
		*rErr = errors.New("roles delete failed")
		return
	}
	log.Println("Exiting deleteUserRoles")
}

func GetUserRoles(colId string, uId string, rErr *error) map[string]interface{} {
	fmt.Printf("Inside GetUserRoles %s\n", uId)
	enUId := url.QueryEscape(uId)
	url := cfg_details.API_URL + "/users/" + enUId + "/roles"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error creating req %v\n", err)
		*rErr = err
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	tkn, err := cfg_details.GetSecretCfg("token", colId, cfg_details.CFG.Region)
	if err != nil {
		log.Println("User roles failed due to token unavailability")
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+tkn)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error in getting the existing roles rest call %v\n", err)
		*rErr = err
		return nil
	} else if res.StatusCode >= 300 {
		log.Println("Error in roles setting statuscode", res.StatusCode)
		*rErr = errors.New("getting existing roles failed response status > 300")
		return nil
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reasing body from response %v\n", err)
		*rErr = err
		return nil
	}
	var respMap []map[string]string
	err = json.Unmarshal(body, &respMap)
	if err != nil {
		log.Printf("Error Unmarshalling the resp %v\n", err)
		*rErr = err
		return nil
	}
	respMapIds := make(map[string]interface{})
	for i := 0; i < len(respMap); i++ {
		respMapIds[respMap[i]["id"]] = nil
	}
	return respMapIds
}

func DeactivateUser(colId string, uId string, uBy string) error {
	fmt.Printf("Deactivating a user for %s by %s", uId, uBy)
	url := cfg_details.API_URL + "/users/" + uId
	req, err := http.NewRequest(http.MethodDelete, url, BLOCK_USER_REQ_BODY)
	if err != nil {
		log.Printf("Error while deactivating user %s %v\n", uId, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	tkn, err := cfg_details.GetSecretCfg("token", colId, "ap-south-1")
	if err != nil {
		log.Println("User creation failed due to token unavailability")
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tkn)
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
