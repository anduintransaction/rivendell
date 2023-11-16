import { build } from "esbuild";
import { fileURLToPath } from "url";
import path from "path";
import fs from "fs";
import ndts from "npm-dts";
const { Generator } = ndts;

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const entryPoint = path.resolve(__dirname, "./src/index.ts");
const distDir = path.resolve(__dirname, "./dist");

const packageJsonPath = path.resolve(__dirname, "package.json");
const packageJson = JSON.parse(fs.readFileSync(packageJsonPath));
const { dependencies, peerDependencies } = packageJson;

const sharedConfig = {
  entryPoints: [entryPoint],
  sourcemap: true,
  bundle: true,
  platform: "node",
  external: [...Object.keys(dependencies), ...Object.keys(peerDependencies)],
};

build({
  ...sharedConfig,
  outfile: path.join(distDir, "index.js"),
});

build({
  ...sharedConfig,
  minify: true,
  outfile: path.join(distDir, "index.min.js"),
});

new Generator({
  entry: entryPoint,
  output: path.join(distDir, "index.d.ts"),
}).generate();
