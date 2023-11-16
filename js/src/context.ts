export class Context {
  configs: Record<string, any>;

  static Empty = new Context({});

  constructor(configs: Record<string, any>) {
    this.configs = configs;
  }

  merge(that: Context): Context {
    return new Context(Object.assign({}, this.configs, that.configs));
  }
}
