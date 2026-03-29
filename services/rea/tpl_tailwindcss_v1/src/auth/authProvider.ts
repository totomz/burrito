// src/auth/authProvider.ts
// THE single auth file. All DEX/OIDC-specific logic lives here.
// To swap to another provider, replace only this file.

const DEX_ISSUER = import.meta.env.VITE_DEX_ISSUER as string;
const DEX_CLIENT_ID = import.meta.env.VITE_DEX_CLIENT_ID as string;
const DEX_REDIRECT_URI = import.meta.env.VITE_DEX_REDIRECT_URI as string;
const DEX_SCOPE = (import.meta.env.VITE_DEX_SCOPE as string) || "openid email profile";
const BACKEND_URL = import.meta.env.VITE_BACKEND_URL as string;

const AUTHORIZE_ENDPOINT = `${DEX_ISSUER}/auth`;
const TOKEN_ENDPOINT = `${DEX_ISSUER}/token`;

// ---- Types ----------------------------------------------------------------

export type AuthUser = {
  email: string;
  name?: string;
  sub: string;
  idToken: string;
  accessToken: string;
};

export type AuthStateListener = (user: AuthUser | null) => void;

// ---- Internal state -------------------------------------------------------

const listeners: Set<AuthStateListener> = new Set();

function notifyListeners(user: AuthUser | null): void {
  listeners.forEach((fn) => fn(user));
}

// ---- PKCE helpers ---------------------------------------------------------

function generateRandomString(length: number): string {
  const array = new Uint8Array(length);
  crypto.getRandomValues(array);
  return Array.from(array, (byte) =>
    byte.toString(36).padStart(2, "0")
  )
    .join("")
    .slice(0, length);
}

async function sha256(plain: string): Promise<ArrayBuffer> {
  const encoder = new TextEncoder();
  const data = encoder.encode(plain);
  return crypto.subtle.digest("SHA-256", data);
}

function base64urlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  bytes.forEach((b) => (binary += String.fromCharCode(b)));
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}

// ---- Token storage --------------------------------------------------------

const STORAGE_KEYS = {
  ACCESS_TOKEN: "access_token",
  ID_TOKEN: "id_token",
  REFRESH_TOKEN: "refresh_token",
  EXPIRES_AT: "expires_at",
} as const;

const SESSION_KEYS = {
  CODE_VERIFIER: "pkce_code_verifier",
  STATE: "pkce_state",
} as const;

function storeTokens(tokenResponse: {
  access_token: string;
  id_token: string;
  refresh_token?: string;
  expires_in: number;
}): void {
  const expiresAt = Date.now() + tokenResponse.expires_in * 1000;
  localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, tokenResponse.access_token);
  localStorage.setItem(STORAGE_KEYS.ID_TOKEN, tokenResponse.id_token);
  if (tokenResponse.refresh_token) {
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, tokenResponse.refresh_token);
  }
  localStorage.setItem(STORAGE_KEYS.EXPIRES_AT, String(expiresAt));
}

function clearTokens(): void {
  localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN);
  localStorage.removeItem(STORAGE_KEYS.ID_TOKEN);
  localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
  localStorage.removeItem(STORAGE_KEYS.EXPIRES_AT);
}

function decodeJwtPayload(token: string): Record<string, any> {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) throw new Error("Invalid JWT format");
    const payload = parts[1];
    // Pad base64url to base64
    const padded = payload.replace(/-/g, "+").replace(/_/g, "/");
    const jsonStr = atob(padded);
    return JSON.parse(jsonStr);
  } catch {
    return {};
  }
}

// ---- Public API -----------------------------------------------------------

/**
 * Initiates the PKCE Authorization Code Flow by redirecting to DEX.
 */
export async function login(): Promise<void> {
  const codeVerifier = generateRandomString(64);
  const state = generateRandomString(32);

  sessionStorage.setItem(SESSION_KEYS.CODE_VERIFIER, codeVerifier);
  sessionStorage.setItem(SESSION_KEYS.STATE, state);

  const codeChallenge = base64urlEncode(await sha256(codeVerifier));

  const params = new URLSearchParams({
    response_type: "code",
    client_id: DEX_CLIENT_ID,
    redirect_uri: DEX_REDIRECT_URI,
    scope: DEX_SCOPE,
    state,
    code_challenge: codeChallenge,
    code_challenge_method: "S256",
  });

  window.location.href = `${AUTHORIZE_ENDPOINT}?${params.toString()}`;
}

/**
 * Handles the OIDC redirect callback.
 * Exchanges the authorization code for tokens and stores them.
 * Throws on error.
 */
