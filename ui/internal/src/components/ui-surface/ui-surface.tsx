import { ReactNode } from "react";
import "./ui-surface.sass";

type UiSurfaceProps = {
    children: ReactNode;
    className?: string;
    tone?: "default" | "muted" | "accent";
};

export default function UiSurface({ children, className = "", tone = "default" }: UiSurfaceProps) {
    const classes = [
        "ui-surface",
        `ui-surface--${tone}`,
        className,
    ].filter(Boolean).join(" ");

    return (
        <div className={classes}>
            {children}
        </div>
    );
}
