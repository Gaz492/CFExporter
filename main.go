package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pterm/pterm"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

var (
	ApiUrl        string
	buildFile     *string
	mcDir         *string
	exportVersion *string
	exportName    *string
	BuildConfig   buildJson
	outputDir     *string
	verbose       *bool
)

func init() {
	buildFile = flag.String("c", ".build.json", "Config file to get build variables")
	mcDir = flag.String("d", "", "Path to root folder of Minecraft instance")
	exportVersion = flag.String("p", "1.0.0", "Pack Version (e.g 1.0.0)")
	exportName = flag.String("n", "Twitch-Export", "Export Name")
	outputDir = flag.String("o", "./", "Sets location for output files")
	verbose = flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	if *verbose {
		pterm.Debug.Prefix = pterm.Prefix{
			Text:  "DEBUG",
			Style: pterm.NewStyle(pterm.BgBlue, pterm.FgBlack),
		}
		pterm.Debug.MessageStyle = pterm.NewStyle(98)
		pterm.EnableDebugMessages()
		pterm.Info.Println("Verbose logging enabled")
		pterm.Debug.Println("Build file location:", *buildFile)
		pterm.Debug.Println("Minecraft instance location:", *mcDir)
		pterm.Debug.Println("Export name:", *exportName)
		pterm.Debug.Println("Export version:", *exportVersion)
		pterm.Debug.Println("Output Dir:", *outputDir)
	}
	ApiUrl = "https://addons-ecs.forgesvc.net/api/v2/"

	if err := validateBuildFile(); err != nil {
		pterm.Fatal.Println(err)
	}

	pterm.Debug.Println("Create tmp folder")
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		tmpDirErr := os.Mkdir("./tmp", 0755)
		if tmpDirErr != nil {
			pterm.Fatal.Println("Error creating tmp folder:", tmpDirErr)
		}
	}
}

func main() {
	instanceFiles := readMCDIR()
	matchedModsRaw := prepareExport(instanceFiles)
	createOverrides(Difference(matchedModsRaw.InstalledFingerprints, matchedModsRaw.ExactFingerprints))
	createExport(matchedModsRaw.ExactMatches)
}

func readMCDIR() (files []fs.FileInfo) {
	pterm.DefaultSection.Println("Reading Minecraft directory")
	mcDIRSpinner, _ := pterm.DefaultSpinner.Start("Reading directory...")
	files, err := ioutil.ReadDir(*mcDir)
	if err != nil {
		mcDIRSpinner.Fail()
		pterm.Fatal.Println("Error reading Minecraft directory:", err)
	}
	mcDIRSpinner.Success()

	if *verbose {
		for _, f := range files {
			pterm.Debug.Println(f.Name())
		}
	}
	return files
}

func prepareExport(instanceFiles []fs.FileInfo) (matchedModsRaw *fingerprintResponse) {
	pterm.DefaultSection.Println("Preparing export")

	peSpinner, _ := pterm.DefaultSpinner.Start("Scanning for includes")
	for _, include := range BuildConfig.Includes {
		if include == "mods" && fsContains(instanceFiles, "mods") {
			matchedModsRaw = scanMods(path.Join(*mcDir, "mods"))
			continue
		}
		if fsContains(instanceFiles, include) {
			pterm.Info.Println("Adding", include, "to overrides")
		} else {
			pterm.Warning.Println("Unable to find", include, "in instance files")
		}

	}
	peSpinner.Success()
	if fsContains(instanceFiles,"mods") && contains(BuildConfig.Includes, "mods") && matchedModsRaw != nil {
		pterm.Debug.Println("Following hashes were not matched:", Difference(matchedModsRaw.InstalledFingerprints, matchedModsRaw.ExactFingerprints))
	}
	return
}

func scanMods(modsDir string) *fingerprintResponse{
	var modFingerprints []int
	pterm.DefaultSection.Println("Scanning mods folder")

	// ! This causes prefix to duplicate output
	//modsSpinner, _ := pterm.DefaultSpinner.Start("Scanning for mods")
	files, err := ioutil.ReadDir(modsDir)
	if err != nil {
		pterm.Error.Println("Error reading mods dir:", err)
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".jar" {
			fileHash, err := GetFileHash(path.Join(modsDir, f.Name()))
			if err != nil {
				pterm.Error.Println(fmt.Sprintf("Error getting file hash (%s): %s", f.Name(), err))
			}else{
				pterm.Debug.Println(fmt.Sprintf("Found mod: %s(%s)", f.Name(), strconv.Itoa(fileHash)))
				modFingerprints = append(modFingerprints, fileHash)
			}
		}
	}
	//modsSpinner.Success()
	fMatchResp, err := getProjectIds(modFingerprints)
	if err != nil {
		pterm.Error.Println("Error with fetching mod project:", err)
	}
	return fMatchResp
}

