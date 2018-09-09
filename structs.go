package main

type buildJson struct {
	PackAuthor string `json:"packAuthor"`
	MinecraftVersion string `json:"minecraftVersion"`
	ForgeVersion string `json:"forgeVersion"`
	ModsFolder string `json:"modsFolder"`
	Includes []string `json:"includes"`
}

/*
=================== Twitch Export Manifest ===================
*/
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

/*
=================== Twitch Fingerprint Response ===================
*/

type fingerprintResponse struct {
	IsCacheBuilt bool `json:"isCacheBuilt"`
	ExactMatches []fingerprintExactMatches `json:"exactMatches"`
	ExactFingerprints []int `json:"exactFingerprints"`
	//PartialMatches []partialMatches `json:"partialMatches"`
	//PartialMatchFingerprints partialMatchFingerprints `json:"partialMatchFingerprints"`
	InstalledFingerprints []int `json:"installedFingerprints"`
}

type fingerprintExactMatches struct {
	Id int `json:"id"`
	File fingerprintFile `json:"file"`
	LatestFiles []fingerprintFile `json:"latestFiles"`
}

type fingerprintFile struct {
	Id int `json:"id"`
	FileName string `json:"fileName"`
	FileNameOnDisk string `json:"fileNameOnDisk"`
	FileDate string `json:"fileDate"`
	ReleaseType int `json:"releaseType"`
	FileStatus int `json:"fileStatus"`
	DownloadUrl string `json:"downloadUrl"`
	IsAlternate bool `json:"isAlternate"`
	AlternateFileId int `json:"alternateFileId"`
	//Dependencies []fileDependencies `json:"dependencies"`
	IsAvailable bool `json:"isAvailable"`
	Modules []fileModules `json:"modules"`
	PackageFingerprint int `json:"packageFingerprint"`
	GameVersion []int `json:"gameVersion"`
}

type fileModules struct {
	FolderName string `json:"folderName"`
	Fingerprint int `json:"fimgerprint"`
} 