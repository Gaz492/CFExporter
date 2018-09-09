package main

type buildJson struct {
	PackAuthor string `json:"packAuthor"`
	MinecraftVersion string `json:"minecraftVersion"`
	ForgeVersion string `json:"forgeVersion"`
	ModsFolder string `json:"modsFolder"`
	Includes []string `json:"includes"`
}

type manifestBase struct {
	Minecraft manifestMinecraft `json:"minecraft"`
	ManifestType string `json:"manifestType"`
	ManifestVersion string `json:"manifestVersion"`
	Name string `json:"name"`
	Version string `json:"version"`
	Author string `json:"author"`
	Files []maifestFiles `json:"files"`
	overrides string `json:"overrides"`
}

type manifestMinecraft struct {
	Version string `json:"version"`
	ModLoaders []manifestMinecraftModLoaders `json:"modLoaders"`
}

type manifestMinecraftModLoaders struct {
	Id string `json:"id"`
	Primary bool `json:"primary"`
}

type maifestFiles struct {
	ProjectID int `json:"projectID"`
	FileID int `json:"fileID"`
	Required bool `json:"required"`
} 