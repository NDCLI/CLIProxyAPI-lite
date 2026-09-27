const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

const [version, archiveDirectory, outputDirectory] = process.argv.slice(2);
if (!version || !archiveDirectory || !outputDirectory) {
  throw new Error('Usage: node package-release.cjs <version> <archive-directory> <output-directory>');
}

const sourceDirectory = path.resolve(__dirname);
const archivesDirectory = path.resolve(archiveDirectory);
const outputRoot = path.resolve(outputDirectory);
const temporaryRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'lumina-npm-'));
const targets = [
  ['win32', 'x64', 'windows', 'amd64', '.zip'],
  ['win32', 'arm64', 'windows', 'aarch64', '.zip'],
  ['darwin', 'x64', 'darwin', 'amd64', '.tar.gz'],
  ['darwin', 'arm64', 'darwin', 'aarch64', '.tar.gz'],
  ['linux', 'x64', 'linux', 'amd64', '.tar.gz'],
  ['linux', 'arm64', 'linux', 'aarch64', '.tar.gz'],
];

fs.mkdirSync(outputRoot, { recursive: true });
const optionalDependencies = {};

function run(command, args) {
  const result = spawnSync(command, args, { stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`${command} exited with status ${result.status}`);
}

function powershellCommand() {
  return process.platform === 'win32' ? 'powershell.exe' : 'pwsh';
}

function findArchive(platform, architecture, extension) {
  const linuxSuffix = platform === 'linux' ? '_no-plugin' : '';
  const archiveName = `CLIProxyAPI_${version}_${platform}_${architecture}${linuxSuffix}${extension}`;
  const archivePath = path.join(archivesDirectory, archiveName);
  if (!fs.existsSync(archivePath)) throw new Error(`Missing release archive: ${archivePath}`);
  return archivePath;
}

function packageTarball(directory, output) {
  const result = spawnSync('npm', ['pack', directory, '--pack-destination', output], { encoding: 'utf8' });
  if (result.stdout) process.stdout.write(result.stdout);
  if (result.stderr) process.stderr.write(result.stderr);
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`npm pack exited with status ${result.status}`);
}

try {
  for (const [platform, architecture, archivePlatform, archiveArchitecture, extension] of targets) {
    const packageName = `@ndcli/lumina-${platform}-${architecture}`;
    const packageDirectory = path.join(temporaryRoot, `${platform}-${architecture}`);
    const extractionDirectory = path.join(packageDirectory, 'payload');
    const archivePath = findArchive(archivePlatform, archiveArchitecture, extension);
    fs.mkdirSync(extractionDirectory, { recursive: true });

    if (extension === '.zip') {
      const escapedArchive = archivePath.replace(/'/g, "''");
      const escapedDestination = extractionDirectory.replace(/'/g, "''");
      run(powershellCommand(), ['-NoProfile', '-Command', `Expand-Archive -LiteralPath '${escapedArchive}' -DestinationPath '${escapedDestination}'`]);
    } else {
      run('tar', ['-xzf', archivePath, '-C', extractionDirectory]);
    }

    const executable = platform === 'win32' ? 'cli-proxy-api.exe' : 'cli-proxy-api';
    const requiredFiles = [executable, 'config.example.yaml', 'LICENSE'];
    if (platform === 'win32') requiredFiles.push('lumina.exe');
    for (const file of requiredFiles) {
      const source = path.join(extractionDirectory, file);
      if (!fs.existsSync(source)) throw new Error(`Archive ${path.basename(archivePath)} is missing ${file}`);
      fs.copyFileSync(source, path.join(packageDirectory, file));
      if (file === executable) fs.chmodSync(path.join(packageDirectory, file), 0o755);
    }
    fs.rmSync(extractionDirectory, { recursive: true, force: true });

    const manifest = {
      name: packageName,
      version,
      os: [platform],
      cpu: [architecture],
      files: requiredFiles,
      license: 'MIT',
    };
    fs.writeFileSync(path.join(packageDirectory, 'package.json'), `${JSON.stringify(manifest, null, 2)}\n`);
    packageTarball(packageDirectory, outputRoot);
    optionalDependencies[packageName] = version;
  }

  const mainDirectory = path.join(temporaryRoot, 'main');
  fs.cpSync(sourceDirectory, mainDirectory, {
    recursive: true,
    filter: (source) => {
      const relative = path.relative(sourceDirectory, source);
      return relative === '' || (!relative.endsWith('.test.cjs') && relative !== 'package-release.cjs');
    },
  });
  const mainManifestPath = path.join(mainDirectory, 'package.json');
  const mainManifest = JSON.parse(fs.readFileSync(mainManifestPath, 'utf8'));
  fs.copyFileSync(path.join(sourceDirectory, '..', 'LICENSE'), path.join(mainDirectory, 'LICENSE'));
  mainManifest.version = version;
  mainManifest.optionalDependencies = optionalDependencies;
  fs.writeFileSync(mainManifestPath, `${JSON.stringify(mainManifest, null, 2)}\n`);
  packageTarball(mainDirectory, outputRoot);
} finally {
  fs.rmSync(temporaryRoot, { recursive: true, force: true });
}
