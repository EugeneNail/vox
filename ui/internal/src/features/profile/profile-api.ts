import { AxiosInstance } from "axios";
import { PublicProfile } from "../../profiles/profile-cache";

export type UpdateOwnProfileRequest = {
    name: string;
    avatar?: string | null;
};

export type BatchProfilesRequest = {
    userUuids: string[];
};

export async function loadProfilesBatch(apiClient: AxiosInstance, payload: BatchProfilesRequest) {
    const { data } = await apiClient.post<PublicProfile[]>("/api/v1/profile/profiles/batch", payload);
    return data;
}

export async function searchProfiles(apiClient: AxiosInstance, query: string, limit: number) {
    const { data } = await apiClient.get<PublicProfile[]>(`/api/v1/profile/search?query=${encodeURIComponent(query)}&limit=${limit}`);
    return data;
}

export async function updateOwnProfile(apiClient: AxiosInstance, payload: UpdateOwnProfileRequest) {
    await apiClient.put("/api/v1/profile/profiles/me", payload);
}
