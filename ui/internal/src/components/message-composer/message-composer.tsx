import { ChangeEvent, ClipboardEvent, DragEvent, FormEvent, KeyboardEvent, useEffect, useLayoutEffect, useRef, useState } from "react";

import { buildAttachmentUrl, generateLocalAttachmentUuid, isImageAttachmentName, isImageFile, MessageAttachment } from "../../messages/message-attachments";
import "./message-composer.sass";

type MessageComposerProps = {
    disabled?: boolean;
    editingText?: string | null;
    editingAttachments?: MessageAttachment[];
    onCancelEdit?: () => void;
    onSubmit: (text: string, attachments: string[]) => Promise<void> | void;
};

type ComposerAttachmentStatus = "uploading" | "ready" | "error";

type ComposerAttachment = {
    id: string;
    fileName: string;
    kind: "image" | "document";
    status: ComposerAttachmentStatus;
    serverName?: string;
    previewUrl?: string;
};

const maxMessageLength = 2000;
const spellCheckStorageKey = "vox.messageComposer.spellCheck";

export default function MessageComposer({ disabled = false, editingText = null, editingAttachments = [], onCancelEdit, onSubmit }: MessageComposerProps) {
    const textareaRef = useRef<HTMLTextAreaElement | null>(null);
    const fileInputRef = useRef<HTMLInputElement | null>(null);
    const attachmentsRef = useRef<ComposerAttachment[]>([]);
    const dragDepthRef = useRef(0);
    const shouldFocusAfterSubmitRef = useRef(false);
    const [text, setText] = useState("");
    const [attachments, setAttachments] = useState<ComposerAttachment[]>([]);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [isDragActive, setIsDragActive] = useState(false);
    const [isSpellCheckEnabled, setIsSpellCheckEnabled] = useState(() => localStorage.getItem(spellCheckStorageKey) === "enabled");
    const isEditing = editingText !== null;
    const hasBlockingAttachments = attachments.some((attachment) => attachment.status === "uploading" || attachment.status === "error");
    const readyAttachmentNames = attachments
        .filter((attachment) => attachment.status === "ready" && attachment.serverName)
        .map((attachment) => attachment.serverName as string);
    const attachmentStatusMessage = getAttachmentStatusMessage(attachments);
    const canSubmit = !disabled && !isSubmitting && !Boolean(error) && !hasBlockingAttachments && (Boolean(text.trim()) || readyAttachmentNames.length > 0);

    useEffect(() => {
        attachmentsRef.current = attachments;
    }, [attachments]);

    useEffect(() => {
        if (!isEditing) {
            clearAttachments();
            setText("");
            setError(null);
            return;
        }

        clearAttachments();
        setText(editingText ?? "");
        setError(null);
        setAttachments(editingAttachments.map((attachment) => ({
            id: attachment.uuid,
            fileName: attachment.name,
            kind: isImageAttachmentName(attachment.name) ? "image" : "document",
            status: "ready",
            serverName: attachment.name,
            previewUrl: isImageAttachmentName(attachment.name) ? buildAttachmentUrl(attachment.name) : undefined,
        })));
        textareaRef.current?.focus();
    }, [editingAttachments, editingText, isEditing]);

    useEffect(() => {
        return () => {
            revokeAttachmentPreviews(attachmentsRef.current);
        };
    }, []);

    useLayoutEffect(() => {
        if (!textareaRef.current) {
            return;
        }

        resizeTextarea(textareaRef.current);
    }, [text, attachments.length]);

    useLayoutEffect(() => {
        if (!shouldFocusAfterSubmitRef.current || disabled || isSubmitting) {
            return;
        }

        shouldFocusAfterSubmitRef.current = false;
        textareaRef.current?.focus({ preventScroll: true });
    }, [disabled, isSubmitting, text]);

    function handleTextChange(event: ChangeEvent<HTMLTextAreaElement>) {
        setError(null);
        setText(event.target.value);
    }

    async function handleSubmit(event?: FormEvent<HTMLFormElement>) {
        event?.preventDefault();

        const message = text.trim();
        if ((!message && readyAttachmentNames.length === 0) || disabled || isSubmitting) {
            return;
        }

        if (hasBlockingAttachments) {
            setError("Wait until file uploads finish.");
            return;
        }

        setIsSubmitting(true);
        try {
            await onSubmit(message, readyAttachmentNames);
            shouldFocusAfterSubmitRef.current = true;
            clearAttachments();
            setText("");
            resetTextarea();
            setError(null);
        } catch {
            setError("Message was not sent.");
        } finally {
            setIsSubmitting(false);
        }
    }

    function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
        if (event.key === "Escape" && isEditing) {
            event.preventDefault();
            handleCancelEdit();
            return;
        }

        if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault();
            void handleSubmit();
        }
    }

    function handleCancelEdit() {
        clearAttachments();
        setText("");
        setError(null);
        resetTextarea();
        onCancelEdit?.();
    }

    function handleSpellCheckToggle() {
        setIsSpellCheckEnabled((currentValue) => {
            const nextValue = !currentValue;
            localStorage.setItem(spellCheckStorageKey, nextValue ? "enabled" : "disabled");

            return nextValue;
        });
    }

    function handleAttachClick() {
        if (disabled || isSubmitting) {
            return;
        }

        fileInputRef.current?.click();
    }

    function handleFileInputChange(event: ChangeEvent<HTMLInputElement>) {
        void addFiles(Array.from(event.target.files ?? []));
        event.target.value = "";
    }

    async function handlePaste(event: ClipboardEvent<HTMLTextAreaElement>) {
        const files = Array.from(event.clipboardData.files ?? []);
        if (files.length === 0) {
            return;
        }

        event.preventDefault();
        await addFiles(files);
    }

    function handleDrop(event: DragEvent<HTMLFormElement>) {
        event.preventDefault();
        dragDepthRef.current = 0;
        setIsDragActive(false);

        const files = Array.from(event.dataTransfer.files ?? []);
        if (files.length === 0) {
            return;
        }

        void addFiles(files);
    }

    function handleDragEnter(event: DragEvent<HTMLFormElement>) {
        if (!hasDraggedFiles(event)) {
            return;
        }

        event.preventDefault();
        dragDepthRef.current += 1;
        setIsDragActive(true);
    }

    function handleDragOver(event: DragEvent<HTMLFormElement>) {
        if (!hasDraggedFiles(event)) {
            return;
        }

        event.preventDefault();
    }

    function handleDragLeave(event: DragEvent<HTMLFormElement>) {
        if (!hasDraggedFiles(event)) {
            return;
        }

        event.preventDefault();
        dragDepthRef.current = Math.max(0, dragDepthRef.current - 1);
        if (dragDepthRef.current === 0) {
            setIsDragActive(false);
        }
    }

    async function addFiles(files: File[]) {
        if (disabled || isSubmitting || files.length === 0) {
            return;
        }

        for (const file of files) {
            await addFile(file);
        }
    }

    async function addFile(file: File) {
        const id = generateLocalAttachmentUuid();
        const fileName = file.name.trim() || buildFallbackFileName(file);
        const kind = isImageFile(file) ? "image" : "document";
        const previewUrl = kind === "image" ? URL.createObjectURL(file) : undefined;

        setAttachments((currentAttachments) => [
            ...currentAttachments,
            {
                id,
                fileName,
                kind,
                status: "uploading",
                previewUrl,
            },
        ]);

        try {
            const serverName = await uploadAttachment(file);
            setAttachments((currentAttachments) => currentAttachments.map((attachment) => {
                if (attachment.id !== id) {
                    return attachment;
                }

                if (attachment.previewUrl?.startsWith("blob:")) {
                    URL.revokeObjectURL(attachment.previewUrl);
                }

                return {
                    ...attachment,
                    status: "ready",
                    serverName,
                    previewUrl: attachment.kind === "image" ? buildAttachmentUrl(serverName) : undefined,
                };
            }));
        } catch {
            setAttachments((currentAttachments) => currentAttachments.map((attachment) => (
                attachment.id === id
                    ? {
                        ...attachment,
                        status: "error",
                    }
                    : attachment
            )));
            setError(`Could not upload ${fileName}.`);
        }
    }

    function removeAttachment(id: string) {
        setAttachments((currentAttachments) => {
            const removedAttachment = currentAttachments.find((attachment) => attachment.id === id);
            if (removedAttachment?.previewUrl?.startsWith("blob:")) {
                URL.revokeObjectURL(removedAttachment.previewUrl);
            }

            return currentAttachments.filter((attachment) => attachment.id !== id);
        });

        setError(null);
    }

    function clearAttachments() {
        revokeAttachmentPreviews(attachmentsRef.current);
        setAttachments([]);
        dragDepthRef.current = 0;
        setIsDragActive(false);
    }

    function resetTextarea() {
        if (!textareaRef.current) {
            return;
        }

        textareaRef.current.style.height = "";
    }

    return (
        <form
            className={isDragActive ? "message-composer message-composer--drag-active" : "message-composer"}
            onSubmit={handleSubmit}
            onDragEnter={handleDragEnter}
            onDragLeave={handleDragLeave}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
        >
            <input
                ref={fileInputRef}
                className="message-composer__file-input"
                type="file"
                multiple
                onChange={handleFileInputChange}
            />

            {isEditing && (
                <div className="message-composer__editing-state">
                    <div className="message-composer__editing-copy">
                        <span className="message-composer__editing-label">Editing message</span>
                        <span className="message-composer__editing-preview">{editingText}</span>
                    </div>
                    <button className="message-composer__cancel-edit-button" type="button" aria-label="Cancel editing" onClick={handleCancelEdit}>
                        <span className="material-symbols-rounded" aria-hidden="true">close</span>
                    </button>
                </div>
            )}

            {isEditing && <span className="message-composer__editing-divider" aria-hidden="true" />}

            <button
                className="message-composer__attach-button"
                type="button"
                aria-label="Attach files"
                disabled={disabled || isSubmitting}
                onClick={handleAttachClick}
            >
                <span className="material-symbols-rounded" aria-hidden="true">add</span>
            </button>

            <div className="message-composer__field">
                <div className="message-composer__textarea-shell">
                    <textarea
                        ref={textareaRef}
                        className="message-composer__textarea"
                        placeholder={isEditing ? "Edit message" : "Write a message"}
                        rows={1}
                        value={text}
                        disabled={disabled || isSubmitting}
                        maxLength={maxMessageLength}
                        spellCheck={isSpellCheckEnabled}
                        onChange={handleTextChange}
                        onKeyDown={handleKeyDown}
                        onPaste={handlePaste}
                    />
                </div>

                {attachments.length > 0 && (
                    <div className="message-composer__attachments" aria-label="Selected files">
                        {attachments.map((attachment) => (
                            <ComposerAttachmentCard
                                attachment={attachment}
                                key={attachment.id}
                                onRemove={() => removeAttachment(attachment.id)}
                            />
                        ))}
                    </div>
                )}

                {attachmentStatusMessage && <span className="message-composer__attachment-note">{attachmentStatusMessage}</span>}

                {text.length >= 500 && (
                    <span className={text.length >= 1900 ? "message-composer__character-counter message-composer__character-counter--danger" : "message-composer__character-counter"}>
                        {text.length} / {maxMessageLength}
                    </span>
                )}
                {error && <span className="message-composer__error">{error}</span>}
            </div>

            <button
                className={
                    isSpellCheckEnabled
                        ? "message-composer__spellcheck-button message-composer__spellcheck-button--active"
                        : "message-composer__spellcheck-button"
                }
                type="button"
                aria-label={isSpellCheckEnabled ? "Disable spellcheck" : "Enable spellcheck"}
                aria-pressed={isSpellCheckEnabled}
                onClick={handleSpellCheckToggle}
            >
                <span className="material-symbols-rounded" aria-hidden="true">spellcheck</span>
            </button>

            <button
                className="message-composer__send-button"
                type="submit"
                aria-label={isEditing ? "Save message" : "Send message"}
                disabled={!canSubmit}
            >
                <span className="material-symbols-rounded" aria-hidden="true">{isEditing ? "check" : "send"}</span>
            </button>
        </form>
    );
}

