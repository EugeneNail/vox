import { AxiosError } from "axios";
import { ChangeEvent, DragEvent, FormEvent, PointerEvent as ReactPointerEvent, useEffect, useLayoutEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getAuthenticatedUserUuid } from "../../auth/auth-tokens";
import FormSubmitButton from "../../components/form-submit-button/form-submit-button";
import FormTextField from "../../components/form-text-field/form-text-field";
import { useApiClient } from "../../hooks/use-api-client";
import { buildAttachmentUrl, isImageFile } from "../../messages/message-attachments";
import "./profile-page.sass";

type PublicProfile = {
    userUuid: string;
    avatar: string | null;
    name: string;
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
    kind: "move" | "resize";
    startX: number;
    startY: number;
    initialCrop: CropRect;
};

const defaultCropSizeRatio = 0.72;
const minimumCropSizeRatio = 0.2;
export default function ProfilePage() {
    const apiClient = useApiClient();
    const navigate = useNavigate();
    const dropzoneRef = useRef<HTMLDivElement | null>(null);
    const cropStageRef = useRef<HTMLDivElement | null>(null);
    const fileInputRef = useRef<HTMLInputElement | null>(null);
    const imageRef = useRef<HTMLImageElement | null>(null);
    const dragDepthRef = useRef(0);
    const resizeObserverRef = useRef<ResizeObserver | null>(null);
    const cropInteractionRef = useRef<CropInteraction | null>(null);
    const [profile, setProfile] = useState<PublicProfile | null>(null);
    const [name, setName] = useState("");
    const [sourceImage, setSourceImage] = useState<SourceImage | null>(null);
    const [geometry, setGeometry] = useState<ImageGeometry | null>(null);
    const [cropRect, setCropRect] = useState<CropRect | null>(null);
    const [uploadedAvatarName, setUploadedAvatarName] = useState<string | undefined>(undefined);
    const [previewAvatarUrl, setPreviewAvatarUrl] = useState<string | null>(null);
    const [stageSize, setStageSize] = useState({ width: 0, height: 0 });
    const [isSourceImageReady, setIsSourceImageReady] = useState(true);
    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [isDraggingFiles, setIsDraggingFiles] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [status, setStatus] = useState<string | null>(null);
    const authenticatedUserUuid = getAuthenticatedUserUuid();
    const resolvedAvatarName = uploadedAvatarName ?? profile?.avatar ?? null;
    const canSave = !isLoading && !isSaving && Boolean(profile) && Boolean(name.trim()) && (!sourceImage || isSourceImageReady);

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

        async function loadProfile() {
            setIsLoading(true);
            setError(null);

            try {
                const { data } = await apiClient.post<PublicProfile[]>("/api/v1/profile/profiles/batch", {
                    userUuids: [authenticatedUserUuid],
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
                setIsSourceImageReady(true);
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
        let isMounted = true;
        let previewBlobUrl: string | null = null;

        async function refreshPreviewAvatarUrl() {
            if (sourceImage && geometry && cropRect) {
                try {
                    previewBlobUrl = await buildCroppedAvatarPreviewUrl(sourceImage, geometry, cropRect);
                    if (!isMounted) {
                        URL.revokeObjectURL(previewBlobUrl);
                        return;
                    }

                    setPreviewAvatarUrl(previewBlobUrl);
                    previewBlobUrl = null;
                    return;
                } catch {
                    if (isMounted) {
                        setPreviewAvatarUrl(null);
                    }

                    return;
                }
            }

            if (resolvedAvatarName) {
                if (isMounted) {
                    setPreviewAvatarUrl(buildAttachmentUrl(resolvedAvatarName));
                }

                return;
            }

            if (isMounted) {
                setPreviewAvatarUrl(null);
            }
        }

        void refreshPreviewAvatarUrl();

        return () => {
            isMounted = false;
            if (previewBlobUrl) {
                URL.revokeObjectURL(previewBlobUrl);
            }
        };
    }, [cropRect, geometry, resolvedAvatarName, sourceImage]);

    useEffect(() => {
        return () => {
            if (previewAvatarUrl?.startsWith("blob:")) {
                URL.revokeObjectURL(previewAvatarUrl);
            }
        };
    }, [previewAvatarUrl]);

    useEffect(() => {
        return () => {
            if (sourceImage?.url.startsWith("blob:")) {
                URL.revokeObjectURL(sourceImage.url);
            }
        };
    }, [sourceImage?.url]);

    useEffect(() => {
        function handlePaste(event: ClipboardEvent) {
            const file = extractPastedImageFile(event.clipboardData);
            if (!file) {
                return;
            }

            event.preventDefault();
            void loadImageFile(file);
        }

        window.addEventListener("paste", handlePaste);

        return () => {
            window.removeEventListener("paste", handlePaste);
        };
    }, []);

    useLayoutEffect(() => {
        if (!cropStageRef.current) {
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
    }, []);

    async function loadImageFile(file: File) {
        if (!isImageFile(file)) {
            setError("Choose a PNG, JPEG or WebP image.");
            return;
        }

        if (sourceImage?.url.startsWith("blob:")) {
            URL.revokeObjectURL(sourceImage.url);
        }

        setError(null);
        setStatus(null);
        setUploadedAvatarName(undefined);
        setIsSourceImageReady(false);
        cropInteractionRef.current = null;
        setCropRect(null);

        const url = URL.createObjectURL(file);
        try {
            const dimensions = await readImageDimensions(url);
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

    function handleFileInputChange(event: ChangeEvent<HTMLInputElement>) {
        const file = event.target.files?.[0];
        event.target.value = "";

        if (!file) {
            return;
        }

        void loadImageFile(file);
    }

    function handlePickImage() {
        fileInputRef.current?.click();
    }

    function handleDragEnter(event: DragEvent<HTMLDivElement>) {
        if (!hasDraggedFiles(event)) {
            return;
        }

        event.preventDefault();
        dragDepthRef.current += 1;
        setIsDraggingFiles(true);
    }

    function handleDragOver(event: DragEvent<HTMLDivElement>) {
        if (!hasDraggedFiles(event)) {
            return;
        }

        event.preventDefault();
    }

    function handleDragLeave(event: DragEvent<HTMLDivElement>) {
        if (!hasDraggedFiles(event)) {
            return;
        }

        event.preventDefault();
        dragDepthRef.current = Math.max(0, dragDepthRef.current - 1);

        if (dragDepthRef.current === 0) {
            setIsDraggingFiles(false);
        }
    }

    function handleDrop(event: DragEvent<HTMLDivElement>) {
        if (!hasDraggedFiles(event)) {
            return;
        }

        event.preventDefault();
        dragDepthRef.current = 0;
        setIsDraggingFiles(false);

        const file = Array.from(event.dataTransfer.files ?? []).find(isImageFile);
        if (!file) {
            setError("Drop a PNG, JPEG or WebP image.");
            return;
        }

        void loadImageFile(file);
    }

    function handleImagePointerDown(event: ReactPointerEvent<HTMLDivElement>) {
        if (!cropRect || !geometry || event.button !== 0) {
            return;
        }

        event.preventDefault();
        event.currentTarget.setPointerCapture(event.pointerId);
        cropInteractionRef.current = {
            kind: "move",
            startX: event.clientX,
            startY: event.clientY,
            initialCrop: cropRect,
        };
    }

    function handleResizePointerDown(event: ReactPointerEvent<HTMLButtonElement>) {
        if (!cropRect || !geometry || event.button !== 0) {
            return;
        }

        event.preventDefault();
        event.stopPropagation();
        event.currentTarget.setPointerCapture(event.pointerId);
        cropInteractionRef.current = {
            kind: "resize",
            startX: event.clientX,
            startY: event.clientY,
            initialCrop: cropRect,
        };
    }

    function handleCropPointerMove(event: ReactPointerEvent<HTMLDivElement>) {
        if (!cropInteractionRef.current || !geometry) {
            return;
        }

        const { kind, startX, startY, initialCrop } = cropInteractionRef.current;
        const deltaX = (event.clientX - startX) / geometry.imageWidth;
        const deltaY = (event.clientY - startY) / geometry.imageWidth;

        if (kind === "move") {
            if (uploadedAvatarName) {
                setUploadedAvatarName(undefined);
            }

            setCropRect(clampCropRect({
                ...initialCrop,
                x: initialCrop.x + deltaX,
                y: initialCrop.y + deltaY,
            }, geometry));
            return;
        }

        if (uploadedAvatarName) {
            setUploadedAvatarName(undefined);
        }

        setCropRect(clampCropRect({
            ...initialCrop,
            size: initialCrop.size + ((deltaX + deltaY) / 2),
        }, geometry));
    }

    function handleCropPointerUp() {
        cropInteractionRef.current = null;
    }

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        if (!profile || !name.trim() || isSaving) {
            return;
        }

        setIsSaving(true);
        setError(null);
        setStatus(null);

        try {
            let nextAvatarName = uploadedAvatarName;
            if (sourceImage && geometry && cropRect && !nextAvatarName) {
                const avatarFile = await exportCroppedAvatar(imageRef.current, sourceImage, geometry, cropRect);
                nextAvatarName = await uploadAvatar(avatarFile);
                setUploadedAvatarName(nextAvatarName);
            }

            await apiClient.put("/api/v1/profile/profiles/me", {
                name: name.trim(),
                ...(nextAvatarName ? { avatar: nextAvatarName } : {}),
            });

            setProfile((currentProfile) => currentProfile ? {
                ...currentProfile,
                name: name.trim(),
                avatar: nextAvatarName ?? currentProfile.avatar,
            } : currentProfile);
            setStatus("Profile saved.");

            if (sourceImage?.url.startsWith("blob:")) {
                URL.revokeObjectURL(sourceImage.url);
            }

            setSourceImage(null);
            setGeometry(null);
            setCropRect(null);
            setUploadedAvatarName(undefined);
            setIsSourceImageReady(true);
        } catch (saveError) {
            if (saveError instanceof AxiosError && saveError.response?.status === 422) {
                setError("Name must contain 2 to 64 characters.");
            } else {
                setError("Could not save profile.");
            }
        } finally {
            setIsSaving(false);
        }
    }

    return (
        <section className="profile-page">
            <div className="profile-page__backdrop" aria-hidden="true" />
            <div className="profile-page__glow profile-page__glow--left" aria-hidden="true" />
            <div className="profile-page__glow profile-page__glow--right" aria-hidden="true" />

            <div className="profile-page__shell">
                <header className="profile-page__header">
                    <p className="profile-page__eyebrow">Profile editor</p>
                    <h1 className="profile-page__title">Edit your public profile.</h1>
                    <p className="profile-page__text">
                        Pick an image, choose the square crop and save the same avatar your contacts will see in chat.
                    </p>
                </header>

                <div className="profile-page__content">
                    <form className="profile-page__editor" onSubmit={handleSubmit}>
                        <FormTextField
                            label="Display name"
                            name="name"
                            type="text"
                            autoComplete="name"
                            placeholder="Your display name"
                            value={name}
                            error={undefined}
                            onChange={(event) => {
                                setName(event.target.value);
                                setError(null);
                                setStatus(null);
                            }}
                        />

                        <div
                            ref={dropzoneRef}
                            className={`profile-page__dropzone${isDraggingFiles ? " profile-page__dropzone--active" : ""}`}
                            onDragEnter={handleDragEnter}
                            onDragOver={handleDragOver}
                            onDragLeave={handleDragLeave}
                            onDrop={handleDrop}
                        >
                            <input
                                ref={fileInputRef}
                                className="profile-page__file-input"
                                type="file"
                                accept="image/png,image/jpeg,image/webp"
                                onChange={handleFileInputChange}
                            />

                            <div className="profile-page__dropzone-copy">
                                <p className="profile-page__dropzone-title">Avatar source</p>
                                <p className="profile-page__dropzone-text">
                                    Drop an image here, paste it from the clipboard or choose it from your files.
                                </p>
                                <div className="profile-page__dropzone-actions">
                                    <button className="profile-page__button profile-page__button--primary" type="button" onClick={handlePickImage}>
                                        Choose file
                                    </button>
                                    <span className="profile-page__hint">PNG, JPEG, WebP</span>
                                </div>
                            </div>

                            <div className="profile-page__dropzone-preview">
                                <AvatarCircle
                                    avatarUrl={previewAvatarUrl}
                                    name={name.trim() || profile?.name || "Profile"}
                                    size={128}
                                />
                                <p className="profile-page__dropzone-preview-text">
                                    {sourceImage ? "Crop the selected image below before saving." : "The current avatar is shown here until you choose a new one."}
                                </p>
                            </div>

                            {isDraggingFiles && (
                                <div className="profile-page__dropzone-overlay">
                                    <span className="material-symbols-rounded profile-page__dropzone-overlay-icon" aria-hidden="true">upload</span>
                                    <p className="profile-page__dropzone-overlay-text">Drop the image to start cropping</p>
                                </div>
                            )}
                        </div>

                        <div className="profile-page__cropper">
                            <div className="profile-page__cropper-header">
                                <div>
                                    <p className="profile-page__section-title">Square crop</p>
                                    <p className="profile-page__section-text">
                                        Drag the crop inside the image or pull the corner handle to resize it.
                                    </p>
                                </div>
                                {sourceImage && (
                                    <button
                                        className="profile-page__button profile-page__button--secondary"
                                        type="button"
                                        onClick={() => {
                                            if (sourceImage.url.startsWith("blob:")) {
                                                URL.revokeObjectURL(sourceImage.url);
                                            }

                                            cropInteractionRef.current = null;
                                            setSourceImage(null);
                                            setGeometry(null);
                                            setCropRect(null);
                                            setUploadedAvatarName(undefined);
                                            setIsSourceImageReady(true);
                                            setStatus(null);
                                            setError(null);
                                        }}
                                    >
                                        Clear image
                                    </button>
                                )}
                            </div>

                            <div
                                ref={cropStageRef}
                                className={`profile-page__crop-stage${sourceImage ? " profile-page__crop-stage--ready" : ""}`}
                                onPointerMove={handleCropPointerMove}
                                onPointerUp={handleCropPointerUp}
                                onPointerCancel={handleCropPointerUp}
                            >
                                {sourceImage ? (
                                    <div className="profile-page__crop-stage-inner">
                                        <img
                                            ref={imageRef}
                                            className="profile-page__crop-image"
                                            src={sourceImage.url}
                                            alt="Selected avatar source"
                                            draggable={false}
                                            onLoad={() => {
                                                setIsSourceImageReady(true);
                                            }}
                                            onError={() => {
                                                setIsSourceImageReady(false);
                                                setError("Could not load the selected image.");
                                            }}
                                        />

                                        {cropRect && geometry && (
                                            <div
                                                className="profile-page__crop-rect"
                                                style={getCropRectStyle(cropRect, geometry)}
                                                onPointerDown={handleImagePointerDown}
                                            >
                                                <span className="profile-page__crop-rect-hint">Move</span>
                                                <button
                                                    className="profile-page__crop-handle"
                                                    type="button"
                                                    aria-label="Resize crop"
                                                    onPointerDown={handleResizePointerDown}
                                                >
                                                    <span className="material-symbols-rounded" aria-hidden="true">open_in_full</span>
                                                </button>
                                            </div>
                                        )}
                                    </div>
                                ) : (
                                    <div className="profile-page__crop-placeholder">
                                        <span className="material-symbols-rounded profile-page__crop-placeholder-icon" aria-hidden="true">crop_free</span>
                                        <p className="profile-page__crop-placeholder-title">Select an image to enable cropping</p>
                                        <p className="profile-page__crop-placeholder-text">
                                            After that you can drag the square area and save the cropped avatar.
                                        </p>
                                    </div>
                                )}
                            </div>
                        </div>

                        <div className="profile-page__feedback">
                            {error && <p className="profile-page__feedback-text profile-page__feedback-text--error">{error}</p>}
                            {status && <p className="profile-page__feedback-text profile-page__feedback-text--success">{status}</p>}
                        </div>

                        <FormSubmitButton disabled={!canSave}>
                            {isSaving ? "Saving profile..." : "Save profile"}
                        </FormSubmitButton>
                    </form>
                </div>
            </div>
        </section>
    );
}

type AvatarCircleProps = {
    avatarUrl: string | null;
    name: string;
    size: number;
};

function AvatarCircle({
    avatarUrl,
    name,
    size,
}: AvatarCircleProps) {
    const letter = name.trim().slice(0, 1).toUpperCase() || "U";

    return (
        <div className="profile-page__avatar" style={{ width: size, height: size, borderRadius: size / 2 }}>
            {avatarUrl ? (
                <img className="profile-page__avatar-image" src={avatarUrl} alt="" aria-hidden="true" draggable={false} />
            ) : (
                <span className="profile-page__avatar-placeholder">{letter}</span>
            )}
        </div>
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
    const maxSize = Math.min(geometry.imageHeight / geometry.imageWidth, 0.84);
    const size = clampNumber(defaultCropSizeRatio, minimumCropSizeRatio, maxSize);
    const maxY = Math.max(0, geometry.imageHeight / geometry.imageWidth - size);

    return {
        x: (1 - size) / 2,
        y: maxY / 2,
        size,
    };
}

function clampCropRect(cropRect: CropRect, geometry: ImageGeometry) {
    const maxSize = Math.min(geometry.imageHeight / geometry.imageWidth, 0.84);
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

async function buildCroppedAvatarPreviewUrl(sourceImage: SourceImage, geometry: ImageGeometry, cropRect: CropRect) {
    const image = new Image();
    image.src = sourceImage.url;
    await image.decode();

    const canvas = document.createElement("canvas");
    const outputSize = 256;
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
                reject(new Error("could not export cropped preview"));
                return;
            }

            resolve(nextBlob);
        }, "image/png");
    });

    return URL.createObjectURL(blob);
}

async function uploadAvatar(file: File) {
    const formData = new FormData();
    formData.append("file", file, file.name || "avatar.png");

    const response = await fetch("/api/v1/attachments/images", {
        method: "POST",
        body: formData,
    });

    if (!response.ok) {
        throw new Error(`upload failed with status ${response.status}`);
    }

    return response.json() as Promise<string>;
}

function buildAvatarFileName() {
    return `profile-avatar-${Date.now()}.png`;
}

function clampNumber(value: number, min: number, max: number) {
    if (max < min) {
        return max;
    }

    return Math.min(max, Math.max(min, value));
}

function hasDraggedFiles(event: DragEvent<Element>) {
    return Array.from(event.dataTransfer.types).includes("Files");
}

function extractPastedImageFile(dataTransfer: DataTransfer | null) {
    if (!dataTransfer) {
        return null;
    }

    const file = Array.from(dataTransfer.files ?? []).find(isImageFile);
    if (file) {
        return file;
    }

    return Array.from(dataTransfer.items ?? [])
        .map((item) => item.getAsFile())
        .find((candidate) => candidate !== null && isImageFile(candidate)) ?? null;
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
