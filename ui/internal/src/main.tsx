import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { MessageWebSocketProvider } from "./contexts/message-web-socket-context/message-web-socket-context";
import DesktopLayout from "./layouts/desktop-layout/desktop-layout";
import GuestLayout from "./layouts/guest-layout/guest-layout";
import ChatsMePage from "./pages/chats-me-page/chats-me-page";
import HomePage from "./pages/home-page/home-page";
import LoginPage from "./pages/login-page/login-page";
import ProfilePage from "./pages/profile-page/profile-page";
import NotFoundPage from "./pages/not-found-page/not-found-page";
import SignupPage from "./pages/signup-page/signup-page";
import "./main.sass";

ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
        <MessageWebSocketProvider>
            <BrowserRouter>
                <Routes>
                    <Route path="/" element={<HomePage />} />
                    <Route element={<GuestLayout />}>
                        <Route path="/login" element={<LoginPage />} />
                        <Route path="/signup" element={<SignupPage />} />
                    </Route>
                    <Route element={<DesktopLayout />}>
                        <Route path="/chats/@me" element={<ChatsMePage />} />
                        <Route path="/chats/@me/:chatUuid" element={<ChatsMePage />} />
                        <Route path="/profile" element={<ProfilePage />} />
                    </Route>
                    <Route path="*" element={<NotFoundPage />} />
                </Routes>
            </BrowserRouter>
        </MessageWebSocketProvider>
    </React.StrictMode>,
);
