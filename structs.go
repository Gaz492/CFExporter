package main

type buildJson struct {
	PackAuthor string `json:"packAuthor"`
	MinecraftVersion string `json:"minecraftVersion"`
	ForgeVersion string `json:"forgeVersion"`
	ModsFolder string `json:"modsFolder"`
	Includes []string `json:"includes"`
}