function ComposerAttachmentCard({ attachment, onRemove }: { attachment: ComposerAttachment; onRemove: () => void; }) {
    return (
        <article className={attachment.status === "uploading" ? "message-composer__attachment-card message-composer__attachment-card--uploading" : "message-composer__attachment-card"}>
            {attachment.kind === "image" ? (
                <div className="message-composer__attachment-preview-shell">
                    <img className="message-composer__attachment-preview" src={attachment.previewUrl ?? buildAttachmentUrl(attachment.serverName ?? attachment.fileName)} alt={attachment.fileName} />
                    {attachment.status === "uploading" && (
                        <div className="message-composer__attachment-overlay">
                            <span className="material-symbols-rounded" aria-hidden="true">progress_activity</span>
                            <span>Uploading</span>
                        </div>
                    )}
                </div>
            ) : (
                <div className="message-composer__attachment-document">
                    <span className="material-symbols-rounded message-composer__attachment-document-icon" aria-hidden="true">description</span>
                    <span className="message-composer__attachment-document-name">{attachment.fileName}</span>
                </div>
            )}

            {attachment.status === "error" && (
                <span className="message-composer__attachment-error">Upload failed</span>
            )}

            <button className="message-composer__attachment-remove" type="button" aria-label={`Remove ${attachment.fileName}`} onClick={onRemove}>
                <span className="material-symbols-rounded" aria-hidden="true">close</span>
            </button>
        </article>
    );
}

