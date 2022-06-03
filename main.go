package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	instanceDir *string
	outputDir   *string
	configPath  *string
	// Optional
	exportName    *string
	exportVersion *string
	silent        *bool
	debug         *bool

	apiURL      = "https://api.curse.tools/v1/cf/"
	buildConfig BuildJson

	appVersion = "${VERSION}"

	tmpDir string
)

func init() {
	instanceDir = flag.String("d", "./", "Path to Minecraft instance")
	outputDir = flag.String("o", "./out", "Location to output export zip")
	configPath = flag.String("c", "./.build.json", "Path to .build.json")
	// Optional
	exportName = flag.String("n", "CurseForge-Export", "Name of the export")
	exportVersion = flag.String("v", "1.0.0", "Version of the export")
	// Other
	silent = flag.Bool("silent", false, "Silent output")
	debug = flag.Bool("debug", false, "Debug output")
	showHelp := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *debug {
		pterm.EnableDebugMessages()
		pterm.Debug.Prefix = pterm.Prefix{
			Text:  "DEBUG",
			Style: pterm.NewStyle(pterm.BgLightMagenta, pterm.FgBlack),
		}
		pterm.Debug.MessageStyle = pterm.NewStyle(98)
		pterm.Debug.Println("Debug mode enabled")
	}
	if *silent {
		pterm.Debug.Println("Silent mode enabled")
		pterm.DisableOutput()
	}

	logo, _ := pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("CF", pterm.NewStyle(pterm.FgYellow)),
		pterm.NewLettersFromStringWithStyle("Exporter", pterm.NewStyle(pterm.FgGray)),
	).Srender()
	pterm.DefaultCenter.Println(logo) // Print BigLetters with the default CenterPrinter
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(fmt.Sprintf("Version: %s", appVersion))

	pterm.DefaultSection.Println("Initializing...")

	buildConfig = readBuildJson(*configPath)

	if buildConfig.PackAuthor == "" {
		pterm.Error.Println("No author specified in .build.json")
		os.Exit(1)
	}
	if buildConfig.MinecraftVersion == "" {
		pterm.Error.Println("No Minecraft version specified in .build.json")
		os.Exit(1)
	}
	if buildConfig.ModLoader == "" {
		pterm.Error.Println("No mod loader specified in .build.json")
		os.Exit(1)
	}
	if buildConfig.MinecraftVersion == "" {
		pterm.Error.Println("No Minecraft version specified in .build.json")
		os.Exit(1)
	}

	pterm.Debug.Println("Config:", buildConfig)

	pterm.Success.Println("Initialized")
}

func main() {
	var err error
	tmpDir, err = os.MkdirTemp("", "cfExporter-*")
	if err != nil {
		pterm.Error.Println("Could not create temp dir:", err)
		os.Exit(1)
	}

	readInstanceDir()
}

func readInstanceDir() {
	pterm.Info.Println("Reading Minecraft instance...")
	files, err := os.ReadDir(*instanceDir)
	if err != nil {
		pterm.Error.Println("Failed to read Minecraft instance:", err)
		os.Exit(1)
	}

	fileProgress, _ := pterm.DefaultProgressbar.WithTotal(len(files)).WithTitle("Reading instance directory").Start()
	var foldersToScan []string
	for _, file := range files {
		fileProgress.UpdateTitle("Found: " + file.Name())
		fileProgress.Increment()
		if strings.ToLower(file.Name()) == "mods" || strings.ToLower(file.Name()) == "resourcepacks" {
			foldersToScan = append(foldersToScan, file.Name())
		}
	}
	scanFiles(foldersToScan)
}

func scanFiles(folders []string) {
	var fMatchResp *FingerprintResponse
	for _, folder := range folders {
		var fingerprints []int64
		pterm.Info.Println("Scanning " + folder + " directory...")
		files, err := os.ReadDir(path.Join(*instanceDir, folder))
		if err != nil {
			pterm.Error.Println("Failed to read "+folder+" directory\n", err)
			os.Exit(1)
		}

		fileProgress, _ := pterm.DefaultProgressbar.WithTotal(len(files)).WithTitle("Reading " + folder + " directory").Start()
		for _, file := range files {
			fileProgress.UpdateTitle("Found: " + file.Name())
			fileHash, _ := GetFileHash(path.Join(*instanceDir, folder, file.Name()))
			fingerprints = append(fingerprints, fileHash)
			fileProgress.Increment()
		}

		fMatchResp, err = getProjectIds(fingerprints)

		if fMatchResp != nil {
			pterm.Debug.Println("Missing matches:", difference(fMatchResp.Data.InstalledFingerprints, fMatchResp.Data.ExactFingerprints))
			genOverrides(difference(fMatchResp.Data.InstalledFingerprints, fMatchResp.Data.ExactFingerprints), folder)
		} else {
			pterm.Error.Println("Failed to get project IDs:", err)
			os.Exit(1)
		}
	}
	extraIncludes()
	genExport(fMatchResp.Data.ExactMatches)
}

