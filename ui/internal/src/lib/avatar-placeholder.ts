import { type CSSProperties } from "react";

export type AvatarPlaceholder = {
    initials: string;
    style: CSSProperties;
};

export function buildAvatarPlaceholder(label: string): AvatarPlaceholder {
    const initials = buildAvatarInitials(label);
    const backgroundColor = hashInitialsToHex(initials);

    return {
        initials,
        style: {
            background: backgroundColor,
            color: "#f8fafc",
        },
    };
}

function buildAvatarInitials(label: string) {
    const normalizedLabel = label.trim().replace(/\s+/g, " ");
    if (normalizedLabel.length === 0) {
        return "??";
    }

    const words = normalizedLabel.split(" ").filter(Boolean);
    if (words.length >= 2) {
        return `${firstSymbol(words[0])}${firstSymbol(words[1])}`.toUpperCase();
    }

    const firstWord = words[0] ?? normalizedLabel;
    const symbols = Array.from(firstWord).filter((symbol) => /\p{L}|\p{N}/u.test(symbol));
    if (symbols.length >= 2) {
        return `${symbols[0]}${symbols[1]}`.toUpperCase();
    }

    if (symbols.length === 1) {
        return `${symbols[0]}${symbols[0]}`.toUpperCase();
    }

    return normalizedLabel.slice(0, 2).toUpperCase().padEnd(2, "?");
}

function firstSymbol(value: string) {
    const symbols = Array.from(value).filter((symbol) => /\p{L}|\p{N}/u.test(symbol));
    return symbols[0] ?? "?";
}

function hashInitialsToHex(initials: string) {
    let hash = 0;

    for (const symbol of initials.toUpperCase()) {
        hash = ((hash << 5) - hash + symbol.charCodeAt(0)) | 0;
    }

    const hue = Math.abs(hash) % 360;
    return hslToHex(hue, 0.56, 0.46);
}

function hslToHex(hue: number, saturation: number, lightness: number) {
    const chroma = (1 - Math.abs(2 * lightness - 1)) * saturation;
    const hueSection = hue / 60;
    const secondary = chroma * (1 - Math.abs((hueSection % 2) - 1));
    let red = 0;
    let green = 0;
    let blue = 0;

    if (hueSection >= 0 && hueSection < 1) {
        red = chroma;
        green = secondary;
    } else if (hueSection < 2) {
        red = secondary;
        green = chroma;
    } else if (hueSection < 3) {
        green = chroma;
        blue = secondary;
    } else if (hueSection < 4) {
        green = secondary;
        blue = chroma;
    } else if (hueSection < 5) {
        red = secondary;
        blue = chroma;
    } else {
        red = chroma;
        blue = secondary;
    }

    const matchValue = lightness - chroma / 2;
    const toHexByte = (channel: number) => Math.round((channel + matchValue) * 255).toString(16).padStart(2, "0");

    return `#${toHexByte(red)}${toHexByte(green)}${toHexByte(blue)}`;
}
