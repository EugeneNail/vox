export const loginTokenStorageKey = "vox.loginToken";
export const refreshTokenStorageKey = "vox.refreshToken";
export const authTokensChangedEventName = "vox:auth-tokens-changed";
const loginTokenExpirationSkewMs = 30_000;

type AuthTokens = {
    loginToken: string;
    refreshToken: string;
};

export function storeAuthTokens(tokens: AuthTokens) {
    localStorage.setItem(loginTokenStorageKey, tokens.loginToken);
    localStorage.setItem(refreshTokenStorageKey, tokens.refreshToken);
    notifyAuthTokensChanged();
}

export function storeLoginToken(loginToken: string) {
    localStorage.setItem(loginTokenStorageKey, loginToken);
    notifyAuthTokensChanged();
}

export function getLoginToken() {
    return localStorage.getItem(loginTokenStorageKey);
}

export function getRefreshToken() {
    return localStorage.getItem(refreshTokenStorageKey);
}

export function hasLoginTokenExpired(loginToken: string) {
    const expiresAt = getTokenExpiresAt(loginToken);

    return expiresAt === null || expiresAt - loginTokenExpirationSkewMs <= Date.now();
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
    notifyAuthTokensChanged();
}

function notifyAuthTokensChanged() {
    window.dispatchEvent(new Event(authTokensChangedEventName));
}

function getTokenExpiresAt(loginToken: string) {
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
        const decodedPayload = JSON.parse(atob(paddedPayload)) as { exp?: number };
        if (!decodedPayload.exp) {
            return null;
        }

        return decodedPayload.exp * 1000;
    } catch {
        return null;
    }
}
