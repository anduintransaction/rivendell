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

  clone(newOpts: ModuleOpts) {
    return new Module(this.name, {
      deps: newOpts.deps || this.deps,
      generator: newOpts.generator || this.generator,
      waits: newOpts.waits || this.waits,
    });
  }

  cloneGeneratorOnly() {
    return this.clone({
      deps: [],
      waits: [],
    });
  }
}
