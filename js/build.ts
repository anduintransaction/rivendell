import path from "path";
import { ELogLevel, Generator, INpmDtsArgs } from "npm-dts";

const entryPoint = path.join(import.meta.dir, "./src/index.ts");
const distDir = path.join(import.meta.dir, "./dist");
const pkgJson = path.join(import.meta.dir, "./package.json");

(async () => {
  const { dependencies, peerDependencies } = await import(pkgJson);
  await Bun.build({
    entrypoints: [entryPoint],
    sourcemap: "none",
    minify: false,
    target: "bun",
    external: [
      ...Object.keys(dependencies || {}),
      ...Object.keys(peerDependencies || {}),
    ],
    outdir: distDir,
  });

  const generatorOpts: INpmDtsArgs = {
    entry: entryPoint,
    output: path.join(distDir, "index.d.ts"),
    logLevel: ELogLevel.warn,
  };
  await new Generator(generatorOpts, true, true).generate();
})();