export async function handleCallback(): Promise<void> {
  const params = new URLSearchParams(window.location.search);
  const code = params.get("code");
  const returnedState = params.get("state");
  const error = params.get("error");

  if (error) {
    throw new Error(`Authorization error: ${error} - ${params.get("error_description") ?? ""}`);
  }

  if (!code) {
    throw new Error("Missing authorization code in callback");
  }

  const storedState = sessionStorage.getItem(SESSION_KEYS.STATE);
  if (returnedState !== storedState) {
    throw new Error("State mismatch – possible CSRF attack");
  }

  const codeVerifier = sessionStorage.getItem(SESSION_KEYS.CODE_VERIFIER);
  if (!codeVerifier) {
    throw new Error("Missing PKCE code verifier");
  }

  const body = new URLSearchParams({
    grant_type: "authorization_code",
    code,
    redirect_uri: DEX_REDIRECT_URI,
    client_id: DEX_CLIENT_ID,
    code_verifier: codeVerifier,
  });

  const response = await fetch(TOKEN_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Token exchange failed: ${response.status} ${text}`);
  }

  const tokenResponse = await response.json();
  storeTokens(tokenResponse);

  // Clean up session storage
  sessionStorage.removeItem(SESSION_KEYS.CODE_VERIFIER);
  sessionStorage.removeItem(SESSION_KEYS.STATE);

  // Notify subscribers
  notifyListeners(getUser());
}

/**
 * Clears tokens and redirects to the login page.
 * Tokens are removed from localStorage and PKCE state from sessionStorage.
 */
export function logout(): void {
  // 1. Clear stored tokens
  clearTokens();

  // 2. Clear any leftover PKCE session state
  sessionStorage.removeItem(SESSION_KEYS.CODE_VERIFIER);
  sessionStorage.removeItem(SESSION_KEYS.STATE);

  // 3. Notify subscribers synchronously before navigating away
  notifyListeners(null);

  // 4. Redirect to login
  window.location.href = window.location.origin + "/login";
}

/**
 * Returns the current user from stored tokens, or null if not authenticated.
 */
export function getUser(): AuthUser | null {
  if (!isAuthenticated()) return null;

  const accessToken = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)!;
  const idToken = localStorage.getItem(STORAGE_KEYS.ID_TOKEN)!;
  const claims = decodeJwtPayload(idToken);

  return {
    email: claims.email ?? "",
    name: claims.name,
    sub: claims.sub ?? "",
    idToken,
    accessToken,
  };
}

/**
 * Returns true if tokens are present and not expired.
 */
export function isAuthenticated(): boolean {
  const accessToken = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN);
  const expiresAt = localStorage.getItem(STORAGE_KEYS.EXPIRES_AT);

  if (!accessToken || !expiresAt) return false;

  // Consider expired if within 30 seconds of expiry
  return Date.now() < Number(expiresAt) - 30_000;
}

/**
 * Subscribe to auth state changes.
 * Returns an unsubscribe function.
 */
export function onAuthStateChange(listener: AuthStateListener): () => void {
  listeners.add(listener);
  // Emit the current state immediately so the listener is in sync
  listener(getUser());
  return () => {
    listeners.delete(listener);
  };
}

/**
 * Calls VITE_BACKEND_URL/auth/signup to create a new account.
 */
export async function signup(
  email: string,
  password: string,
  firstName: string,
  lastName: string
): Promise<void> {
  const response = await fetch(`${BACKEND_URL}/auth/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, firstName, lastName }),
  });

  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.message ?? `Signup failed: ${response.status}`);
  }
}

/**
 * Requests a password reset email for the given address.
 * Backend expects the email as plain text at POST /password/request-reset.
 */
export async function requestPasswordReset(email: string): Promise<void> {
  const response = await fetch(`${BACKEND_URL}/password/request-reset`, {
    method: "POST",
    headers: { "Content-Type": "text/plain" },
    body: email,
  });

  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.message ?? `Password reset request failed: ${response.status}`);
  }
}

/**
 * Confirms a password reset using the token from the reset email.
 * Backend expects the new password as plain text at POST /password/reset?token=<token>.
 */
export async function resetPassword(token: string, newPassword: string): Promise<void> {
  const response = await fetch(`${BACKEND_URL}/password/reset?token=${encodeURIComponent(token)}`, {
    method: "POST",
    headers: { "Content-Type": "text/plain" },
    body: newPassword,
  });

  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.message ?? `Password reset failed: ${response.status}`);
  }
}
