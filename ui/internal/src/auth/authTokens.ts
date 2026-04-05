export const loginTokenStorageKey = "vox.loginToken";
export const refreshTokenStorageKey = "vox.refreshToken";

type AuthTokens = {
  loginToken: string;
  refreshToken: string;
};

export function storeAuthTokens(tokens: AuthTokens) {
  localStorage.setItem(loginTokenStorageKey, tokens.loginToken);
  localStorage.setItem(refreshTokenStorageKey, tokens.refreshToken);
}
