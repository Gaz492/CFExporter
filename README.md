# Twitch Minecraft Modpack Exporter

This tools was designed to allows you to create a Twitch Minecraft modpack export from other launchers such as MultiMC all you need to do s provide the path to the root of your Minecraft instance and follow the instructions.
## Usage
### Windows
Open command prompt in the same folder as `twitch-export-win.exe` and run the following command

`twitch-export-win.exe -d=<path to mc instance> -n=<export name> -p=<pack version> -c=<build config> -ct=<curseAuthenticationToken>`

Run `twitch-export-win.exe -h` for help

### Mac/Linux
Open terminal in the same folder as `twitch-export-linux` and run the following command

`twitch-export-linux -d=<path to mc instance> -n=<export name> -p=<pack version> -c=<build config> -ct=<curseAuthenticationToken>`

Run `twitch-export-linux -h` for help

### Curse Auth token
To get a curse authentication token you will need to send a post request to https://logins-v1.curseapp.net/login with the body of a Curse email and password. See https://logins-v1.curseapp.net/help for help
