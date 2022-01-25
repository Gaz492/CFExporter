package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aviddiviner/go-murmur"
	"github.com/pterm/pterm"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func readBuildJson(buildJsonDir string) BuildJson {
	// Open our jsonFile
	jsonFile, err := os.Open(buildJsonDir)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var buildCfg BuildJson
	err = json.Unmarshal(byteValue, &buildCfg)
	if err != nil {
		pterm.Fatal.Println("Failed to parse build.json\n", err)
		os.Exit(1)
	}
	return buildCfg
}

func intInSlice(a int64, list []int64) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func difference(a, b []int64) (diff []int64) {
	m := make(map[int64]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func getFileCount(path string) int64 {
	var count int64
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		count++
		return nil
	})
	if err != nil {
		pterm.Error.Println("Error counting files:", err)
	}
	return count
}

func RecursiveZip(pathToZip, destinationPath string) error {
	fileCount := getFileCount(pathToZip)
	zipProgress, _ := pterm.DefaultProgressbar.WithTotal(int(fileCount)).WithTitle("Creating zip file").Start()
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		zipProgress.UpdateTitle("Adding " + info.Name() + " to zip")
		if info.IsDir() {
			pterm.Debug.Println("info.IsDir is Nil")
			zipProgress.Increment()
			return nil
		}
		if err != nil {
			pterm.Error.Println("FilePath walk error")
			zipProgress.Stop()
			return err
		}

		relPath := strings.TrimPrefix(filepath.ToSlash(filePath), filepath.ToSlash(pathToZip+"/"))
		zipFile, err := myZip.Create(relPath)
		if err != nil {
			pterm.Error.Println("Creating zip error")
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			pterm.Error.Println("os.Open error")
			return err
		}
		defer fsFile.Close()
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			pterm.Error.Println("io.copy error")
			return err
		}
		zipProgress.Increment()
		return nil
	})
	if err != nil {
		pterm.Error.Println("File path walk error ln151")
		return err
	}
	err = myZip.Close()
	if err != nil {
		pterm.Error.Println("Close zip error")
		return err
	}
	return nil
}

func getProjectIds(addons []int64) (*FingerprintResponse, error) {
	jsonPayload, _ := json.Marshal(FingerprintRequest{Fingerprints: addons})
	response, err := GetHTTPResponse("POST", apiURL+"fingerprints", jsonPayload)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	var addonResponse *FingerprintResponse
	err = json.NewDecoder(response.Body).Decode(&addonResponse)
	if err != nil {
		return nil, err
	}
	return addonResponse, nil
}

// Credit goes to modmuss https://github.com/modmuss50/CAV2/blob/master/murmur.go
func GetByteArrayHash(bytes []byte) int64 {
	return int64(murmur.MurmurHash2(computeNormalizedArray(bytes), 1))
}

func GetFileHash(file string) (int64, error) {
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

	req.Header.Add("User-Agent", fmt.Sprintf("CFExporter-%s", appVersion))
	req.Header.Add("Content-Type", "application/json")
	return client.Do(req)
}
