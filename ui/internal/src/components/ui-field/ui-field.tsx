import { InputHTMLAttributes } from "react";
import "./ui-field.sass";

type UiFieldProps = InputHTMLAttributes<HTMLInputElement> & {
    label: string;
    error?: string;
};

export default function UiField({ label, error, className = "", ...props }: UiFieldProps) {
    const inputClasses = [
        "ui-field__input",
        error ? "ui-field__input--error" : "",
        className,
    ].filter(Boolean).join(" ");

    return (
        <label className="ui-field">
            <span className="ui-field__label">{label}</span>
            <input className={inputClasses} {...props} />
            {error ? <span className="ui-field__error">{error}</span> : null}
        </label>
    );
}
