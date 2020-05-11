package dtclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// TokenScopes is a list of scopes assigned to a token
type TokenScopes []string

// Contains returns true if scope is included on the scopes, or false otherwise.
func (s TokenScopes) Contains(scope string) bool {
	for _, x := range s {
		if x == scope {
			return true
		}
	}
	return false
}

func (dc *dynatraceClient) CreatePaaSToken(token string) (string, error) {
	var model struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	model.Name = "dynatrace-oneagent-operator-paas-token"
	model.Scopes[0] = TokenScopeInstallerDownload

	jsonStr, err := json.Marshal(model)
	if err != nil {
		return "", err
	}

	data, err := dc.tokenAPIRequest(token, "v1/tokens/createToken", jsonStr)
	if err != nil {
		return "", err
	}

	return dc.readResponseForTokenCreation(data)
}

func (dc *dynatraceClient) GetTokenScopes(token string) (TokenScopes, error) {
	var model struct {
		Token string `json:"token"`
	}
	model.Token = token

	jsonStr, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}

	data, err := dc.tokenAPIRequest(token, "v1/tokens/lookup", jsonStr)
	if err != nil {
		return nil, err
	}

	return dc.readResponseForTokenScopes(data)
}

func (dc *dynatraceClient) readResponseForTokenScopes(response []byte) (TokenScopes, error) {
	var jr struct {
		Scopes []string `json:"scopes"`
	}

	if err := json.Unmarshal(response, &jr); err != nil {
		return nil, fmt.Errorf("error unmarshalling json response: %w", err)
	}

	return jr.Scopes, nil
}

func (dc *dynatraceClient) tokenAPIRequest(token string, endpoint string, jsonStr []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", dc.url, endpoint), bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("error initialising http request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Api-Token %s", token))

	resp, err := dc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making post request to dynatrace api: %w", err)
	}
	defer resp.Body.Close()

	data, err := dc.getServerResponseData(resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dc *dynatraceClient) readResponseForTokenCreation(response []byte) (string, error) {
	var jr struct {
		PaasToken string `json:"token"`
	}

	if err := json.Unmarshal(response, &jr); err != nil {
		return "", fmt.Errorf("error unmarshalling json response: %w", err)
	}

	return jr.PaasToken, nil
}
