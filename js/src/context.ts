export class Context {
  configs: any;

  static Empty = new Context({});

  constructor(configs: any) {
    this.configs = configs;
  }
}
