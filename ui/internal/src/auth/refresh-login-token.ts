import axios, { AxiosError } from "axios";

import { clearAuthTokens, getLoginToken, getRefreshToken, hasLoginTokenExpired, storeLoginToken } from "./auth-tokens";

type RefreshResponse = {
    loginToken: string;
};

const refreshClient = axios.create({
    baseURL: "/",
    headers: {
        "Content-Type": "application/json",
    },
});

let refreshLoginTokenPromise: Promise<string> | null = null;

export function refreshLoginToken() {
    if (refreshLoginTokenPromise) {
        return refreshLoginTokenPromise;
    }

    refreshLoginTokenPromise = refreshLoginTokenOnce().finally(() => {
        refreshLoginTokenPromise = null;
    });

    return refreshLoginTokenPromise;
}

export async function getValidLoginToken() {
    const loginToken = getLoginToken();
    if (loginToken && !hasLoginTokenExpired(loginToken)) {
        return loginToken;
    }

    return refreshLoginToken();
}

export function redirectToLogin() {
    clearAuthTokens();
    window.location.assign("/login");
}

async function refreshLoginTokenOnce() {
    const refreshToken = getRefreshToken();
    if (!refreshToken) {
        redirectToLogin();
        throw new Error("refresh token is missing");
    }

    try {
        const { data } = await refreshClient.post<RefreshResponse>("/api/v1/auth/refresh", {
            refreshToken,
        });

        storeLoginToken(data.loginToken);

        return data.loginToken;
    } catch (error) {
        if (error instanceof AxiosError && error.response?.status === 401) {
            redirectToLogin();
        }

        throw error;
    }
}
