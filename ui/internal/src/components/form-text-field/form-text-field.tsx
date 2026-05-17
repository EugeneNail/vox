import UiField from "../ui-field/ui-field";
import "./form-text-field.sass";

type FormTextFieldProps = {
    label: string;
    name: string;
    type: string;
    autoComplete: string;
    placeholder: string;
    value: string;
    error?: string;
    onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
};

export default function FormTextField({
    label,
    name,
    type,
    autoComplete,
    placeholder,
    value,
    error,
    onChange,
}: FormTextFieldProps) {
    return (
        <UiField
            autoComplete={autoComplete}
            className="form-text-field__input"
            error={error}
            label={label}
            name={name}
            placeholder={placeholder}
            type={type}
            value={value}
            onChange={onChange}
        />
    );
}