func createOverrides(missingMods []int) {
	pterm.DefaultSection.Println("Creating Overrides")
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		tmpDirErr := os.Mkdir("./tmp", 0755)
		if tmpDirErr != nil {
			pterm.Error.Println(tmpDirErr)
		}
	}
	if _, err := os.Stat("./tmp/overrides"); os.IsNotExist(err) {
		overridesDirErr := os.Mkdir("./tmp/overrides", 0755)
		if overridesDirErr != nil {
			pterm.Error.Println(overridesDirErr)
		}
	}
	if _, err := os.Stat("./tmp/overrides/mods"); os.IsNotExist(err) {
		modsDirErr := os.Mkdir("./tmp/overrides/mods", 0755)
		if modsDirErr != nil {
			pterm.Error.Println(modsDirErr)
		}
	}

	files, err := ioutil.ReadDir(path.Join(*mcDir, "mods"))
	if err != nil {
		pterm.Fatal.Println(err)
	}
	if missingMods != nil {
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".jar" {
				fileHash, _ := GetFileHash(path.Join(*mcDir, "mods", f.Name()))
				if intInSlice(fileHash, missingMods) {
					pterm.Info.Println("Failed to find mod: " + strconv.Itoa(fileHash) + " " + f.Name() + " on CurseForge, adding to overrides")
					modSrc := path.Join(*mcDir, "mods", f.Name())
					cpModErr := CopyFile(modSrc, "./tmp/overrides/mods/"+f.Name())
					if cpModErr != nil {
						pterm.Error.Println(cpModErr)
					}
				}
			}
		}
	} else {
		pterm.Info.Println("No mods to override")
	}

	for _, includes := range BuildConfig.Includes {
		if includes != "mods" {
			pterm.Info.Println("Adding " + includes + " to overrides")
			fToInclude := path.Join(*mcDir, includes)
			fi, err := os.Stat(fToInclude)
			if err != nil {
				continue
			}
			switch mode := fi.Mode(); {
			case mode.IsDir():
				// do directory stuff
				cpDirErr := CopyDir(fToInclude, "./tmp/overrides/"+includes)
				if cpDirErr != nil {
					pterm.Error.Println(cpDirErr)
				}
			case mode.IsRegular():
				// do file stuff
				cpFileErr := CopyFile(fToInclude, "./tmp/overrides/"+includes)
				if cpFileErr != nil {
					pterm.Error.Println(cpFileErr)
				}
			}
		}
	}
}

func createExport(projectFiles []fingerprintExactMatches) {
	pterm.DefaultSection.Println("Creating Export Zip")
	var modloader []manifestMinecraftModLoaders
	var tempFiles []manifestFiles
	for _, file := range projectFiles {
		tempFiles = append(tempFiles, manifestFiles{file.Id, file.File.Id, true})
	}
	modloader = append(modloader, manifestMinecraftModLoaders{BuildConfig.ModLoader + "-" + BuildConfig.ModLoaderVersion, true})
	manifestMc := manifestMinecraft{BuildConfig.MinecraftVersion, modloader}
	manifestB := manifestBase{manifestMc, "minecraftModpack", 1, *exportName, *exportVersion, BuildConfig.PackAuthor, tempFiles, "overrides"}
	// test below
	addonsJson, _ := json.Marshal(manifestB)
	ioErr := ioutil.WriteFile("./tmp/manifest.json", addonsJson, 0644)
	if ioErr != nil {
		pterm.Error.Println(ioErr)
	}
	zipErr := RecursiveZip("./tmp", path.Join(*outputDir, *exportName+"-"+*exportVersion+".zip"))
	if zipErr != nil {
		pterm.Error.Println(zipErr)
	}
	pterm.Info.Println("Created zip: " + *exportName + "-" + *exportVersion + ".zip")
	pterm.Info.Println("Cleaning Up")
	rErr := os.RemoveAll("./tmp")
	if rErr != nil {
		pterm.Error.Println(rErr)
	}
}
