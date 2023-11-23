import fs from "fs/promises";
import path from "path";
import url from "url";
import esbuild from "esbuild";
import npmdts from "npm-dts";
const { ELogLevel, Generator } = npmdts;

const __dirname = path.dirname(url.fileURLToPath(import.meta.url));
const entryPoint = path.join(__dirname, "./src/index.ts");
const distDir = path.join(__dirname, "./dist");
const pkgJson = path.join(__dirname, "./package.json");

(async () => {
  const { dependencies, peerDependencies } = JSON.parse(
    await fs.readFile(pkgJson),
  );
  esbuild.build({
    entryPoints: [entryPoint],
    outfile: path.join(distDir, "index.js"),
    bundle: true,
    minify: false,
    platform: "node",
    format: "cjs",
    external: [
      ...Object.keys(dependencies || {}),
      ...Object.keys(peerDependencies || {}),
    ],
  });

  const generatorOpts = {
    entry: entryPoint,
    output: path.join(distDir, "index.d.ts"),
    logLevel: ELogLevel.warn,
  };
  await new Generator(generatorOpts, true, true).generate();
})();
