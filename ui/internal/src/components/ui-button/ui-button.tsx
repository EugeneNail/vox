import { ButtonHTMLAttributes, ReactNode } from "react";
import "./ui-button.sass";

type UiButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
    children: ReactNode;
    variant?: "primary" | "secondary" | "ghost" | "danger";
};

export default function UiButton({ children, className = "", variant = "primary", type = "button", ...props }: UiButtonProps) {
    const classes = [
        "ui-button",
        `ui-button--${variant}`,
        className,
    ].filter(Boolean).join(" ");

    return (
        <button className={classes} type={type} {...props}>
            {children}
        </button>
    );
}
