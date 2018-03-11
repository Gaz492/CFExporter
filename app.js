const program = require('commander');
const inquirer = require('inquirer');
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const request = require('request');
const rimraf = require('rimraf');
const ncp = require('ncp');
const archiver = require('archiver');
const pJson = require('./package');

const directories = {
    base: './',
    export: {
        root: 'temp',
        overrides: 'overrides',
        mods: 'mods'
    },
    meta: 'meta'
};

const appVersion = pJson['version'];

const apiURL = 'https://curse.gaz492.uk';
const questions = [];
let packName;
let packVersion;
let packAuthor;
let mcVersion;
let forgeVersion;
let curseJson;

let projectObj = [];
let modList = [];
let foundMods = [];
let copyList = [];


if (fs.existsSync(path.join(directories.base, directories.export.root))) {
    rimraf(path.join(directories.base, directories.export.root), (err) => {
        if (err) return console.log(err);
    })
}

function getCurseMeta() {
    if(!fs.existsSync(path.join(directories.base, directories.meta))){
        fs.mkdirSync(path.join(directories.base, directories.meta))
    }
    let jsonMD5Hash;
    let md5Options = {
        url: apiURL + '/curseProjects.json.md5',
        method: 'GET',
        headers: {
            'User-Agent': 'Twitch-Exporter/' + appVersion + ' (+https://github.com/Gaz492/twitch-export-builder)'
        }
    };
    let jsonOptions = {
        url: apiURL + '/curseProjects.json',
        method: 'GET',
        headers: {
            'User-Agent': 'Twitch-Exporter/' + appVersion + ' (+https://github.com/Gaz492/twitch-export-builder)'
        },
        json: true
    };
    if (fs.existsSync(path.join(directories.meta, 'curseProjects.json'))) {
        fs.createReadStream(path.join(directories.meta, 'curseProjects.json')).pipe(crypto.createHash('md5').setEncoding('hex')).on('finish', function () {
            let jsonHash = this.read();
            request(md5Options, function (error, response, body) {
                if (error) console.log(error);

                jsonMD5Hash = body;
                if (jsonMD5Hash.split('\n')[0] !== jsonHash) {
                    console.log('Outdated JSON downloading update', jsonMD5Hash.split('\n')[0], jsonHash);
                    request(jsonOptions)
                        .pipe(fs.createWriteStream(path.join(directories.meta, 'curseProjects.json')))
                        .on('close', function () {
                            console.log('JSON file written!');
                            run();
                        });
                } else {
                    run();
                }
            });
        })
    } else {
        request(jsonOptions)
            .pipe(fs.createWriteStream(path.join(directories.meta, 'curseProjects.json')))
            .on('close', function () {
                console.log('JSON file written!');
                run();
            });
    }
}

function list(val) {
    return val.split(',')
}

function run() {
    curseJson = JSON.parse(fs.readFileSync(path.join(directories.meta, 'curseProjects.json')))['data'];
    program
        .version(appVersion, '-v, --version')
        .usage('[options] <filepath>')
        .option('-d, --dir <path>', 'Path to root folder of Minecraft instance')
        .option('-i, --include <config,maps,options.txt>', "List of files/folders to include in export")
        .option('-n, --packName <packName>', 'Export Name')
        .option('-m, --mcVersion <version>', 'Minecraft Version (e.g 1.12.2)')
        .option('-p, --packVersion <packversion>', 'Pack Version (e.g 1.0.0')
        .option('-a, --packAuthor <author>', 'Author of pack')
        .option('-f, --forgeVersion <version>', 'Forge version (e.g 14.23.2.2624)')
        .option('-c, --config <file>', 'Config file to get pack variables')
        .parse(process.argv);

    if (program.include) {
        list(program.include).forEach(item => {
            copyList.push(item)
        });
    }
    if (!program.packName) {
        questions.push({
            type: 'input',
            name: 'packName',
            message: 'Please enter pack name'
        })
    } else {
        packName = program.packName;
    }
    if (!program.packVersion) {
        questions.push({
            type: 'input',
            name: 'packVersion',
            message: 'Please enter pack version (e.g 1.0.0)',
            default: function () {
                return "1.0.0"
            },
            validate: function (value) {
                let pass = value.match(/(\d+)\.(\d+)\.(\d+)/i);
                if (pass) {
                    return true;
                }
                return "Please enter valid version (e.g. 1.0.0)"
            }
        })
    } else {
        packVersion = program.packVersion;
    }
    if (!program.packAuthor) {
        questions.push({
            type: 'input',
            name: 'packAuthor',
            message: 'Please enter pack author'
        })
    } else {
        packAuthor = program.packAuthor;
    }
    if (!program.mcVersion) {
        questions.push({
            type: 'input',
            name: 'mcVersion',
            message: 'Minecraft version (e.g 1.12.2)',
            default: function () {
                return "1.12.2"
            },
            validate: function (value) {
                let pass = value.match(/(\d+)\.(\d+)\.(\d+)/i);
                if (pass) {
                    return true;
                }
                return "Please enter valid version (e.g. 1.12.2)"
            }
        })
    } else {
        mcVersion = program.mcVersion;
    }
    if (!program.forgeVersion) {
        questions.push({
            type: 'input',
            name: 'forgeVersion',
            message: 'Forge Version (e.g 14.23.2.2624)',
            default: function () {
                return "14.23.2.2624"
            },
            validate: function (value) {
                let pass = value.match(/(\d+)\.(\d+)\.(\d+)\.(\d+)/i);
                if (pass) {
                    return true;
                }
                return "Please enter valid version (e.g. 14.23.2.2624)"
            }
        })
    } else {
        forgeVersion = program.forgeVersion;
    }

    if(program.config && program.dir && program.packName && program.packVersion){
        const exportCfg = JSON.parse(fs.readFileSync(path.join(directories.base, program.config)));
        // packName = exportCfg['packName'];
        packAuthor = exportCfg['packAuthor'];
        // packVersion = exportCfg['packVersion'];
        mcVersion = exportCfg['minecraftVersion'];
        forgeVersion = exportCfg['forgeVersion'];
        copyList = exportCfg['includes'];
        readDirectory(program.dir)
    }
    else if (program.dir && program.packName && program.packAuthor && program.packVersion && program.mcVersion && program.forgeVersion) {
        readDirectory(program.dir)
    }
    else if (program.dir) {
        inquirer.prompt(questions).then(answers => {
            packName = answers.packName;
            packVersion = answers.packVersion;
            packAuthor = answers.packAuthor;
            mcVersion = answers.mcVersion;
            forgeVersion = answers.forgeVersion;
            readDirectory(program.dir)
        });
    } else {
        console.error("No file path specified use -h for help")
    }
}


