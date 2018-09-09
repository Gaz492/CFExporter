package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
)

var (
	PROXY_API      = "https://curse.gaz492.uk/api/"
	ProxyAuthToken *string
	PackVersion    *string
	ExportName     *string
	BuildConfig    buildJson
)

func main() {
	mcDirCLI := flag.String("d", "./", "Path to root folder of Minecraft instance")
	PackVersion = flag.String("p", "1.0.0", "Pack Version (e.g 1.0.0)")
	ExportName = flag.String("n", "Twitch-Export", "Export Name")
	buildConfig := flag.String("c", ".build.json", "Config file to get build variables")
	ProxyAuthToken = flag.String("pt", "changeme", "Authentication token used to authenticate with Gaz's Twitch Proxy")
	flag.Parse()
	fmt.Println("Pack Ver:", *PackVersion)
	fmt.Println("Export Name:", *ExportName)

	BuildConfig = readBuildJson(*buildConfig)
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
		//fmt.Println(f.Name())
	}
}

func listMods(modsFolder string) {
	var jarFingerprints []int
	files, err := ioutil.ReadDir(path.Join(modsFolder, "mods"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".jar" {
			//fmt.Println(path.Join(modsFolder, "mods", f.Name()))
			//fmt.Println(GetFileHash(path.Join(modsFolder, "mods", f.Name())))
			fileHash, _ := GetFileHash(path.Join(modsFolder, "mods", f.Name()))
			jarFingerprints = append(jarFingerprints, fileHash)
			//fmt.Printf("%v %v\n", jarFingerprints, f.Name())
		}
	}

	//fMatchResp, _ := getProjectIds(jarFingerprints)

	//var test2 []fingerprintExactMatches
	//test2 = fMatchResp.ExactMatches
	//fmt.Println(test2[0].Id)


}

func createExport(projectFiles []manifestFiles) {
	var modloader []manifestMinecraftModLoaders
	var tempFiles []manifestFiles
	tempFiles = append(tempFiles, manifestFiles{1278361, 4522342, true})
	modloader = append(modloader, manifestMinecraftModLoaders{"forge-" + BuildConfig.ForgeVersion, true})
	manifestMc := manifestMinecraft{BuildConfig.MinecraftVersion, modloader}
	manifestB := manifestBase{manifestMc, "minecraftModpack", 1, *ExportName, *PackVersion, BuildConfig.PackAuthor, tempFiles, "overrides"}
	// test below
	addonsJson, _ := json.Marshal(manifestB)
	ioutil.WriteFile("output.json", addonsJson, 0644)
	//fmt.Printf("%+v", jsonResp)
}
