"use client";

import { argon2id } from "@noble/hashes/argon2.js";

import {
  db,
  getOfflineCredential,
  saveOfflineCredential,
} from "@/lib/offline/db";

import type { AuthUser } from "@/types/auth";

const ARGON2_MEMORY = 64 * 1024;
const ARGON2_ITERATIONS = 3;
const ARGON2_PARALLELISM = 1;
const ARGON2_HASH_LENGTH = 32;

interface OfflineCredentialInput {
  user: AuthUser;
  tenantId: string;
  branchId: string;
  branchName: string;
  password: string;
  credentialVersion: number;
}

interface OfflineLoginResult {
  user: AuthUser;
  branchId: string;
  branchName: string;
}

function generateSalt(): Uint8Array {
  const salt = new Uint8Array(16);

  crypto.getRandomValues(salt);

  return salt;
}

function passwordToBytes(
  password: string
): Uint8Array {
  return new TextEncoder().encode(password);
}

function bytesToHex(
  bytes: Uint8Array
): string {
  return Array.from(bytes)
    .map((byte) =>
      byte.toString(16).padStart(2, "0")
    )
    .join("");
}

function hexToBytes(
  hex: string
): Uint8Array {
  if (hex.length % 2 !== 0) {
    throw new Error(
      "Invalid password verifier"
    );
  }

  const bytes = new Uint8Array(
    hex.length / 2
  );

  for (
    let i = 0;
    i < bytes.length;
    i += 1
  ) {
    bytes[i] = Number.parseInt(
      hex.slice(i * 2, i * 2 + 2),
      16
    );
  }

  return bytes;
}

async function derivePasswordVerifier(
  password: string,
  salt: Uint8Array
): Promise<string> {
  const hash = argon2id(
    passwordToBytes(password),
    salt,
    {
      t: ARGON2_ITERATIONS,
      m: ARGON2_MEMORY,
      p: ARGON2_PARALLELISM,
      dkLen: ARGON2_HASH_LENGTH,
    }
  );

  return bytesToHex(hash);
}

function constantTimeEqual(
  left: Uint8Array,
  right: Uint8Array
): boolean {
  if (left.length !== right.length) {
    return false;
  }

  let difference = 0;

  for (
    let i = 0;
    i < left.length;
    i += 1
  ) {
    difference |= left[i] ^ right[i];
  }

  return difference === 0;
}

export async function createOfflineCredential(
  input: OfflineCredentialInput
): Promise<void> {
  if (!input.user.id) {
    throw new Error("User ID is required");
  }

  if (!input.tenantId) {
    throw new Error("Tenant ID is required");
  }

  if (!input.branchId) {
    throw new Error("Branch ID is required");
  }

  if (!input.branchName) {
    throw new Error("Branch name is required");
  }

  if (!input.password) {
    throw new Error("Password is required");
  }

  const salt = generateSalt();

  const passwordVerifier =
    await derivePasswordVerifier(
      input.password,
      salt
    );

  const now =
    new Date().toISOString();

  /*
   * saveOfflineCredential() already manages
   * its own Dexie transaction.
   *
   * Do NOT wrap it in another transaction here.
   */
  await saveOfflineCredential({
    id: crypto.randomUUID(),

    user_id: input.user.id,

    tenant_id: input.tenantId,

    email: input.user.email,

    role: "attendant",

    branch_id: input.branchId,

    branch_name: input.branchName,

    password_verifier:
      passwordVerifier,

    password_salt:
      bytesToHex(salt),

    account_status: "active",

    credential_version:
      input.credentialVersion,

    last_online_verified_at: now,

    created_at: now,

    updated_at: now,
  });
}

export async function loginOffline(
  email: string,
  password: string
): Promise<OfflineLoginResult> {
  const credential =
    await getOfflineCredential();

  if (!credential) {
    throw new Error(
      "Offline login is not available on this workstation."
    );
  }

  if (credential.role !== "attendant") {
    throw new Error(
      "Only attendants can use offline login."
    );
  }

  if (
    credential.account_status !==
    "active"
  ) {
    throw new Error(
      "This attendant account is no longer active."
    );
  }

  if (
    credential.email.toLowerCase() !==
    email.trim().toLowerCase()
  ) {
    throw new Error(
      "Invalid email or password"
    );
  }

  const salt = hexToBytes(
    credential.password_salt
  );

  const expectedVerifier =
    hexToBytes(
      credential.password_verifier
    );

  const actualVerifier =
    hexToBytes(
      await derivePasswordVerifier(
        password,
        salt
      )
    );

  if (
    !constantTimeEqual(
      expectedVerifier,
      actualVerifier
    )
  ) {
    throw new Error(
      "Invalid email or password"
    );
  }

  const user: AuthUser = {
    id: credential.user_id,

    tenant_id:
      credential.tenant_id,

    full_name:
      credential.email,

    email:
      credential.email,

    phone: "",

    role: "attendant",

    status: "active",

    last_login_at:
      credential.last_online_verified_at,

    created_at:
      credential.created_at,

    updated_at:
      credential.updated_at,
  };

  return {
    user,

    branchId:
      credential.branch_id,

    branchName:
      credential.branch_name,
  };
}

export async function getOfflineAttendant(): Promise<{
  userId: string;
  tenantId: string;
  branchId: string;
  branchName: string;
  email: string;
  status: string;
} | null> {
  const credential =
    await getOfflineCredential();

  if (!credential) {
    return null;
  }

  return {
    userId:
      credential.user_id,

    tenantId:
      credential.tenant_id,

    branchId:
      credential.branch_id,

    branchName:
      credential.branch_name,

    email:
      credential.email,

    status:
      credential.account_status,
  };
}

export async function removeOfflineAttendant(): Promise<void> {
  await db.offlineCredentials.clear();
}

export async function isOfflineCredentialActive(): Promise<boolean> {
  const credential =
    await getOfflineCredential();

  return (
    credential !== undefined &&
    credential.role === "attendant" &&
    credential.account_status === "active"
  );
}