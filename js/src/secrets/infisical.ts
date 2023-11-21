import InfisicalClient from "infisical-node";
import { SecretProvider } from "../context";

export class InfisicalSecretProvider implements SecretProvider {
  client: InfisicalClient;

  constructor(siteURL: string, token: string, debug: boolean = false) {
    this.client = new InfisicalClient({
      token: token,
      siteURL: siteURL,
      debug: debug,
    });
  }

  async get(env: string, name: string) {
    const parts = name.split("/");
    const [secretName, ...paths] = parts.reverse();
    const secret = await this.client.getSecret(secretName, {
      environment: env,
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
      environment: env,
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