function getAttachmentStatusMessage(attachments: ComposerAttachment[]) {
    if (attachments.some((attachment) => attachment.status === "uploading")) {
        return "Uploading selected files...";
    }

    if (attachments.some((attachment) => attachment.status === "error")) {
        return "Remove failed uploads before sending.";
    }

    return null;
}

async function uploadAttachment(file: File) {
    const formData = new FormData();
    formData.append("file", file, file.name || "attachment");

    const response = await fetch("/api/v1/attachments/images", {
        method: "POST",
        body: formData,
    });

    if (!response.ok) {
        throw new Error(`upload failed with status ${response.status}`);
    }

    return response.json() as Promise<string>;
}

function hasDraggedFiles(event: DragEvent<HTMLFormElement>) {
    return Array.from(event.dataTransfer.types).includes("Files");
}

function buildFallbackFileName(file: File) {
    const extension = file.type.startsWith("image/") ? getImageExtension(file.type) : "";
    return `file-${Date.now()}${extension}`;
}

function getImageExtension(contentType: string) {
    switch (contentType.trim().toLowerCase()) {
        case "image/gif":
            return ".gif";
        case "image/jpeg":
            return ".jpg";
        case "image/png":
            return ".png";
        case "image/webp":
            return ".webp";
        default:
            return "";
    }
}

function resizeTextarea(textarea: HTMLTextAreaElement) {
    textarea.style.height = "auto";
    textarea.style.height = `${textarea.scrollHeight}px`;
}

function revokeAttachmentPreviews(attachments: ComposerAttachment[]) {
    attachments.forEach((attachment) => {
        if (attachment.previewUrl?.startsWith("blob:")) {
            URL.revokeObjectURL(attachment.previewUrl);
        }
    });
}
