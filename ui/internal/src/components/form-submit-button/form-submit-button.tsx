import { ReactNode } from "react";
import "./form-submit-button.sass";

type FormSubmitButtonProps = {
    disabled?: boolean;
    children: ReactNode;
};

export default function FormSubmitButton({ disabled, children }: FormSubmitButtonProps) {
    return (
        <button className="form-submit-button" type="submit" disabled={disabled}>
            {children}
        </button>
    );
}
