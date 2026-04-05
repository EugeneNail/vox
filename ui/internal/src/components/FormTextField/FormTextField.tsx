import "./FormTextField.sass";

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
        <label className="form-text-field">
            <span className="form-text-field__label">{label}</span>
            <input
                className="form-text-field__input"
                name={name}
                type={type}
                autoComplete={autoComplete}
                placeholder={placeholder}
                value={value}
                onChange={onChange}
            />
            {error ? <span className="form-text-field__error">{error}</span> : null}
        </label>
    );
}
