export class Context {
  configs: any;
  secrets: any;

  constructor(configs: any = {}, secrets: any = {}) {
    this.configs = configs;
    this.secrets = secrets;
  }
}
