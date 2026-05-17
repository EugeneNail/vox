import { ReactNode } from "react";
import UiButton from "../ui-button/ui-button";
import "./form-submit-button.sass";

type FormSubmitButtonProps = {
    disabled?: boolean;
    children: ReactNode;
};

export default function FormSubmitButton({ disabled, children }: FormSubmitButtonProps) {
    return (
        <UiButton className="form-submit-button" disabled={disabled} type="submit" variant="primary">
            {children}
        </UiButton>
    );
}
