# Usage Guide

## Step 1
You will need to create a config file named `.build.json`.
This file contains the settings and values that will be used to create the export.

You can use the following template for the `.build.json`

```json
{
    "packAuthor": "FTB",
    "minecraftVersion": "1.12.2",
    "forgeVersion": "14.23.4.2756",
    "includes": ["config", "options.txt", "map", "resources"]
}
```

| Object | Type | Value |
| ------ | ---- | ----- |
| `packAuthor` | `String` | Sets the author of the pack |
| `minecraftVersion` | `String` | Sets the minecraft version to use |
| `forgeVersion` | `String` | Defines the forge version to use |
| `includes` | `String Array` | Array defining the files/folders to include in the export |

---
## Step 2
Use the following command to run the exporter tool

### Windows
Open command prompt in the same folder as `twitch_export-win.exe` and run the following command

`twitch_export-win.exe -d "<path to mc instance>" -n "<export name>" -p "<pack version>" -c "<path to .build.json>" -ct "<curseAuthenticationToken>"`

Run `twitch_export-win.exe -h` for help

### Mac/Linux
Open terminal in the same folder as `twitch-export-linux` and run the following command

`twitch_export-linux -d=<path to mc instance> -n=<export name> -p=<pack version> -c=<path to .build.json> -ct=<curseAuthenticationToken>`

Run `twitch_export-linux -h` for help

### Curse Auth token
To get a curse authentication token you will need to send a post request to https://logins-v1.curseapp.net/login with the body of a Curse email and password. See https://logins-v1.curseapp.net/help for help


---
# Issues

If you are having any issues please create a new issue
