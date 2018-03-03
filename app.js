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

const questions = [
    {
        type: 'input',
        name: 'packName',
        message: 'Please enter pack name'
    },
    {
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
    },
    {
        type: 'input',
        name: 'packAuthor',
        message: 'Please enter pack author'
    },
    {
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
    },
    {
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
    }
];

let packName;
let packVersion;
let packAuthor;
let mcVersion;
let forgeVersion;
let projectObj = [];
let modList = [];
let foundMods = [];
let copyList = ['config'];
let curseJson;

if (!fs.existsSync('meta')) {
    fs.mkdirSync('meta')
}

if (fs.existsSync('export')) {
    rimraf('./export', (err) => {
        if (err) return console.log(err);
    })
}

checkMeta();

function checkMeta() {
    if (fs.existsSync('./meta/curse.json')) {
        fs.createReadStream('./meta/curse.json').pipe(crypto.createHash('md5').setEncoding('hex')).on('finish', function () {
            let jsonHash = this.read();
            request('https://fdn.redstone.tech/theoneclient/hl3/onemeta/curse.json.md5')
                .pipe(fs.createWriteStream('./meta/curse.json.md5'))
                .on('close', function () {
                    fs.readFile('./meta/curse.json.md5', 'utf8', function (err, data) {
                        if (err) {
                            console.log(err);
                        }
                        if (data.split('\n')[0] !== jsonHash) {
                            request('https://fdn.redstone.tech/theoneclient/hl3/onemeta/curse.zip')
                                .pipe(fs.createWriteStream('./meta/curse.zip'))
                                .on('close', function () {
                                    console.log('File written!');
                                    let zip = new AdmZip("./meta/curse.zip");
                                    zip.extractAllTo("./meta/", true);
                                    run();
                                });
                        } else {
                            run();
                        }
                    });
                });
        })
    } else {
        request('https://fdn.redstone.tech/theoneclient/hl3/onemeta/curse.zip')
            .pipe(fs.createWriteStream('./meta/curse.zip'))
            .on('close', function () {
                console.log('File written!');
                let zip = new AdmZip("./meta/curse.zip");
                zip.extractAllTo("./meta/", true);
                run();
            });
    }
}

function list(val) {
    return val.split(',')
}

function run() {
    curseJson = JSON.parse(fs.readFileSync('./meta/curse.json', 'utf8'));
    program
        .version('1.0.0', '-v, --version')
        .usage('[options] <filepath>')
        .option('-d, --dir <path>', 'Path to root folder of Minecraft instance')
        .option('-i, --include <config,maps,options.txt>', "List of files/folders to include in export")
        .option('-n, --name <packName>', 'Export Name')
        .option('-mv, --mcVersion <version>', 'Minecraft Version (e.g 1.12.2)')
        .option('-pv, --packVersion <packversion>', 'Pack Version (e.g 1.0.0')
        .option('-a, --author <author>', 'Author of pack')
        .option('-f, --forgeVersion <version>', 'Forge version (e.g 14.23.2.2624)')
        .parse(process.argv);

    if(program.include){
        list(program.include).forEach(item => {
            copyList.push(item)
        });
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
        Object.entries(curseJson['files']).forEach(project => {
            if (project[1]['filename'] === mod) {
                if (project[1]['minecraft'].find(mcVer => {
                        if (mcVer === mcVersion) {
                            projectObj.push({
                                projectID: project[1]['project'],
                                fileID: project[1]['id'],
                                filename: project[1]['filename'],
                                required: true
                            });
                            foundMods.push(mod);
                        }
                    })) {
                }
            }
        })
    });

    createExport();
}

function createExport() {
    if (!fs.existsSync('export')) {
        fs.mkdirSync('export')
    }
    if (!fs.existsSync('export/overrides')) {
        fs.mkdirSync('export/overrides')
    }
    if (!fs.existsSync('export/overrides/mods')) {
        fs.mkdirSync('export/overrides/mods')
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
    fs.writeFile("./export/manifest.json", JSON.stringify(manifest), function (err) {
        if (err) {
            return console.log(err);
        }
        console.log("manifest.json created");
    });
    let checkDiff = modList.filter(function (n) {
        return !this.has(n)
    }, new Set(foundMods));

    checkDiff.forEach(mod => {
        fs.copyFile(path.join(program.dir, 'mods', mod), './export/overrides/mods/' + mod, (err) => {
            if (err) return console.log('An error occurred during file copying', err);
        })
    });

    let fileToCopy = new Promise((resolve, reject) => {
        let itemsCopied = 0;
        copyList.forEach((item, index, array) => {
            ncp(path.join(program.dir, item), './export/overrides/' + item, (err) => {
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
    console.log('Compressing Output');
    let output = fs.createWriteStream(packName + '-' + packVersion + '.zip');
    let archive = archiver('zip', {});

    output.on('close', function () {
        console.log(archive.pointer() + ' total bytes');
        console.log('archiver has been finalized and the output file descriptor has closed.');
        console.log(packName + '-' + packVersion + '.zip created')
    });

    output.on('end', function () {
        console.log('Data has been drained');
    });
    archive.on('warning', function (err) {
        if (err.code === 'ENOENT') {
            // log warning
        } else {
            // throw error
            throw err;
        }
    });
    archive.on('error', function (err) {
        throw err;
    });

    archive.pipe(output);

    archive.directory('export', false);
    archive.finalize();
}
