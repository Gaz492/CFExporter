package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ApiUrl         string
	ProxyAuthToken *string
	CurseAuthToken *string
	PackVersion    *string
	ExportName     *string
	BuildConfig    buildJson
	PackDIR        string
)

func main() {
	mcDirCLI := flag.String("d", "./", "Path to root folder of Minecraft instance")
	PackVersion = flag.String("p", "1.0.0", "Pack Version (e.g 1.0.0)")
	ExportName = flag.String("n", "Twitch-Export", "Export Name")
	buildConfig := flag.String("c", ".build.json", "Config file to get build variables")
	ProxyAuthToken = flag.String("pt", "", "Proxy authentication token used to authenticate with Gaz's Twitch Proxy")
	CurseAuthToken = flag.String("ct", "", "Curse authentication token used to authenticate with the Curseforge/twitch API")
	flag.Parse()

	if *ProxyAuthToken == "" && *CurseAuthToken == "" {
		fmt.Println("Please enter a authentication token using -pt <toke>")
		os.Exit(1)
	}

	if *ProxyAuthToken != "" && *CurseAuthToken == "" {
		ApiUrl = "https://curse.gaz492.uk/api/"
		fmt.Println("Using: " + ApiUrl + " as API")
		fmt.Println(ProxyAuthToken)
	} else if *CurseAuthToken != "" && *ProxyAuthToken == "" {
		ApiUrl = "https://addons-ecs.forgesvc.net"
		fmt.Println("Using: " + ApiUrl + " as API")
		fmt.Println(CurseAuthToken)
	}

	BuildConfig = readBuildJson(*buildConfig)
	PackDIR = *mcDirCLI
	readMCDIR(*mcDirCLI)
}

func readMCDIR(dirPath string) {
	fmt.Println("Reading Minecraft Directory")
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if strings.ToLower(f.Name()) == "mods" {
			listMods(dirPath)
		}
	}
}

func listMods(modsFolder string) {
	fmt.Println("Listing Mods")
	var jarFingerprints []int
	files, err := ioutil.ReadDir(path.Join(modsFolder, "mods"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".jar" {
			fileHash, _ := GetFileHash(path.Join(modsFolder, "mods", f.Name()))
			jarFingerprints = append(jarFingerprints, fileHash)
		}
	}

	fMatchResp, _ := getProjectIds(jarFingerprints)
	if fMatchResp != nil {
		//fmt.Printf("Unable to find %v", Difference(fMatchResp.InstalledFingerprints, fMatchResp.ExactFingerprints))
		createOverrides(Difference(fMatchResp.InstalledFingerprints, fMatchResp.ExactFingerprints))
		createExport(fMatchResp.ExactMatches)
	} else {
		fmt.Println("Unable to read data exiting, contact maintainer")
		os.Exit(1)
	}

}

func createOverrides(missingMods []int) {
	fmt.Println("Creating Overrides")
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		os.Mkdir("./tmp", 0755)
	}
	if _, err := os.Stat("./tmp/overrides"); os.IsNotExist(err) {
		os.Mkdir("./tmp/overrides", 0755)
	}
	if _, err := os.Stat("./tmp/overrides/mods"); os.IsNotExist(err) {
		os.Mkdir("./tmp/overrides/mods", 0755)
	}

	files, err := ioutil.ReadDir(path.Join(PackDIR, "mods"))
	if err != nil {
		log.Fatal(err)
	}
	if missingMods != nil {
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".jar" {
				fileHash, _ := GetFileHash(path.Join(PackDIR, "mods", f.Name()))
				if intInSlice(fileHash, missingMods) {
					fmt.Println("Failed to find mod: " + strconv.Itoa(fileHash) + " " + f.Name() + " on CurseForge, adding to overrides")
					modSrc := path.Join(PackDIR, "mods", f.Name())
					CopyFile(modSrc, "./tmp/overrides/mods/"+f.Name())
				}
			}
		}
	} else {
		fmt.Println("Skipping Mods")
	}

	for _, includes := range BuildConfig.Includes {
		fmt.Println("Adding " + includes + " to overrides")
		fToInclude := path.Join(PackDIR, includes)
		fi, err := os.Stat(fToInclude)
		if err != nil {
			continue
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do directory stuff
			CopyDir(fToInclude, "./tmp/overrides/"+includes)
		case mode.IsRegular():
			// do file stuff
			CopyFile(fToInclude, "./tmp/overrides/"+includes)
		}
	}

}

func createExport(projectFiles []fingerprintExactMatches) {
	fmt.Println("Creating Export Zip")
	var modloader []manifestMinecraftModLoaders
	var tempFiles []manifestFiles
	for _, file := range projectFiles {
		tempFiles = append(tempFiles, manifestFiles{file.Id, file.File.Id, true})
	}
	modloader = append(modloader, manifestMinecraftModLoaders{"forge-" + BuildConfig.ForgeVersion, true})
	manifestMc := manifestMinecraft{BuildConfig.MinecraftVersion, modloader}
	manifestB := manifestBase{manifestMc, "minecraftModpack", 1, *ExportName, *PackVersion, BuildConfig.PackAuthor, tempFiles, "overrides"}
	// test below
	addonsJson, _ := json.Marshal(manifestB)
	ioutil.WriteFile("./tmp/manifest.json", addonsJson, 0644)
	RecursiveZip("./tmp", "./"+*ExportName+"-"+*PackVersion+".zip")
	fmt.Println("Created zip: " + *ExportName + "-" + *PackVersion + ".zip")
	fmt.Println("Cleaning Up")
	rErr := os.RemoveAll("./tmp")
	if rErr != nil {
		fmt.Println(rErr)
	}
}
