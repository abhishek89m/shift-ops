import { readFile, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const rootDir = path.dirname(fileURLToPath(import.meta.url));
const repoDir = path.resolve(rootDir, '..');

const rootPackagePath = path.join(repoDir, 'package.json');
const webPackagePath = path.join(repoDir, 'apps', 'web', 'package.json');
const apiVersionPath = path.join(repoDir, 'services', 'api', 'internal', 'buildinfo', 'version.go');

const rootPackage = JSON.parse(await readFile(rootPackagePath, 'utf8'));
const webPackage = JSON.parse(await readFile(webPackagePath, 'utf8'));

webPackage.version = rootPackage.version;

await writeFile(webPackagePath, `${JSON.stringify(webPackage, null, 2)}\n`);
await writeFile(
  apiVersionPath,
  `package buildinfo\n\nconst Version = "${rootPackage.version}"\n`
);

process.stdout.write(`Synced version ${rootPackage.version}\n`);

