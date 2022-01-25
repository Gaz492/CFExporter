CurseforgeExporter

# Usage Guide

## Step 1
You will need to create a config file named `.build.json`.
This file contains the settings and values that will be used to create the export.

You can use the following template for the `.build.json`

```json
{
    "packAuthor": "FTB",
    "minecraftVersion": "1.12.2",
    "modLoader": "forge",
    "modLoaderVersion": "14.23.4.2756",
    "includes": ["config", "options.txt", "map", "resources"]
}
```

| Object             | Type           | Value                                                                                                 |
|--------------------|----------------|-------------------------------------------------------------------------------------------------------|
| `packAuthor`       | `String`       | Sets the author of the pack                                                                           |
| `minecraftVersion` | `String`       | Sets the minecraft version to use                                                                     |
| `modLoader`        | `string`       | Defines what mod loader to use                                                                        |
| `modLoaderVersion` | `String`       | Defines the mod loader version to use                                                                 |
| `includes`         | `String Array` | Array defining the files/folders to include in the export, you should not add the mods folder to this |

---
## Step 2
Use the following command to run the exporter tool

### Flags

| Flag      | Default             | Description                   |
|-----------|---------------------|-------------------------------|
| `-d`      | `./`                | Path to Minecraft instance    |
| `-o`      | `./out`             | Location to output export zip |
| `-c`      | `./.build.json`     | Path to .build.json           |
| `-n`      | `CurseForge-Export` | Name of export                |
| `-v`      | `1.0.0`             | Version of the export         |
  | `-silent` | `false`             | Enable silent output          |
| `-debug`  | `false`             | Enable debug logging          |
| `-help`   | `false`             | Shows help text               |

### Windows
Open command prompt in the same folder as `CFExporter.exe` and run the following command

`CFExporter.exe -d "<path to mc instance>" -c "<path to .build.json>"`

Run `twitch_export-win.exe -help` for help

### Mac/Linux
Open terminal in the same folder as `CFExporter` and run the following command

`CFExporter -d "<path to mc instance>" -c "<path to .build.json>"`

Run `CFExporter -h` for help

### Curse Auth token
To get a curse authentication token you will need to send a post request to https://logins-v1.curseapp.net/login with the body of a Curse email and password. See https://logins-v1.curseapp.net/help for help


---
# Issues

If you are having any issues please create a new issue
