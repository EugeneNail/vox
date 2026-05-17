import { type MouseEvent, type RefObject } from "react";
import { ChatListLink, SearchResultButton, type Chat } from "./chats-ui";
import { type PublicProfile } from "../../../profiles/profile-cache";

type ChatsSidebarProps = {
    authenticatedUserUuid: string | null;
    chatPreviewByChatUuid: Record<string, string>;
    chats: Chat[];
    clearSearch: () => void;
    createPrivateChat: (profile: PublicProfile) => Promise<void>;
    error: string | null;
    isCreatingChat: boolean;
    isLoading: boolean;
    isSearchActive: boolean;
    isSearchFocused: boolean;
    isSearchingProfiles: boolean;
    localLastSeenRevisionByChatUuid: Record<string, number>;
    openCreateGroupChatModal: () => void;
    profilesByUserUuid: Record<string, PublicProfile>;
    searchInputRef: RefObject<HTMLInputElement | null>;
    searchProfiles: PublicProfile[];
    searchProfilesError: string | null;
    searchQuery: string;
    startResize: (event: MouseEvent<HTMLButtonElement>) => void;
    setIsSearchFocused: (value: boolean) => void;
    setSearchQuery: (value: string) => void;
};

export function ChatsSidebar(props: ChatsSidebarProps) {
    const {
        authenticatedUserUuid,
        chatPreviewByChatUuid,
        chats,
        clearSearch,
        createPrivateChat,
        error,
        isCreatingChat,
        isLoading,
        isSearchActive,
        isSearchingProfiles,
        localLastSeenRevisionByChatUuid,
        openCreateGroupChatModal,
        profilesByUserUuid,
        searchInputRef,
        searchProfiles,
        searchProfilesError,
        searchQuery,
        startResize,
        setIsSearchFocused,
        setSearchQuery,
    } = props;

    return (
        <aside className="chats-me-page__sidebar" aria-label="Chats">
            <div className="chats-me-page__sidebar-header">
                <div className="chats-me-page__sidebar-title-row">
                    <h1 className="chats-me-page__title">Chats</h1>
                    <button className="chats-me-page__create-group-button" type="button" onClick={openCreateGroupChatModal}>
                        <span className="material-symbols-rounded chats-me-page__create-group-button-icon" aria-hidden="true">group_add</span>
                        Group
                    </button>
                </div>
                <label className="chats-me-page__search">
                    <span className="material-symbols-rounded chats-me-page__search-icon" aria-hidden="true">search</span>
                    <input
                        className="chats-me-page__search-input"
                        type="text"
                        name="search"
                        placeholder="Search users"
                        ref={searchInputRef}
                        value={searchQuery}
                        onChange={(event) => setSearchQuery(event.target.value)}
                        onFocus={() => setIsSearchFocused(true)}
                        onBlur={() => setIsSearchFocused(false)}
                    />
                    {isSearchActive && (
                        <button
                            className="chats-me-page__search-clear"
                            type="button"
                            onMouseDown={(event) => event.preventDefault()}
                            onClick={(event) => {
                                event.preventDefault();
                                event.stopPropagation();
                                clearSearch();
                            }}
                        >
                            <span className="material-symbols-rounded" aria-hidden="true">close</span>
                        </button>
                    )}
                </label>
            </div>

            <div className="chats-me-page__list">
                {isSearchActive ? (
                    <>
                        {searchQuery.trim().length === 0 && (
                            <p className="chats-me-page__state">Type a name.</p>
                        )}
                        {isSearchingProfiles && <p className="chats-me-page__state">Searching users...</p>}
                        {searchProfilesError && <p className="chats-me-page__state chats-me-page__state--error">{searchProfilesError}</p>}
                        {!isSearchingProfiles && !searchProfilesError && searchQuery.trim().length > 0 && searchProfiles.length === 0 && (
                            <p className="chats-me-page__state">No users found.</p>
                        )}
                        {searchProfiles.map((profile) => (
                            <SearchResultButton
                                disabled={isCreatingChat}
                                key={profile.userUuid}
                                profile={profile}
                                onClick={() => void createPrivateChat(profile)}
                            />
                        ))}
                    </>
                ) : (
                    <>
                        {isLoading && <p className="chats-me-page__state">Loading chats...</p>}
                        {error && <p className="chats-me-page__state chats-me-page__state--error">{error}</p>}
                        {!isLoading && !error && chats.length === 0 && (
                            <p className="chats-me-page__state">No chats yet.</p>
                        )}
                        {chats.map((chat) => (
                            <ChatListLink
                                authenticatedUserUuid={authenticatedUserUuid}
                                chat={chat}
                                chatPreview={chatPreviewByChatUuid[chat.uuid]}
                                key={chat.uuid}
                                localLastSeenRevisionByChatUuid={localLastSeenRevisionByChatUuid}
                                profilesByUserUuid={profilesByUserUuid}
                            />
                        ))}
                    </>
                )}
            </div>
            <button
                aria-label="Resize chats sidebar"
                className="chats-me-page__sidebar-resize-handle"
                type="button"
                onMouseDown={startResize}
            />
        </aside>
    );
}
