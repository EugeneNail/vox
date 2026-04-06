import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";

import {
    clearAuthTokens,
    getLoginToken,
    getRefreshToken,
    storeLoginToken,
} from "../auth/auth-tokens";

type RefreshResponse = {
    loginToken: string;
};

type RetriableRequestConfig = InternalAxiosRequestConfig & {
    _retry?: boolean;
};

const apiClient = axios.create({
    baseURL: "/",
    headers: {
        "Content-Type": "application/json",
    },
});

const refreshClient = axios.create({
    baseURL: "/",
    headers: {
        "Content-Type": "application/json",
    },
});

const publicRoutes = new Set([
    "/api/v1/auth/users",
    "/api/v1/auth/users/authenticate",
    "/api/v1/auth/refresh",
]);

apiClient.interceptors.request.use((config) => {
    if (!isPublicRequest(config)) {
        const loginToken = getLoginToken();
        if (loginToken) {
            config.headers.Authorization = `Bearer ${loginToken}`;
        }
    }

    return config;
});

apiClient.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
        const originalRequest = error.config as RetriableRequestConfig | undefined;

        if (!originalRequest || error.response?.status !== 401 || isPublicRequest(originalRequest) || originalRequest._retry) {
            return Promise.reject(error);
        }

        originalRequest._retry = true;

        const refreshToken = getRefreshToken();
        if (!refreshToken) {
            redirectToLogin();
            return Promise.reject(error);
        }

        try {
            const { data } = await refreshClient.post<RefreshResponse>("/api/v1/auth/refresh", {
                refreshToken,
            });

            storeLoginToken(data.loginToken);
            originalRequest.headers.Authorization = `Bearer ${data.loginToken}`;

            return apiClient(originalRequest);
        } catch (refreshError) {
            if (refreshError instanceof AxiosError && refreshError.response?.status === 401) {
                redirectToLogin();
            }

            return Promise.reject(refreshError);
        }
    },
);

function isPublicRequest(config: Pick<InternalAxiosRequestConfig, "url">) {
    if (!config.url) {
        return false;
    }

    const url = new URL(config.url, window.location.origin);
    return publicRoutes.has(url.pathname);
}

function redirectToLogin() {
    clearAuthTokens();
    window.location.assign("/login");
}

export default apiClient;
