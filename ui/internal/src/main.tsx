import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import DesktopLayout from "./layouts/desktop-layout/desktop-layout";
import GuestLayout from "./layouts/guest-layout/guest-layout";
import ChatsMePage from "./pages/chats-me-page/chats-me-page";
import HomePage from "./pages/home-page/home-page";
import LoginPage from "./pages/login-page/login-page";
import NotFoundPage from "./pages/not-found-page/not-found-page";
import SignupPage from "./pages/signup-page/signup-page";
import "./main.sass";

ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<HomePage />} />
                <Route element={<GuestLayout />}>
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/signup" element={<SignupPage />} />
                </Route>
                <Route element={<DesktopLayout />}>
                    <Route path="/chats/@me" element={<ChatsMePage />} />
                </Route>
                <Route path="*" element={<NotFoundPage />} />
            </Routes>
        </BrowserRouter>
    </React.StrictMode>,
);
