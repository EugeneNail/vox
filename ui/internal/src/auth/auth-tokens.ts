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

export function getAuthenticatedUserUuid() {
    const loginToken = getLoginToken();
    if (!loginToken) {
        return null;
    }

    const payload = loginToken.split(".")[1];
    if (!payload) {
        return null;
    }

    try {
        const normalizedPayload = payload.replace(/-/g, "+").replace(/_/g, "/");
        const paddedPayload = normalizedPayload.padEnd(
            normalizedPayload.length + ((4 - (normalizedPayload.length % 4)) % 4),
            "=",
        );
        const decodedPayload = JSON.parse(atob(paddedPayload)) as { sub?: string };

        return decodedPayload.sub ?? null;
    } catch {
        return null;
    }
}

export function clearAuthTokens() {
    localStorage.removeItem(loginTokenStorageKey);
    localStorage.removeItem(refreshTokenStorageKey);
}
