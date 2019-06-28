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
	ApiUrl      string
	PackVersion *string
	ExportName  *string
	BuildConfig buildJson
	PackDIR     string
	outputDir   *string
)

func main() {
	mcDirCLI := flag.String("d", "./", "Path to root folder of Minecraft instance")
	PackVersion = flag.String("p", "1.0.0", "Pack Version (e.g 1.0.0)")
	ExportName = flag.String("n", "Twitch-Export", "Export Name")
	buildConfig := flag.String("c", ".build.json", "Config file to get build variables")
	outputDir = flag.String("o", "./", "Sets location for output files")
	flag.Parse()

	ApiUrl = "https://addons-ecs.forgesvc.net/api/v2/"

	BuildConfig = readBuildJson(*buildConfig)
	if BuildConfig.PackAuthor == "" {
		fmt.Println("Invalid .build.json, Author not specified")
		os.Exit(1)
	} else if BuildConfig.MinecraftVersion == "" {
		fmt.Println("Invalid .build.json, Minecraft Version not specified")
		os.Exit(1)
	} else if BuildConfig.ModLoader == "" {
		fmt.Println("Invalid .build.json, Mod Loader not specified")
		os.Exit(1)
	} else if BuildConfig.ModLoaderVersion == "" {
		fmt.Println("Invalid .build.json, Mod Loader Version not specified")
		os.Exit(1)
	}
	PackDIR = *mcDirCLI
	readMCDIR(*mcDirCLI)
}

func readMCDIR(dirPath string) {
	fmt.Println("Reading Minecraft Directory")
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
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

	fMatchResp, err := getProjectIds(jarFingerprints)
	if fMatchResp != nil {
		//fmt.Printf("Unable to find %v", Difference(fMatchResp.InstalledFingerprints, fMatchResp.ExactFingerprints))
		createOverrides(Difference(fMatchResp.InstalledFingerprints, fMatchResp.ExactFingerprints))
		createExport(fMatchResp.ExactMatches)
	} else {
		fmt.Println("Unable to read data exiting, contact maintainer")
		if err != nil {
			log.Fatal(err)

		}
		os.Exit(1)
	}

}

func createOverrides(missingMods []int) {
	fmt.Println("Creating Overrides")
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		tmpDirErr := os.Mkdir("./tmp", 0755)
		if tmpDirErr != nil {
			fmt.Println(tmpDirErr)
		}
	}
	if _, err := os.Stat("./tmp/overrides"); os.IsNotExist(err) {
		overridesDirErr := os.Mkdir("./tmp/overrides", 0755)
		if overridesDirErr != nil {
			fmt.Println(overridesDirErr)
		}
	}
	if _, err := os.Stat("./tmp/overrides/mods"); os.IsNotExist(err) {
		modsDirErr := os.Mkdir("./tmp/overrides/mods", 0755)
		if modsDirErr != nil {
			fmt.Println(modsDirErr)
		}
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
					cpModErr := CopyFile(modSrc, "./tmp/overrides/mods/"+f.Name())
					if cpModErr != nil {
						fmt.Println(cpModErr)
					}
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
			cpDirErr := CopyDir(fToInclude, "./tmp/overrides/"+includes)
			if cpDirErr != nil {
				fmt.Println(cpDirErr)
			}
		case mode.IsRegular():
			// do file stuff
			cpFileErr := CopyFile(fToInclude, "./tmp/overrides/"+includes)
			if cpFileErr != nil {
				fmt.Println(cpFileErr)
			}
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
	modloader = append(modloader, manifestMinecraftModLoaders{BuildConfig.ModLoader + "-" + BuildConfig.ModLoaderVersion, true})
	manifestMc := manifestMinecraft{BuildConfig.MinecraftVersion, modloader}
	manifestB := manifestBase{manifestMc, "minecraftModpack", 1, *ExportName, *PackVersion, BuildConfig.PackAuthor, tempFiles, "overrides"}
	// test below
	addonsJson, _ := json.Marshal(manifestB)
	ioErr := ioutil.WriteFile("./tmp/manifest.json", addonsJson, 0644)
	if ioErr != nil {
		fmt.Println(ioErr)
	}
	zipErr := RecursiveZip("./tmp", path.Join(*outputDir, *ExportName+"-"+*PackVersion+".zip"))
	if zipErr != nil {
		fmt.Println(zipErr)
	}
	fmt.Println("Created zip: " + *ExportName + "-" + *PackVersion + ".zip")
	fmt.Println("Cleaning Up")
	rErr := os.RemoveAll("./tmp")
	if rErr != nil {
		fmt.Println(rErr)
	}
}
