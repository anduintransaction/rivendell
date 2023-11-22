import { K8sObject, Wait } from "./common";
import { Context } from "./context";

export interface SourceGenerator {
  (ctx: Context): Promise<K8sObject[]>;
}

export interface ModuleOpts {
  deps?: string[];
  generator?: SourceGenerator;
  waits?: Wait[];
}

export class Module {
  name: string;
  deps: string[];
  generator: SourceGenerator;
  waits: Wait[];

  constructor(
    name: string,
    opts?: ModuleOpts,
  ) {
    this.name = name;
    this.deps = [...(opts?.deps || [])].sort();
    this.generator = opts?.generator || ((_) => Promise.resolve([]));
    this.waits = opts?.waits || [];
  }
}
