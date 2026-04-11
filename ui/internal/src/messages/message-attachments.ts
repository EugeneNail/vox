export type MessageAttachment = {
    uuid: string;
    name: string;
};

const imageMimeTypes = new Set([
    "image/gif",
    "image/jpeg",
    "image/png",
    "image/webp",
]);

const imageExtensions = new Set([
    ".gif",
    ".jpg",
    ".jpeg",
    ".png",
    ".webp",
]);

export function isImageAttachmentName(name: string) {
    return imageExtensions.has(extractExtension(name));
}

export function isImageFile(file: File) {
    const contentType = file.type.trim().toLowerCase();
    if (imageMimeTypes.has(contentType)) {
        return true;
    }

    return isImageAttachmentName(file.name);
}

export function buildAttachmentUrl(name: string) {
    return new URL(`/api/v1/attachments/images/${encodeURIComponent(name)}`, window.location.origin).toString();
}

export function generateLocalAttachmentUuid() {
    return window.crypto?.randomUUID?.() ?? `attachment-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function extractExtension(name: string) {
    const lastDotIndex = name.lastIndexOf(".");
    if (lastDotIndex === -1) {
        return "";
    }

    return name.slice(lastDotIndex).toLowerCase();
}
