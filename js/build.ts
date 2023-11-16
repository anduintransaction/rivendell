import path from "path";
import { Generator } from "npm-dts";

const entryPoint = path.resolve(import.meta.dir, "./src/index.ts");
const distDir = path.resolve(import.meta.dir, "./dist");
const packageJsonFile = path.resolve(import.meta.dir, "./package.json");

const pkgJson = JSON.parse(await Bun.file(packageJsonFile).text());
const { dependencies, peerDependencies } = pkgJson;

const isCI = (process.env.CI || "false").toLowerCase() === "true";

Bun.build({
  entrypoints: [entryPoint],
  sourcemap: "external",
  minify: isCI,
  target: "bun",
  external: [...Object.keys(dependencies), ...Object.keys(peerDependencies)],
  outdir: distDir,
});

new Generator({
  entry: entryPoint,
  output: path.join(distDir, "index.d.ts"),
}).generate();
