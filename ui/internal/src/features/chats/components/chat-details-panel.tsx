import { type ChangeEvent, type MouseEvent as ReactMouseEvent, type PointerEvent as ReactPointerEvent, useEffect, useLayoutEffect, useRef, useState } from "react";
import { uploadImageAttachment } from "../../attachments/attachments-api";
import { type PublicProfile } from "../../../profiles/profile-cache";
import { buildAttachmentUrl, isImageFile } from "../../../messages/message-attachments";
import {
    Chat,
    UserAvatar,
    getChatAvatarUrl,
    getChatMemberAvatarLabel,
    getChatMemberAvatarUrl,
    getChatMemberName,
    getChatTitle,
} from "./chats-ui";

type ChatDetailsPanelProps = {
    authenticatedUserUuid: string | null;
    isKickingMember: boolean;
    isSelectedChatGroup: boolean;
    isSelectedChatOwnedByAuthenticatedUser: boolean;
    onKickMember: (memberUuid: string) => void;
    onOpenAddMembersModal: () => void;
    onSaveChat: (payload: { name?: string; avatar?: string; }) => Promise<void>;
    profilesByUserUuid: Record<string, PublicProfile>;
    selectedChat: Chat | null;
};

type SourceImage = {
    url: string;
    naturalWidth: number;
    naturalHeight: number;
};

type ImageGeometry = {
    imageLeft: number;
    imageTop: number;
    imageWidth: number;
    imageHeight: number;
};

type CropRect = {
    x: number;
    y: number;
    size: number;
};

type CropInteraction = {
    startX: number;
    startY: number;
    initialCrop: CropRect;
};

type MemberContextMenu = {
    memberUuid: string;
    x: number;
    y: number;
};

const defaultCropSizeRatio = 0.72;
const minimumCropSizeRatio = 0.2;
const contextMenuViewportPaddingPx = 8;
const memberContextMenuWidthPx = 220;
const memberContextMenuHeightPx = 72;

