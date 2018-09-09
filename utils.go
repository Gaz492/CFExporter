package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aviddiviner/go-murmur"
	"io/ioutil"
	"net/http"
	"os"
)

func readBuildJson(buildJsonDir string) (buildJson) {
	// Open our jsonFile
	jsonFile, err := os.Open(buildJsonDir)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var buildCfg buildJson
	json.Unmarshal(byteValue, &buildCfg)
	return buildCfg
}

func getProjectIds(addons []int) (*fingerprintResponse, error) {
	jsonPayload, _ := json.Marshal(addons)
	response, err := GetHTTPResponse("POST", PROXY_API+"fingerprint", jsonPayload)
	if err != nil {
		return nil, err
	}
	var addonResponse *fingerprintResponse
	err = json.NewDecoder(response.Body).Decode(&addonResponse)
	if err != nil {
		return nil, err
	}
	return addonResponse, nil
}

// Credit goes to modmuss https://github.com/modmuss50/CAV2/blob/master/murmur.go
func GetByteArrayHash(bytes []byte) int {
	return int(murmur.MurmurHash2(computeNormalizedArray(bytes), 1))
}

func GetFileHash(file string) (int, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}
	result := GetByteArrayHash(bytes)
	return result, nil
}

func computeNormalizedArray(bytes []byte) []byte {
	var newArray []byte
	for _, b := range bytes {
		if !isWhitespaceCharacter(b) {
			newArray = append(newArray, b)
		}
	}
	return newArray
}

func isWhitespaceCharacter(b byte) bool {
	return b == 9 || b == 10 || b == 13 || b == 32
}

func GetHTTPResponse(method, url string, b []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("authToken", *ProxyAuthToken)
	req.Header.Add("Content-Type", "application/json")
	return client.Do(req)

}