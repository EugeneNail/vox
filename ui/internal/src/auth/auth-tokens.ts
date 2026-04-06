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

export function storeLoginToken(loginToken: string) {
    localStorage.setItem(loginTokenStorageKey, loginToken);
}

export function getLoginToken() {
    return localStorage.getItem(loginTokenStorageKey);
}

export function getRefreshToken() {
    return localStorage.getItem(refreshTokenStorageKey);
}

export function clearAuthTokens() {
    localStorage.removeItem(loginTokenStorageKey);
    localStorage.removeItem(refreshTokenStorageKey);
}
