import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getAuthenticatedUserUuid } from "../../../auth/auth-tokens";
import { useApiClient } from "../../../hooks/use-api-client";
import { PublicProfile } from "../../../profiles/profile-cache";
import { loadProfilesBatch } from "../profile-api";

export function useOwnProfile() {
    const apiClient = useApiClient();
    const navigate = useNavigate();
    const authenticatedUserUuid = getAuthenticatedUserUuid();
    const [profile, setProfile] = useState<PublicProfile | null>(null);
    const [name, setName] = useState("");
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (authenticatedUserUuid) {
            return;
        }

        navigate("/login", { replace: true });
    }, [authenticatedUserUuid, navigate]);

    useEffect(() => {
        if (!authenticatedUserUuid) {
            return;
        }

        let isMounted = true;
        const currentAuthenticatedUserUuid = authenticatedUserUuid;

        async function loadProfile() {
            setIsLoading(true);
            setError(null);

            try {
                const data = await loadProfilesBatch(apiClient, {
                    userUuids: [currentAuthenticatedUserUuid],
                });

                if (!isMounted) {
                    return;
                }

                const nextProfile = data[0] ?? null;
                if (!nextProfile) {
                    setError("Profile was not found.");
                    return;
                }

                setProfile(nextProfile);
                setName(nextProfile.name);
            } catch {
                if (isMounted) {
                    setError("Could not load profile.");
                }
            } finally {
                if (isMounted) {
                    setIsLoading(false);
                }
            }
        }

        void loadProfile();

        return () => {
            isMounted = false;
        };
    }, [apiClient, authenticatedUserUuid]);

    return {
        authenticatedUserUuid,
        profile,
        setProfile,
        name,
        setName,
        isLoading,
        setIsLoading,
        error,
        setError,
    };
}
