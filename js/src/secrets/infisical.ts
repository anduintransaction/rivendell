import InfisicalClient from "infisical-node";
import { SecretProvider } from "../context";

export class InfisicalSecretProvider implements SecretProvider {
  client: InfisicalClient;
  forceEnv: string | undefined;

  constructor(
    siteURL: string,
    token: string,
    debug: boolean = false,
    forceEnv?: string,
  ) {
    this.forceEnv = forceEnv;
    this.client = new InfisicalClient({
      token: token,
      siteURL: siteURL,
      debug: debug,
    });
  }

  getTargetEnv(env: string): string {
    return this.forceEnv || env;
  }

  async get(env: string, name: string) {
    const parts = name.split("/");
    const [secretName, ...paths] = parts.reverse();
    const secret = await this.client.getSecret(secretName, {
      environment: this.getTargetEnv(env),
      path: `/${paths.reverse().join("/")}`,
      type: "shared",
    });
    return {
      name: name,
      value: secret.secretValue,
    };
  }

  async getPrefix(env: string, prefix: string) {
    const secrets = await this.client.getAllSecrets({
      environment: this.getTargetEnv(env),
      path: `/${prefix}`,
      attachToProcessEnv: false,
      includeImports: false,
    });
    return secrets.map((s) => ({
      name: s.secretName,
      value: s.secretValue,
    }));
  }
}
