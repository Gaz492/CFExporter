package main

type (
	BuildJson struct {
		PackAuthor       string   `json:"packAuthor"`
		MinecraftVersion string   `json:"minecraftVersion"`
		ModLoader        string   `json:"modLoader"`
		ModLoaderVersion string   `json:"modLoaderVersion"`
		Includes         []string `json:"includes"`
	}

	FingerprintRequest struct {
		Fingerprints []int64 `json:"fingerprints"`
	}

	/*
		=================== Manifest ===================
	*/

	ExportManifest struct {
		Minecraft       Minecraft `json:"minecraft"`
		ManifestType    string    `json:"manifestType"`
		ManifestVersion int       `json:"manifestVersion"`
		Name            string    `json:"name"`
		Version         string    `json:"version"`
		Author          string    `json:"author"`
		Files           []Files   `json:"files"`
		Overrides       string    `json:"overrides"`
	}
	ModLoaders struct {
		ID      string `json:"id"`
		Primary bool   `json:"primary"`
	}
	Minecraft struct {
		Version    string       `json:"version"`
		ModLoaders []ModLoaders `json:"modLoaders"`
	}
	Files struct {
		ProjectID int  `json:"projectID"`
		FileID    int  `json:"fileID"`
		Required  bool `json:"required"`
	}

	/*
		=================== Twitch Fingerprint Response ===================
	*/

	FingerprintResponse struct {
		Data struct {
			IsCacheBuilt          bool                      `json:"isCacheBuilt"`
			ExactMatches          []FingerprintExactMatches `json:"exactMatches"`
			ExactFingerprints     []int64                   `json:"exactFingerprints"`
			InstalledFingerprints []int64                   `json:"installedFingerprints"`
		}
	}

	FingerprintExactMatches struct {
		Id   int             `json:"id"`
		File FingerprintFile `json:"file"`
	}

	FingerprintFile struct {
		Id              int    `json:"id"`
		FileName        string `json:"fileName"`
		FileNameOnDisk  string `json:"fileNameOnDisk"`
		FileDate        string `json:"fileDate"`
		ReleaseType     int    `json:"releaseType"`
		FileStatus      int    `json:"fileStatus"`
		DownloadUrl     string `json:"downloadUrl"`
		IsAlternate     bool   `json:"isAlternate"`
		AlternateFileId int    `json:"alternateFileId"`
		//Dependencies []fileDependencies `json:"dependencies"`
		IsAvailable        bool          `json:"isAvailable"`
		Modules            []FileModules `json:"modules"`
		PackageFingerprint int           `json:"packageFingerprint"`
		GameVersion        []string      `json:"gameVersion"`
	}

	FileModules struct {
		FolderName  string `json:"folderName"`
		Fingerprint int    `json:"fimgerprint"`
	}
)
