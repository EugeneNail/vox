import { ChangeEvent, FormEvent, KeyboardEvent, useRef, useState } from "react";
import "./message-composer.sass";

type MessageComposerProps = {
    disabled?: boolean;
    onSubmit: (text: string) => Promise<void> | void;
};

const maxMessageLength = 2000;

export default function MessageComposer({ disabled = false, onSubmit }: MessageComposerProps) {
    const textareaRef = useRef<HTMLTextAreaElement | null>(null);
    const [text, setText] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    function handleTextChange(event: ChangeEvent<HTMLTextAreaElement>) {
        setError(null);
        setText(event.target.value);
        resizeTextarea(event.target);
    }

    async function handleSubmit(event?: FormEvent<HTMLFormElement>) {
        event?.preventDefault();

        const message = text.trim();
        if (!message || disabled || isSubmitting) {
            return;
        }

        setIsSubmitting(true);
        try {
            await onSubmit(message);
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
        if (event.key === "Enter" && !event.shiftKey) {
            event.preventDefault();
            void handleSubmit();
        }
    }

    function resetTextarea() {
        if (!textareaRef.current) {
            return;
        }

        textareaRef.current.style.height = "";
    }

    return (
        <form className="message-composer" onSubmit={handleSubmit}>
            <button className="message-composer__attach-button" type="button" aria-label="Attach image" disabled={disabled}>
                <span className="material-symbols-rounded" aria-hidden="true">add</span>
            </button>

            <div className="message-composer__field">
                <div className="message-composer__textarea-shell">
                    <textarea
                        ref={textareaRef}
                        className="message-composer__textarea"
                        placeholder="Write a message"
                    rows={1}
                    value={text}
                    disabled={disabled || isSubmitting}
                    maxLength={maxMessageLength}
                    onChange={handleTextChange}
                    onKeyDown={handleKeyDown}
                />
                </div>
                {text.length >= 500 && (
                    <span className={text.length >= 1900 ? "message-composer__character-counter message-composer__character-counter--danger" : "message-composer__character-counter"}>
                        {text.length} / {maxMessageLength}
                    </span>
                )}
                {error && <span className="message-composer__error">{error}</span>}
            </div>

            <button
                className="message-composer__send-button"
                type="submit"
                aria-label="Send message"
                disabled={disabled || isSubmitting || Boolean(error) || !text.trim()}
            >
                <span className="material-symbols-rounded" aria-hidden="true">send</span>
            </button>
        </form>
    );

}

function resizeTextarea(textarea: HTMLTextAreaElement) {
    textarea.style.height = "auto";
    textarea.style.height = `${textarea.scrollHeight}px`;
}
