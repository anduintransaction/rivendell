export interface SecretValue {
  name: string;
  value: string | undefined;
}

export interface SecretProvider {
  get(env: string, name: string): Promise<SecretValue>;
  getPrefix(env: string, prefix: string): Promise<SecretValue[]>;
}

export class NoopSecretProvider implements SecretProvider {
  get(_: string, name: string): Promise<SecretValue> {
    const paths = name.split("/");
    const secretName = paths[paths.length - 1];
    return Promise.resolve({
      name: secretName,
      value: undefined,
    });
  }

  getPrefix(_: string, __: string): Promise<SecretValue[]> {
    return Promise.resolve([]);
  }
}

export class Context {
  env: string;
  configs: any;
  secretProvider: SecretProvider;

  constructor(
    env: string,
    configs: any = {},
    secretProvider: SecretProvider = new NoopSecretProvider(),
  ) {
    this.env = env;
    this.configs = configs;
    this.secretProvider = secretProvider;
  }

  getSecret(name: string): Promise<SecretValue> {
    return this.secretProvider.get(this.env, name);
  }

  getSecretPrefix(prefix: string): Promise<SecretValue[]> {
    return this.secretProvider.getPrefix(this.env, prefix);
  }
}

export const Finder = {
  optinal<T>(
    obj: any,
    paths: string[],
    defaultValue?: T,
  ): T | undefined {
    let current = obj;
    for (const p of paths) {
      const val = current[p];
      if (val === undefined || val === null) {
        return (defaultValue === undefined) ? undefined : defaultValue;
      }
      current = val;
    }
    return current as T;
  },

  required<T>(
    obj: any,
    paths: string[],
    defaultValue?: T,
  ): T {
    const res = this.optinal(obj, paths, defaultValue);
    if (res === undefined) {
      throw new Error(`cannot find value for paths ${paths.join(".")}`);
    }
    return res!;
  },

  optionalString(obj: any, paths: string[], defaultValue?: string) {
    return this.optinal<string>(obj, paths, defaultValue);
  },
  requiredString(obj: any, paths: string[], defaultValue?: string) {
    return this.required<string>(obj, paths, defaultValue);
  },
  optionalNumber(obj: any, paths: string[], defaultValue?: number) {
    return this.optinal<number>(obj, paths, defaultValue);
  },
  requiredNumber(obj: any, paths: string[], defaultValue?: number) {
    return this.required<number>(obj, paths, defaultValue);
  },
  optionalBool(obj: any, paths: string[], defaultValue?: boolean) {
    return this.optinal<boolean>(obj, paths, defaultValue);
  },
  requiredBool(obj: any, paths: string[], defaultValue?: boolean) {
    return this.required<boolean>(obj, paths, defaultValue);
  },

  getAsArray<K>(obj: any, paths: string[], defaultValue?: K[]) {
    const value = this.required(obj, paths, defaultValue);
    if (Array.isArray(value)) return value;
    return [value];
  },
};
