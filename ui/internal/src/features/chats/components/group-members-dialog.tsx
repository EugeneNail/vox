import { type RefObject } from "react";
import { type PublicProfile } from "../../../profiles/profile-cache";
import { ModalBackdrop, UserAvatar, getProfileAvatarUrl } from "./chats-ui";

type GroupMembersDialogProps = {
    actionText: string;
    busyText: string;
    emptyStateText: string;
    error: string | null;
    inputName: string;
    inputPlaceholder: string;
    invitees: PublicProfile[];
    isBusy: boolean;
    isFocused: boolean;
    isOpen: boolean;
    isSearching: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void>;
    onFocusChange: (value: boolean) => void;
    onQueryChange: (value: string) => void;
    onToggleInvitee: (profile: PublicProfile) => void;
    query: string;
    queryInputRef: RefObject<HTMLInputElement | null>;
    queryTrimmed: string;
    secondaryText: string;
    selectedUserUuids: Set<string>;
    title: string;
    titleEyebrow: string;
    visibleProfiles: PublicProfile[];
    name?: string;
    onNameChange?: (value: string) => void;
};

export function GroupMembersDialog(props: GroupMembersDialogProps) {
    const {
        actionText,
        busyText,
        emptyStateText,
        error,
        inputName,
        inputPlaceholder,
        invitees,
        isBusy,
        isFocused,
        isOpen,
        isSearching,
        name,
        onClose,
        onConfirm,
        onFocusChange,
        onNameChange,
        onQueryChange,
        onToggleInvitee,
        query,
        queryInputRef,
        queryTrimmed,
        secondaryText,
        selectedUserUuids,
        title,
        titleEyebrow,
        visibleProfiles,
    } = props;

    if (!isOpen) {
        return null;
    }

    return (
        <ModalBackdrop onClick={onClose}>
            <section
                className="chats-me-page__group-dialog"
                role="dialog"
                aria-modal="true"
                aria-labelledby={`${inputName}-title`}
                onClick={(event) => event.stopPropagation()}
            >
                <div className="chats-me-page__group-dialog-header">
                    <div>
                        <p className="chats-me-page__eyebrow">{titleEyebrow}</p>
                        <h2 className="chats-me-page__group-dialog-title" id={`${inputName}-title`}>{title}</h2>
                    </div>
                    <p className="chats-me-page__group-dialog-text">
                        {secondaryText}
                    </p>
                </div>

                {typeof name === "string" && onNameChange && (
                    <label className="chats-me-page__group-field">
                        <span className="chats-me-page__group-field-label">Chat name</span>
                        <input
                            className="chats-me-page__group-field-input"
                            type="text"
                            name="group-name"
                            placeholder="Project team"
                            value={name}
                            onChange={(event) => onNameChange(event.target.value)}
                        />
                    </label>
                )}

                <label className="chats-me-page__group-search">
                    <span className="material-symbols-rounded chats-me-page__group-search-icon" aria-hidden="true">search</span>
                    <input
                        className="chats-me-page__group-search-input"
                        type="text"
                        name={inputName}
                        placeholder={inputPlaceholder}
                        ref={queryInputRef}
                        value={query}
                        onChange={(event) => onQueryChange(event.target.value)}
                        onFocus={() => onFocusChange(true)}
                        onBlur={() => onFocusChange(false)}
                    />
                    {(isFocused || query.length > 0) && (
                        <button
                            className="chats-me-page__group-search-clear"
                            type="button"
                            onMouseDown={(event) => event.preventDefault()}
                            onClick={(event) => {
                                event.preventDefault();
                                event.stopPropagation();
                                onQueryChange("");
                            }}
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">close</span>
                        </button>
                    )}
                </label>

                <div className="chats-me-page__group-dialog-content">
                    <div className="chats-me-page__group-results">
                        {queryTrimmed.length === 0 && invitees.length === 0 && (
                            <p className="chats-me-page__state">Type a name.</p>
                        )}
                        {isSearching && <p className="chats-me-page__state">Searching users...</p>}
                        {error && <p className="chats-me-page__state chats-me-page__state--error">{error}</p>}
                        {!isSearching && !error && queryTrimmed.length > 0 && visibleProfiles.length === 0 && (
                            <p className="chats-me-page__state">{emptyStateText}</p>
                        )}
                        {visibleProfiles.map((profile) => {
                            const isSelected = selectedUserUuids.has(profile.userUuid);

                            return (
                                <label
                                    className={
                                        [
                                            "chats-me-page__chat-button",
                                            "chats-me-page__chat-button--search-result",
                                            isSelected ? "chats-me-page__chat-button--selected" : "",
                                        ].filter(Boolean).join(" ")
                                    }
                                    key={profile.userUuid}
                                >
                                    <input
                                        checked={isSelected}
                                        className="chats-me-page__group-checkbox"
                                        disabled={isBusy}
                                        type="checkbox"
                                        onChange={() => onToggleInvitee(profile)}
                                    />
                                    <UserAvatar
                                        className="chats-me-page__avatar"
                                        src={getProfileAvatarUrl(profile)}
                                        label={profile.name}
                                    />
                                    <span className="chats-me-page__chat-preview">
                                        <span className="chats-me-page__chat-name">{profile.name}</span>
                                    </span>
                                </label>
                            );
                        })}
                    </div>
                </div>

                <div className="chats-me-page__group-dialog-actions">
                    <button className="chats-me-page__group-dialog-button" type="button" disabled={isBusy} onClick={onClose}>
                        Cancel
                    </button>
                    <button className="chats-me-page__group-dialog-button chats-me-page__group-dialog-button--primary" type="button" disabled={isBusy || (typeof name !== "string" && invitees.length === 0)} onClick={() => void onConfirm()}>
                        {isBusy ? busyText : actionText}
                    </button>
                </div>
            </section>
        </ModalBackdrop>
    );
}
