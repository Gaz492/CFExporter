const program = require('commander');
const inquirer = require('inquirer');
const fs = require('fs');
const path = require('path');
const request = require('request');
const crypto = require('crypto');
const AdmZip = require('adm-zip');
const rimraf = require('rimraf');
const ncp = require('ncp');
const archiver = require('archiver');
const gunzip = require('gunzip-file');

const questions = [];
const directories = {
    export: {
        root: 'temp',
        overrides: 'overrides',
        mods: 'mods'
    },
    meta: 'meta'
};

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


if (fs.existsSync(path.join(__dirname, directories.export.root))) {
    rimraf(path.join(__dirname, directories.export.root), (err) => {
        if (err) return console.log(err);
    })
}

getCurseMeta();

function getCurseMeta() {
    // let options = {
    //     url: 'https://cursemeta.dries007.net/raw_mods.json',
    //     method: 'GET',
    //     json: true
    // };
    // request(options, function (error, response, body) {
    //     if (error) console.log(error);
    //
    //     curseJson = body['Data'];
    //     run()
    // });
    if (!fs.existsSync(path.join(__dirname, directories.meta))) {
        fs.mkdirSync(path.join(__dirname, directories.meta))
    }

    request('https://cursemeta.dries007.net/raw_mods.json.gz')
        .pipe(fs.createWriteStream(path.join(__dirname, directories.meta, 'raw_mods.json.gz')))
        .on('close', function () {
            console.log('File written!');
            gunzip(path.join(__dirname, directories.meta, 'raw_mods.json.gz'), path.join(__dirname, directories.meta, 'raw_mods.json'), () => {
                console.log('JSON Archive extracted');
                fs.readFile(path.join(__dirname, directories.meta, 'raw_mods.json'), 'utf8', (err, data) => {
                    if (err) return console.log(err);
                    curseJson = JSON.parse(data)['Data'];
                    run();
                });
            });
        });
}

function list(val) {
    return val.split(',')
}

function run() {
    program
        .version('1.0.0', '-v, --version')
        .usage('[options] <filepath>')
        .option('-d, --dir <path>', 'Path to root folder of Minecraft instance')
        .option('-i, --include <config,maps,options.txt>', "List of files/folders to include in export")
        .option('-n, --packName <packName>', 'Export Name')
        .option('-mv, --mcVersion <version>', 'Minecraft Version (e.g 1.12.2)')
        .option('-pv, --packVersion <packversion>', 'Pack Version (e.g 1.0.0')
        .option('-a, --author <author>', 'Author of pack')
        .option('-f, --forgeVersion <version>', 'Forge version (e.g 14.23.2.2624)')
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
    }
    if (!program.author) {
        questions.push({
            type: 'input',
            name: 'packAuthor',
            message: 'Please enter pack author'
        })
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
    }


    if (program.dir) {
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
    modList.forEach(mod => {
        Object.entries(curseJson).forEach(project => {
            project[1]['GameVersionLatestFiles'].find(files => {
                if (files['ProjectFileName'] === mod) {
                    if (files['GameVesion'] >= mcVersion.split('.').slice(0, 2).join('.')) {
                        projectObj.push({
                            projectID: project[1]['Id'],
                            fileID: files['ProjectFileID'],
                            filename: files['ProjectFileName'],
                            required: true
                        });
                        foundMods.push(mod);
                        return true;
                    }
                }
            })
        });
    });
    createExport();
}

function createExport() {
    if (!fs.existsSync(path.join(__dirname, directories.export.root))) {
        fs.mkdirSync(path.join(__dirname, directories.export.root))
    }
    if (!fs.existsSync(path.join(__dirname, directories.export.root, directories.export.overrides))) {
        fs.mkdirSync(path.join(__dirname, directories.export.root, directories.export.overrides))
    }
    if (!fs.existsSync(path.join(__dirname, directories.export.root, directories.export.overrides, directories.export.mods))) {
        fs.mkdirSync(path.join(__dirname, directories.export.root, directories.export.overrides, directories.export.mods))
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
    fs.writeFile(path.join(__dirname, directories.export, 'manifest.json'), JSON.stringify(manifest), function (err) {
        if (err) {
            return console.log(err);
        }
        console.log("manifest.json created");
    });

    let checkDiff = modList.filter(function (n) {
        return !this.has(n)
    }, new Set(foundMods));

    checkDiff.forEach(mod => {
        fs.copyFile(path.join(program.dir, 'mods', mod), path.join(__dirname, directories.export.root, directories.export.overrides, directories.export.mods) + mod, (err) => {
            if (err) return console.log('An error occurred during file copying', err);
        })
    });

    let fileToCopy = new Promise((resolve, reject) => {
        let itemsCopied = 0;
        copyList.forEach((item, index, array) => {
            ncp(path.join(program.dir, item), path.join(__dirname, directories.export.root, directories.export.overrides) + item, (err) => {
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
        // rimraf(path.join(__dirname, directories.export.root), (err) => {
        //     if (err) return console.log(err);
        //     console.log('Cleaning up foldesrs')
        // })
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
    archive.directory(path.join(__dirname, directories.export), false);
    archive.finalize();
}
