import { ReactNode } from "react";
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
        <div className="auth-form-card">
            <p className="auth-form-card__eyebrow">{eyebrow}</p>
            <h1 className="auth-form-card__title">{title}</h1>
            <p className="auth-form-card__text">{text}</p>
            {children}
        </div>
    );
}
