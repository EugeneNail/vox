import { ReactNode } from "react";
import UiSurface from "../ui-surface/ui-surface";
import "./auth-form-card.sass";

type AuthFormCardProps = {
    eyebrow: string;
    title: string;
    text: string;
    children: ReactNode;
};

export default function AuthFormCard({
    eyebrow,
    title,
    text,
    children,
}: AuthFormCardProps) {
    return (
        <UiSurface className="auth-form-card" tone="accent">
            <p className="auth-form-card__eyebrow">{eyebrow}</p>
            <h1 className="auth-form-card__title">{title}</h1>
            <p className="auth-form-card__text">{text}</p>
            {children}
        </UiSurface>
    );
}