function readDirectory(dirPath) {
    fs.readdir(dirPath, (err, files) => {
        files.forEach(file => {
            if (file === 'mods') {
                listMods(path.join(dirPath, file))
            }
        });
    });
}

function listMods(modsFolder) {
    let mods = 0;
    fs.readdir(modsFolder, (err, files) => {
        files.forEach(file => {
            if (path.extname(file) === '.jar') {
                mods++;
                modList.push(file);
                if (mods === files.length) {
                    getProjectID()
                }
            }
        });
    });
}

function getProjectID() {
    let temp = 1;
    modList.forEach(mod => {
        Object.entries(curseJson).forEach(project => {
            if((project[1]['fileName'].split('.jar')[0] === mod.split('.jar')[0])){
                if(project[1]['gameVersion'].includes(mcVersion)){
                    foundMods.push(mod);
                    console.log(temp++, mod);
                    projectObj.push({
                        projectID: project[1]['projectID'],
                        fileID: project[1]['projectFileID'],
                        filename: project[1]['fileName'],
                        required: true
                    });
                }
            }
        });
    });
    createExport();
}

function createExport() {
    if (!fs.existsSync(path.join(directories.base, directories.export.root))) {
        fs.mkdirSync(path.join(directories.base, directories.export.root))
    }
    if (!fs.existsSync(path.join(directories.base, directories.export.root, directories.export.overrides))) {
        fs.mkdirSync(path.join(directories.base, directories.export.root, directories.export.overrides))
    }
    if (!fs.existsSync(path.join(directories.base, directories.export.root, directories.export.overrides, directories.export.mods))) {
        fs.mkdirSync(path.join(directories.base, directories.export.root, directories.export.overrides, directories.export.mods))
    }

    let manifest = {
        minecraft: {
            version: mcVersion,
            modLoaders: [
                {
                    id: 'forge-' + forgeVersion,
                    primary: true
                }
            ],
        },
        manifestType: "minecraftModpack",
        manifestVersion: 1,
        name: packName,
        version: packVersion,
        author: packAuthor,
        files: projectObj,
        overrides: "overrides"
    };
    fs.writeFile(path.join(directories.base, directories.export.root, 'manifest.json'), JSON.stringify(manifest), function (err) {
        if (err) {
            return console.log(err);
        }
        console.log("manifest.json created");
    });

    let checkDiff = modList.filter(function (n) {
        return !this.has(n)
    }, new Set(foundMods));

    checkDiff.forEach(mod => {
        fs.copyFile(path.join(program.dir, 'mods', mod), path.join(directories.base, directories.export.root, directories.export.overrides, directories.export.mods, mod), (err) => {
            if (err) return console.log('An error occurred during file copying', err);
        })
    });

    let fileToCopy = new Promise((resolve, reject) => {
        let itemsCopied = 0;
        copyList.forEach((item, index, array) => {
            ncp(path.join(program.dir, item), path.join(directories.base, directories.export.root, directories.export.overrides, item), (err) => {
                if (err) return console.log('An error occurred during copying: ' + item, err);
                console.log('Copied:', item);
                itemsCopied++;
                if (itemsCopied === array.length) resolve();
            });
        })
    });
    fileToCopy.then(() => {
        compress()
    })
}

function compress() {
    console.log('Creating Export...');
    let output = fs.createWriteStream(packName + '-' + packVersion + '.zip');
    let archive = archiver('zip', {zlib: {level: 9}});
    output.on('close', function () {
        console.log(archive.pointer() + ' total bytes');
        console.log('Export: ', packName + '-' + packVersion + '.zip created');
        rimraf(path.join(directories.base, directories.export.root), (err) => {
            if (err) return console.log(err);

            console.log('Cleaning up folders')
        })
    });
    output.on('end', function () {
        console.log('Data has been drained');
    });
    archive.on('warning', function (err) {
        if (err.code === 'ENOENT') {
            console.warn("Warning: ENOENT")
        } else {
            throw err;
        }
    });
    archive.on('error', function (err) {
        throw err;
    });
    archive.pipe(output);
    archive.directory(path.join(directories.base, directories.export.root), false);
    archive.finalize();
}


getCurseMeta();