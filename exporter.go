package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
)

//type buildJson struct {
//	PackAuthor       string   `json:"packAuthor"`
//	MinecraftVersion string   `json:"minecraftVersion"`
//	ForgeVersion     string   `json:"forgeVersion"`
//	Includes         []string `json:"includes"`
//}

func main() {
	mcDirCLI := flag.String("d", "./", "Path to root folder of Minecraft instance")
	//pVerCLI := flag.String("p", "1.0.0", "Pack Version (e.g 1.0.0)")
	//exportName := flag.String("n", "Twitch-Export", "Export Name")
	buildConfig := flag.String("c", ".build.json", "Config file to get build variables")
	flag.Parse()
	//fmt.Println("Mc DIR:", *mcDirCLI)
	//fmt.Println("Pack Ver:", *pVerCLI)
	//fmt.Println("Export Name:", *exportName)
	//fmt.Println("Build Config:", *buildConfig)

	//fmt.Println(filepath.Abs(filepath.Dir(os.Args[0])))

	readBuildJson(*buildConfig)
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
	files, err := ioutil.ReadDir(path.Join(modsFolder, "mods"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".jar" {
			fmt.Println(f.Name())
		}
	}
}
