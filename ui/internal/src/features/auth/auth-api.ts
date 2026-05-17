import { AxiosInstance } from "axios";

export type AuthenticateRequest = {
    email: string;
    password: string;
};

export type RegisterUserRequest = {
    name: string;
    email: string;
    password: string;
    passwordConfirmation: string;
};

export type AuthTokensResponse = {
    loginToken: string;
    refreshToken: string;
};

export async function authenticateUser(apiClient: AxiosInstance, payload: AuthenticateRequest) {
    const { data } = await apiClient.post<AuthTokensResponse>("/api/v1/auth/users/authenticate", payload);
    return data;
}

export async function registerUser(apiClient: AxiosInstance, payload: RegisterUserRequest) {
    await apiClient.post("/api/v1/auth/users", payload);
}