func genOverrides(missingFiles []int64, folder string) {
	pterm.DefaultSection.Println("Generating overrides for " + folder + "...")

	if _, err := os.Stat(path.Join(tmpDir, "overrides")); os.IsNotExist(err) {
		err := os.Mkdir(path.Join(tmpDir, "overrides"), os.ModePerm)
		if err != nil {
			pterm.Error.Println("Failed to create overrides directory:", err)
		}
	}

	if len(missingFiles) > 0 {
		pterm.Debug.Println("Adding missing files to overrides:", missingFiles)
		if _, err := os.Stat(path.Join(tmpDir, "overrides", folder)); os.IsNotExist(err) {
			err := os.Mkdir(path.Join(tmpDir, "overrides", folder), os.ModePerm)
			if err != nil {
				pterm.Error.Println(fmt.Sprintf("Failed to create overrides/%s directory:", folder), err)
			}
		}

		files, err := os.ReadDir(path.Join(*instanceDir, folder))
		if err != nil {
			pterm.Error.Println("Failed to read "+folder+" directory\n", err)
			os.Exit(1)
		}
		fileProgress, _ := pterm.DefaultProgressbar.WithTotal(len(files)).WithTitle("Reading " + folder + " directory for overrides").Start()
		for _, file := range files {
			fileHash, _ := GetFileHash(path.Join(*instanceDir, folder, file.Name()))
			if intInSlice(fileHash, missingFiles) {
				fileProgress.UpdateTitle("Adding to overrides: " + file.Name())
				pterm.Debug.Println(fmt.Sprintf("Failed to find file %s on CurseForge - generating override", file.Name()))
				modSrc := path.Join(*instanceDir, folder, file.Name())
				err = CopyFile(modSrc, path.Join(tmpDir, "overrides", folder, file.Name()))
				if err != nil {
					pterm.Error.Println("Failed to copy file:", err)
				}
			}

			fileProgress.Increment()
		}
	}
}

func extraIncludes() {
	includeProgress, _ := pterm.DefaultProgressbar.WithTotal(len(buildConfig.Includes)).WithTitle("Adding extra includes to overrides").Start()
	for _, include := range buildConfig.Includes {
		if include != "mods" && include != "resourcepacks" {
			includeProgress.UpdateTitle("Adding: " + include + " to overrides")
			fToInclude := path.Join(*instanceDir, include)
			fi, err := os.Stat(fToInclude)
			if err != nil {
				includeProgress.UpdateTitle("Skipping adding: " + include + " to overrides")
				pterm.Warning.Println("Failed to read "+include+" directory\n", err)
				includeProgress.Increment()
				continue
			}
			switch mode := fi.Mode(); {
			case mode.IsDir():
				err = CopyDir(fToInclude, path.Join(tmpDir, "overrides", include))
				if err != nil {
					pterm.Error.Println("Failed to copy directory:", err)
				}
			case mode.IsRegular():
				err = CopyFile(fToInclude, path.Join(tmpDir, "overrides", include))
				if err != nil {
					pterm.Error.Println("Failed to copy file:", err)
				}
			}
		} else {
			pterm.Debug.Println("Skipping", include)
		}
		includeProgress.Increment()
	}
}

func genExport(projectFiles []FingerprintExactMatches) {
	pterm.DefaultSection.Println("Generating export...")
	var modLoader []ModLoaders
	var tempFiles []Files

	projectFilesProgress, _ := pterm.DefaultProgressbar.WithTotal(len(projectFiles)).WithTitle("Adding files to manifest").Start()
	for _, file := range projectFiles {
		projectFilesProgress.UpdateTitle("Adding " + strconv.Itoa(file.Id) + " to manifest")
		tempFiles = append(tempFiles, Files{ProjectID: file.Id, FileID: file.File.Id, Required: true})
		projectFilesProgress.Increment()
	}

	modLoader = append(modLoader, ModLoaders{buildConfig.ModLoader + "-" + buildConfig.ModLoaderVersion, true})
	minecraft := Minecraft{buildConfig.MinecraftVersion, modLoader}
	manifest := ExportManifest{minecraft, "minecraftModpack", 1, *exportName, *exportVersion, buildConfig.PackAuthor, tempFiles, "overrides"}
	manifestJson, err := json.Marshal(manifest)
	if err != nil {
		pterm.Error.Println("Failed to marshal manifest:", err)
		os.Exit(1)
	}
	err = os.WriteFile(path.Join(tmpDir, "manifest.json"), manifestJson, 0644)
	if err != nil {
		pterm.Error.Println("Failed to write manifest:", err)
		os.Exit(1)
	}
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		err := os.Mkdir(*outputDir, os.ModePerm)
		if err != nil {
			pterm.Error.Println("Failed to create output directory:", err)
		}
	}
	err = RecursiveZip(tmpDir, path.Join(*outputDir, *exportName+"-"+*exportVersion+".zip"))
	if err != nil {
		pterm.Error.Println("Failed to zip export:", err)
		os.Exit(1)
	}
	pterm.Success.Println("Export generated successfully\n", "Created zip: "+*exportName+"-"+*exportVersion+".zip")
	pterm.Info.Println("Removing temporary directory:", tmpDir)
	err = os.RemoveAll(tmpDir)
	if err != nil {
		pterm.Error.Println("Failed to remove temporary directory:", err)
	}
}
