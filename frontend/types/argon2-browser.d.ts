declare module "argon2-browser" {
  export enum ArgonType {
    Argon2d = 0,
    Argon2i = 1,
    Argon2id = 2,
  }

  export interface Argon2HashOptions {
    pass: string | Uint8Array;
    salt: string | Uint8Array;

    type?: ArgonType;

    time?: number;
    mem?: number;
    parallelism?: number;
    hashLen?: number;

    distPath?: string;
  }

  export interface Argon2HashResult {
    encoded: string;
    hash: Uint8Array;
  }

  export function hash(
    options: Argon2HashOptions
  ): Promise<Argon2HashResult>;
}