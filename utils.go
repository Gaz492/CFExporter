package main

import (
	"encoding/json"
	"fmt"
	"github.com/aviddiviner/go-murmur"
	"io/ioutil"
	"os"
)

func readBuildJson(buildJsonDir string) {
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
	fmt.Println(buildCfg.ForgeVersion)
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
