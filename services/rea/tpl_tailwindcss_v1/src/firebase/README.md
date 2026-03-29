# Firebase replaced by DEX

Firebase authentication has been removed from this project.

All authentication logic is now handled by DEX (OIDC) via PKCE Authorization Code Flow.

See `src/auth/authProvider.ts` for the single file that contains all provider-specific code.
