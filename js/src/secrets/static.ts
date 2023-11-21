import { Finder, SecretProvider, SecretValue } from "../context";

export class StaticSecretProvider implements SecretProvider {
  secrets: any;

  constructor(secrets: any) {
    this.secrets = secrets;
  }

  get(_: string, name: string): Promise<SecretValue> {
    const paths = name.split("/");
    const secretName = paths[paths.length - 1];
    const value = Finder.optinal<string>(this.secrets, paths);
    if (value !== undefined && typeof value !== "string") {
      throw new Error(`invalid secret value type of secret ${name}`);
    }
    return Promise.resolve({ name: secretName, value });
  }

  getPrefix(_: string, prefix: string): Promise<SecretValue[]> {
    const paths = prefix.split("/");
    const value = Finder.optinal<any>(this.secrets, paths);
    if (value === undefined) return Promise.resolve([]);
    const res: SecretValue[] = [];
    for (const [key, val] of Object.entries(value)) {
      res.push({
        name: key,
        value: val as string,
      });
    }
    return Promise.resolve(res);
  }
}