export function ChatDetailsPanel(props: ChatDetailsPanelProps) {
    const {
        authenticatedUserUuid,
        isKickingMember,
        isSelectedChatGroup,
        isSelectedChatOwnedByAuthenticatedUser,
        onKickMember,
        onOpenAddMembersModal,
        onSaveChat,
        profilesByUserUuid,
        selectedChat,
    } = props;

    const fileInputRef = useRef<HTMLInputElement | null>(null);
    const cropStageRef = useRef<HTMLDivElement | null>(null);
    const imageRef = useRef<HTMLImageElement | null>(null);
    const resizeObserverRef = useRef<ResizeObserver | null>(null);
    const cropInteractionRef = useRef<CropInteraction | null>(null);
    const [isEditingChatInfo, setIsEditingChatInfo] = useState(false);
    const [draftName, setDraftName] = useState("");
    const [draftAvatarName, setDraftAvatarName] = useState<string | null>(null);
    const [sourceImage, setSourceImage] = useState<SourceImage | null>(null);
    const [geometry, setGeometry] = useState<ImageGeometry | null>(null);
    const [cropRect, setCropRect] = useState<CropRect | null>(null);
    const [stageSize, setStageSize] = useState({ width: 0, height: 0 });
    const [isSavingChat, setIsSavingChat] = useState(false);
    const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [memberContextMenu, setMemberContextMenu] = useState<MemberContextMenu | null>(null);

    const canEditChat = Boolean(selectedChat && isSelectedChatOwnedByAuthenticatedUser);
    const readOnlyChatTitle = selectedChat ? getChatTitle(selectedChat, authenticatedUserUuid, profilesByUserUuid) : "";
    const currentChatName = selectedChat?.name ?? "";
    const currentChatAvatarName = selectedChat?.avatar ?? null;
    const displayedAvatarUrl = draftAvatarName ? buildAttachmentUrl(draftAvatarName) : selectedChat ? getChatAvatarUrl(selectedChat, authenticatedUserUuid, profilesByUserUuid) : null;
    const hasNameChange = draftName.trim().length > 0 && draftName.trim() !== currentChatName;
    const hasAvatarChange = draftAvatarName !== currentChatAvatarName;
    const hasPendingChanges = isEditingChatInfo && (hasNameChange || hasAvatarChange);
    const isBusy = isSavingChat || isUploadingAvatar;

    useEffect(() => {
        if (!selectedChat) {
            setIsEditingChatInfo(false);
            setDraftName("");
            setDraftAvatarName(null);
            setSourceImage(null);
            setGeometry(null);
            setCropRect(null);
            setError(null);
            return;
        }

        setIsEditingChatInfo(false);
        setDraftName(selectedChat.name ?? "");
        setDraftAvatarName(selectedChat.avatar ?? null);
        setSourceImage(null);
        setGeometry(null);
        setCropRect(null);
        setError(null);
        setIsSavingChat(false);
        setIsUploadingAvatar(false);
        setMemberContextMenu(null);
    }, [selectedChat?.uuid]);

    useEffect(() => {
        if (!sourceImage) {
            setGeometry(null);
            return;
        }

        const nextGeometry = buildGeometry(stageSize.width, stageSize.height, sourceImage.naturalWidth, sourceImage.naturalHeight);
        setGeometry(nextGeometry);
    }, [sourceImage, stageSize]);

    useEffect(() => {
        if (!geometry) {
            setCropRect(null);
            return;
        }

        setCropRect((currentCropRect) => clampCropRect(currentCropRect ?? buildInitialCropRect(geometry), geometry));
    }, [geometry]);

    useEffect(() => {
        return () => {
            if (sourceImage?.url.startsWith("blob:")) {
                URL.revokeObjectURL(sourceImage.url);
            }
        };
    }, [sourceImage?.url]);

    useLayoutEffect(() => {
        if (!sourceImage || !cropStageRef.current) {
            return;
        }

        const updateStageSize = () => {
            const rect = cropStageRef.current?.getBoundingClientRect();
            if (!rect) {
                return;
            }

            setStageSize({
                width: rect.width,
                height: rect.height,
            });
        };

        updateStageSize();

        if (!window.ResizeObserver) {
            window.addEventListener("resize", updateStageSize);

            return () => {
                window.removeEventListener("resize", updateStageSize);
            };
        }

        resizeObserverRef.current = new ResizeObserver(updateStageSize);
        resizeObserverRef.current.observe(cropStageRef.current);

        return () => {
            resizeObserverRef.current?.disconnect();
            resizeObserverRef.current = null;
        };
    }, [sourceImage]);

    useEffect(() => {
        function handleWindowClick() {
            setMemberContextMenu(null);
        }

        function handleWindowKeyDown(event: KeyboardEvent) {
            if (event.key === "Escape") {
                setMemberContextMenu(null);
            }
        }

        window.addEventListener("click", handleWindowClick);
        window.addEventListener("keydown", handleWindowKeyDown);

        return () => {
            window.removeEventListener("click", handleWindowClick);
            window.removeEventListener("keydown", handleWindowKeyDown);
        };
    }, []);

    function startEditMode() {
        if (!canEditChat || !selectedChat) {
            return;
        }

        setIsEditingChatInfo(true);
        setDraftName(selectedChat.name ?? "");
        setDraftAvatarName(selectedChat.avatar ?? null);
        setError(null);
    }

    function stopEditMode() {
        if (!selectedChat) {
            return;
        }

        setIsEditingChatInfo(false);
        setDraftName(selectedChat.name ?? "");
        setDraftAvatarName(selectedChat.avatar ?? null);
        clearAvatarSource();
        setError(null);
    }

    function handleEditButtonClick() {
        if (!selectedChat) {
            return;
        }

        if (isEditingChatInfo) {
            stopEditMode();
            return;
        }

        startEditMode();
    }

    function handleNameChange(event: ChangeEvent<HTMLInputElement>) {
        setDraftName(event.target.value);
        setError(null);
    }

    function handlePickAvatarFile() {
        if (!isEditingChatInfo || isBusy) {
            return;
        }

        fileInputRef.current?.click();
    }

    function handleFileInputChange(event: ChangeEvent<HTMLInputElement>) {
        const file = event.target.files?.[0];
        event.target.value = "";

        if (!file) {
            return;
        }

        void loadImageFile(file);
    }

    function handleMemberContextMenu(event: ReactMouseEvent<HTMLElement>, memberUuid: string) {
        if (!selectedChat || !isSelectedChatGroup || !isSelectedChatOwnedByAuthenticatedUser || memberUuid === authenticatedUserUuid) {
            return;
        }

        event.preventDefault();

        setMemberContextMenu({
            memberUuid,
            x: event.clientX,
            y: event.clientY,
        });
    }

    async function loadImageFile(file: File) {
        if (!isImageFile(file)) {
            setError("Choose a PNG, JPEG or WebP image.");
            return;
        }

        clearAvatarSource();

        const url = URL.createObjectURL(file);
        try {
            const dimensions = await readImageDimensions(url);
            setError(null);
            setSourceImage({
                url,
                naturalWidth: dimensions.width,
                naturalHeight: dimensions.height,
            });
        } catch {
            URL.revokeObjectURL(url);
            setError("Could not read the selected image.");
        }
    }

    function clearAvatarSource() {
        if (sourceImage?.url.startsWith("blob:")) {
            URL.revokeObjectURL(sourceImage.url);
        }

        cropInteractionRef.current = null;
        setSourceImage(null);
        setGeometry(null);
        setCropRect(null);
        setStageSize({ width: 0, height: 0 });
        setIsUploadingAvatar(false);
    }

    function handleBackdropClick(event: ReactMouseEvent<HTMLDivElement>) {
        if (event.target !== event.currentTarget) {
            return;
        }

        clearAvatarSource();
    }

    function handleCropPointerDown(event: ReactPointerEvent<HTMLDivElement>) {
        if (!cropRect || !geometry || event.button !== 0) {
            return;
        }

        event.preventDefault();
        event.currentTarget.setPointerCapture(event.pointerId);
        cropInteractionRef.current = {
            startX: event.clientX,
            startY: event.clientY,
            initialCrop: cropRect,
        };
    }

    function handleCropPointerMove(event: ReactPointerEvent<HTMLDivElement>) {
        if (!cropInteractionRef.current || !geometry) {
            return;
        }

        const { startX, startY, initialCrop } = cropInteractionRef.current;
        const deltaX = (event.clientX - startX) / geometry.imageWidth;
        const deltaY = (event.clientY - startY) / geometry.imageWidth;

        setCropRect(clampCropRect({
            ...initialCrop,
            x: initialCrop.x + deltaX,
            y: initialCrop.y + deltaY,
        }, geometry));
    }

    function handleCropPointerUp() {
        cropInteractionRef.current = null;
    }

    function handleCropSizeChange(event: ChangeEvent<HTMLInputElement>) {
        const nextSize = Number(event.target.value);
        if (!geometry || Number.isNaN(nextSize)) {
            return;
        }

        setCropRect((currentCropRect) => {
            const current = currentCropRect ?? buildInitialCropRect(geometry);
            const size = clampNumber(nextSize, minimumCropSizeRatio, getMaximumCropSizeRatio(geometry));
            const centerX = current.x + (current.size / 2);
            const centerY = current.y + (current.size / 2);

            return clampCropRect({
                size,
                x: centerX - (size / 2),
                y: centerY - (size / 2),
            }, geometry);
        });
    }

    async function handleAvatarCropConfirm() {
        if (!sourceImage || !geometry || !cropRect || !imageRef.current || isUploadingAvatar) {
            return;
        }

        setIsUploadingAvatar(true);
        setError(null);

        try {
            const avatarFile = await exportCroppedAvatar(imageRef.current, sourceImage, geometry, cropRect);
            const nextAvatarName = await uploadImageAttachment(avatarFile, "chat-avatar.png");
            setDraftAvatarName(nextAvatarName);
            clearAvatarSource();
        } catch {
            setError("Could not upload avatar.");
        } finally {
            setIsUploadingAvatar(false);
        }
    }

    async function handleSaveChat() {
        if (!selectedChat || !hasPendingChanges || isBusy) {
            return;
        }

        const nextPayload: { name?: string; avatar?: string; } = {};
        const nextName = draftName.trim();
        if (hasNameChange) {
            nextPayload.name = nextName;
        }

        if (hasAvatarChange) {
            nextPayload.avatar = draftAvatarName ?? undefined;
        }

        if (Object.keys(nextPayload).length === 0) {
            return;
        }

        setIsSavingChat(true);
        setError(null);

        try {
            await onSaveChat(nextPayload);
            setIsEditingChatInfo(false);
        } catch {
            setError("Could not save chat.");
        } finally {
            setIsSavingChat(false);
        }
    }

    return (
        <aside className="chats-me-page__details" aria-label="Chat details">
            {selectedChat ? (
                <div className={isEditingChatInfo ? "chats-me-page__details-shell chats-me-page__details-shell--editing" : "chats-me-page__details-shell"}>
                    <div className="chats-me-page__details-topbar">
                        <button
                            className="chats-me-page__details-icon-button"
                            type="button"
                            aria-label={isEditingChatInfo ? "Cancel editing" : "Close chat info"}
                            onClick={isEditingChatInfo ? stopEditMode : undefined}
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">close</span>
                        </button>
                        <h2 className="chats-me-page__details-title">{isEditingChatInfo ? "Edit chat" : "Chat Info"}</h2>
                        <button
                            className="chats-me-page__details-icon-button"
                            type="button"
                            aria-label={isEditingChatInfo ? "Close edit mode" : "Edit chat info"}
                            onClick={handleEditButtonClick}
                            disabled={!canEditChat}
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">{isEditingChatInfo ? "edit" : "edit"}</span>
                        </button>
                    </div>

                    <div className="chats-me-page__details-profile">
                        <div className="chats-me-page__details-avatar-shell">
                            {isEditingChatInfo ? (
                                <>
                                    <button
                                        className="chats-me-page__details-avatar-button"
                                        type="button"
                                        aria-label="Change chat avatar"
                                        onClick={handlePickAvatarFile}
                                        disabled={isBusy}
                                    >
                                        <UserAvatar
                                            className="chats-me-page__details-chat-avatar"
                                            src={displayedAvatarUrl}
                                            label={readOnlyChatTitle || selectedChat.uuid}
                                        />
                                        <span className="chats-me-page__details-avatar-overlay" aria-hidden="true">
                                            <span className="material-symbols-rounded chats-me-page__details-avatar-overlay-icon">add_a_photo</span>
                                        </span>
                                    </button>
                                    <input
                                        ref={fileInputRef}
                                        className="chats-me-page__details-file-input"
                                        type="file"
                                        accept="image/png,image/jpeg,image/webp"
                                        onChange={handleFileInputChange}
                                    />
                                </>
                            ) : (
                                <UserAvatar
                                    className="chats-me-page__details-chat-avatar"
                                    src={displayedAvatarUrl}
                                    label={readOnlyChatTitle || selectedChat.uuid}
                                />
                            )}
                        </div>

                        {isEditingChatInfo ? (
                            <div className="chats-me-page__details-name-field">
                                <label className="chats-me-page__details-chat-name-label" htmlFor="chat-name-input">
                                    Group name
                                </label>
                                <input
                                    id="chat-name-input"
                                    className="chats-me-page__details-chat-name-input"
                                    type="text"
                                    value={draftName}
                                    onChange={handleNameChange}
                                    autoComplete="off"
                                />
                                <p className="chats-me-page__details-chat-meta">Changes stay local until you confirm them.</p>
                            </div>
                        ) : (
                            <h3 className="chats-me-page__details-chat-name">{readOnlyChatTitle}</h3>
                        )}

                        <p className="chats-me-page__details-chat-meta">{selectedChat.memberUuids.length} members</p>
                    </div>

                    <div className="chats-me-page__members-panel">
                        <div className="chats-me-page__members" aria-label="Chat members">
                            {selectedChat.memberUuids.map((memberUuid) => (
                                <div
                                    className={
                                        [
                                            "chats-me-page__member",
                                            selectedChat.createdByUserUuid === memberUuid ? "chats-me-page__member--owner" : "",
                                            isSelectedChatGroup && isSelectedChatOwnedByAuthenticatedUser && memberUuid !== authenticatedUserUuid ? "chats-me-page__member--removable" : "",
                                        ].filter(Boolean).join(" ")
                                    }
                                    key={memberUuid}
                                    onContextMenu={(event) => handleMemberContextMenu(event, memberUuid)}
                                >
                                    <UserAvatar
                                        className="chats-me-page__member-avatar"
                                        src={getChatMemberAvatarUrl(memberUuid, profilesByUserUuid)}
                                        label={getChatMemberAvatarLabel(memberUuid, profilesByUserUuid)}
                                    />
                                    <span className="chats-me-page__member-name">
                                        {getChatMemberName(memberUuid, profilesByUserUuid)}
                                    </span>
                                    <div className="chats-me-page__member-badges">
                                        {selectedChat.createdByUserUuid === memberUuid && (
                                            <span className="chats-me-page__member-role">owner</span>
                                        )}
                                        {memberUuid === authenticatedUserUuid && (
                                            <span className="chats-me-page__member-role chats-me-page__member-role--me">me</span>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {memberContextMenu && (
                        <div
                            className="chats-me-page__context-menu"
                            style={clampContextMenuPosition(memberContextMenu.x, memberContextMenu.y, memberContextMenuWidthPx, memberContextMenuHeightPx)}
                            onClick={(event) => event.stopPropagation()}
                        >
                            <button
                                className="chats-me-page__context-menu-button chats-me-page__context-menu-button--danger"
                                type="button"
                                disabled={isKickingMember}
                                onClick={() => {
                                    setMemberContextMenu(null);
                                    onKickMember(memberContextMenu.memberUuid);
                                }}
                            >
                                <span className="material-symbols-rounded" aria-hidden="true">person_remove</span>
                                Kick User
                            </button>
                        </div>
                    )}

                    {isSelectedChatGroup && !isEditingChatInfo && (
                        <button
                            className="chats-me-page__details-floating-action"
                            type="button"
                            onClick={onOpenAddMembersModal}
                            aria-label="Add members"
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">person_add</span>
                        </button>
                    )}

                    {isEditingChatInfo && hasPendingChanges && (
                        <button
                            className="chats-me-page__details-floating-action chats-me-page__details-floating-action--save"
                            type="button"
                            onClick={() => void handleSaveChat()}
                            disabled={isBusy}
                            aria-label="Save chat changes"
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">{isSavingChat ? "progress_activity" : "check"}</span>
                        </button>
                    )}

                    {sourceImage && (
                        <div className="chats-me-page__avatar-modal" role="presentation" onClick={handleBackdropClick}>
                            <section
                                className="chats-me-page__avatar-dialog"
                                role="dialog"
                                aria-modal="true"
                                aria-labelledby="chat-avatar-dialog-title"
                            >
                                <div className="chats-me-page__avatar-dialog-header">
                                    <div>
                                        <p className="chats-me-page__avatar-dialog-eyebrow">Avatar</p>
                                        <h3 className="chats-me-page__avatar-dialog-title" id="chat-avatar-dialog-title">Crop the new avatar</h3>
                                    </div>
                                    <button
                                        className="chats-me-page__details-icon-button"
                                        type="button"
                                        aria-label="Close avatar cropper"
                                        onClick={clearAvatarSource}
                                    >
                                        <span className="material-symbols-rounded" aria-hidden="true">close</span>
                                    </button>
                                </div>

                                <div
                                    ref={cropStageRef}
                                    className="chats-me-page__avatar-stage"
                                    onPointerMove={handleCropPointerMove}
                                    onPointerUp={handleCropPointerUp}
                                    onPointerCancel={handleCropPointerUp}
                                >
                                    <img
                                        ref={imageRef}
                                        className="chats-me-page__avatar-stage-image"
                                        src={sourceImage.url}
                                        alt="Selected avatar source"
                                        draggable={false}
                                    />

                                    {cropRect && geometry && (
                                        <div
                                            className="chats-me-page__avatar-crop"
                                            style={getCropRectStyle(cropRect, geometry)}
                                            onPointerDown={handleCropPointerDown}
                                        />
                                    )}
                                </div>

                                <div className="chats-me-page__avatar-controls">
                                    <label className="chats-me-page__avatar-size">
                                        <span className="chats-me-page__avatar-size-label">Zoom</span>
                                        <input
                                            className="chats-me-page__avatar-size-input"
                                            type="range"
                                            min={minimumCropSizeRatio}
                                            max={geometry ? getMaximumCropSizeRatio(geometry) : 0.84}
                                            step="0.01"
                                            value={cropRect?.size ?? defaultCropSizeRatio}
                                            onChange={handleCropSizeChange}
                                            disabled={!geometry || !cropRect}
                                        />
                                    </label>

                                    <button
                                        className="chats-me-page__avatar-confirm"
                                        type="button"
                                        onClick={() => void handleAvatarCropConfirm()}
                                        disabled={!geometry || !cropRect || isUploadingAvatar}
                                        aria-label="Confirm avatar crop"
                                    >
                                        <span className="material-symbols-rounded" aria-hidden="true">
                                            {isUploadingAvatar ? "progress_activity" : "check"}
                                        </span>
                                    </button>
                                </div>
                            </section>
                        </div>
                    )}

                    {error && (
                        <p className="chats-me-page__details-feedback chats-me-page__details-feedback--error">{error}</p>
                    )}
                </div>
            ) : (
                <div className="chats-me-page__details-shell chats-me-page__details-shell--empty">
                    <p className="chats-me-page__state">Choose a chat to see members.</p>
                </div>
            )}
        </aside>
    );
}

function buildGeometry(stageWidth: number, stageHeight: number, imageWidth: number, imageHeight: number) {
    if (stageWidth <= 0 || stageHeight <= 0 || imageWidth <= 0 || imageHeight <= 0) {
        return null;
    }

    const scale = Math.min(stageWidth / imageWidth, stageHeight / imageHeight);
    const displayedWidth = imageWidth * scale;
    const displayedHeight = imageHeight * scale;

    return {
        imageLeft: (stageWidth - displayedWidth) / 2,
        imageTop: (stageHeight - displayedHeight) / 2,
        imageWidth: displayedWidth,
        imageHeight: displayedHeight,
    };
}

function buildInitialCropRect(geometry: ImageGeometry): CropRect {
    const maxSize = getMaximumCropSizeRatio(geometry);
    const size = clampNumber(defaultCropSizeRatio, minimumCropSizeRatio, maxSize);
    const maxY = Math.max(0, geometry.imageHeight / geometry.imageWidth - size);

    return {
        x: (1 - size) / 2,
        y: maxY / 2,
        size,
    };
}

function clampCropRect(cropRect: CropRect, geometry: ImageGeometry) {
    const maxSize = getMaximumCropSizeRatio(geometry);
    const size = clampNumber(cropRect.size, minimumCropSizeRatio, maxSize);
    const maxX = Math.max(0, 1 - size);
    const maxY = Math.max(0, geometry.imageHeight / geometry.imageWidth - size);

    return {
        size,
        x: clampNumber(cropRect.x, 0, maxX),
        y: clampNumber(cropRect.y, 0, maxY),
    };
}

function getCropRectStyle(cropRect: CropRect, geometry: ImageGeometry) {
    const left = geometry.imageLeft + (cropRect.x * geometry.imageWidth);
    const top = geometry.imageTop + (cropRect.y * geometry.imageWidth);
    const size = cropRect.size * geometry.imageWidth;

    return {
        left: `${left}px`,
        top: `${top}px`,
        width: `${size}px`,
        height: `${size}px`,
    };
}

function getMaximumCropSizeRatio(geometry: ImageGeometry) {
    return Math.min(geometry.imageHeight / geometry.imageWidth, 0.84);
}

async function exportCroppedAvatar(image: HTMLImageElement | null, sourceImage: SourceImage, geometry: ImageGeometry, cropRect: CropRect) {
    if (!image) {
        throw new Error("selected image is unavailable");
    }

    const canvas = document.createElement("canvas");
    const outputSize = 512;
    canvas.width = outputSize;
    canvas.height = outputSize;

    const context = canvas.getContext("2d");
    if (!context) {
        throw new Error("canvas context is unavailable");
    }

    context.imageSmoothingQuality = "high";
    const sourceScale = sourceImage.naturalWidth / geometry.imageWidth;
    const sourceX = cropRect.x * geometry.imageWidth * sourceScale;
    const sourceY = cropRect.y * geometry.imageWidth * sourceScale;
    const sourceSize = cropRect.size * geometry.imageWidth * sourceScale;
    context.drawImage(image, sourceX, sourceY, sourceSize, sourceSize, 0, 0, outputSize, outputSize);

    const blob = await new Promise<Blob>((resolve, reject) => {
        canvas.toBlob((nextBlob) => {
            if (!nextBlob) {
                reject(new Error("could not export cropped avatar"));
                return;
            }

            resolve(nextBlob);
        }, "image/png");
    });

    return new File([blob], buildAvatarFileName(), { type: "image/png" });
}

function buildAvatarFileName() {
    return `chat-avatar-${Date.now()}.png`;
}

function clampContextMenuPosition(x: number, y: number, menuWidth: number, menuHeight: number) {
    if (typeof window === "undefined") {
        return { left: x, top: y };
    }

    const maxLeft = Math.max(contextMenuViewportPaddingPx, window.innerWidth - menuWidth - contextMenuViewportPaddingPx);
    const maxTop = Math.max(contextMenuViewportPaddingPx, window.innerHeight - menuHeight - contextMenuViewportPaddingPx);

    return {
        left: Math.min(Math.max(x, contextMenuViewportPaddingPx), maxLeft),
        top: Math.min(Math.max(y, contextMenuViewportPaddingPx), maxTop),
    };
}

function clampNumber(value: number, min: number, max: number) {
    if (max < min) {
        return max;
    }

    return Math.min(max, Math.max(min, value));
}

function readImageDimensions(url: string) {
    return new Promise<{ width: number; height: number }>((resolve, reject) => {
        const image = new Image();
        image.onload = () => {
            resolve({
                width: image.naturalWidth,
                height: image.naturalHeight,
            });
        };
        image.onerror = () => {
            reject(new Error("could not load image"));
        };
        image.src = url;
    });
}
