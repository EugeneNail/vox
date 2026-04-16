export type PublicProfile = {
    userUuid: string;
    avatar: string | null;
    name: string;
};

type CachedProfile = PublicProfile & {
    staleAt: string;
};

const profileCacheStorageKey = "profile-cache";
const profileCacheTtlMs = 60 * 60 * 1000;

let cachedProfilesByUserUuid: Record<string, CachedProfile> | null = null;

export function getCachedProfilesByUserUuids(userUuids: string[]) {
    const cache = getProfileCache();

    return userUuids.reduce<Record<string, PublicProfile>>((profilesByUserUuid, userUuid) => {
        const profile = cache[userUuid];
        if (!profile) {
            return profilesByUserUuid;
        }

        profilesByUserUuid[userUuid] = toPublicProfile(profile);
        return profilesByUserUuid;
    }, {});
}

export function getStaleCachedUserUuids(now = Date.now()) {
    const cache = getProfileCache();

    return Object.values(cache)
        .filter((profile) => Date.parse(profile.staleAt) <= now)
        .map((profile) => profile.userUuid);
}

export function getMissingOrStaleUserUuids(userUuids: string[], now = Date.now()) {
    const cache = getProfileCache();

    return Array.from(new Set(userUuids)).filter((userUuid) => {
        const profile = cache[userUuid];
        if (!profile) {
            return true;
        }

        return Date.parse(profile.staleAt) <= now;
    });
}

export function upsertProfiles(profiles: PublicProfile[], now = Date.now()) {
    const cache = getProfileCache();

    for (const profile of profiles) {
        cache[profile.userUuid] = {
            ...profile,
            staleAt: new Date(now + profileCacheTtlMs).toISOString(),
        };
    }

    persistProfileCache();
}

function getProfileCache() {
    if (cachedProfilesByUserUuid !== null) {
        return cachedProfilesByUserUuid;
    }

    try {
        const rawCache = localStorage.getItem(profileCacheStorageKey);
        if (!rawCache) {
            cachedProfilesByUserUuid = {};
            return cachedProfilesByUserUuid;
        }

        const parsedCache = JSON.parse(rawCache) as Record<string, CachedProfile>;
        cachedProfilesByUserUuid = parsedCache ?? {};
        return cachedProfilesByUserUuid;
    } catch {
        cachedProfilesByUserUuid = {};
        return cachedProfilesByUserUuid;
    }
}

function persistProfileCache() {
    if (cachedProfilesByUserUuid === null) {
        return;
    }

    localStorage.setItem(profileCacheStorageKey, JSON.stringify(cachedProfilesByUserUuid));
}

function toPublicProfile(profile: CachedProfile): PublicProfile {
    return {
        userUuid: profile.userUuid,
        avatar: profile.avatar,
        name: profile.name,
    };
}
