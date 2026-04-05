import { AxiosError } from "axios";
import { ChangeEvent, FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { getApiMessage, getApiViolations } from "../../api/get-api-violations";
import { storeAuthTokens } from "../../auth/auth-tokens";
import AuthFormCard from "../../components/auth-form-card/auth-form-card";
import FormSubmitButton from "../../components/form-submit-button/form-submit-button";
import FormTextField from "../../components/form-text-field/form-text-field";
import { useApiClient } from "../../hooks/use-api-client";
import "./signup-page.sass";

type SignupForm = {
  email: string;
  password: string;
  passwordConfirmation: string;
};

type SignupViolations = Partial<Record<keyof SignupForm, string>>;

type AuthenticateResponse = {
  loginToken: string;
  refreshToken: string;
};

const initialForm: SignupForm = {
  email: "",
  password: "",
  passwordConfirmation: "",
};

export default function SignupPage() {
    const apiClient = useApiClient();
    const [form, setForm] = useState<SignupForm>(initialForm);
    const [violations, setViolations] = useState<SignupViolations>({});
    const [isSubmitting, setIsSubmitting] = useState(false);

    function handleInputChange(event: ChangeEvent<HTMLInputElement>) {
        const { name, value } = event.target;

        setForm((currentForm) => ({
            ...currentForm,
            [name]: value,
        }));

        setViolations((currentViolations) => ({
            ...currentViolations,
            [name]: undefined,
        }));
    }

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        setIsSubmitting(true);
        setViolations({});

        try {
            await apiClient.post("/api/v1/auth/users", form);

            const { data } = await apiClient.post<AuthenticateResponse>(
                "/api/v1/auth/users/authenticate",
                {
                    email: form.email,
                    password: form.password,
                },
            );

            storeAuthTokens(data);
        } catch (error) {
            const nextViolations = getApiViolations(error);
            if (nextViolations) {
                setViolations(nextViolations);
                return;
            }

            if (error instanceof AxiosError) {
                const responseMessage = getApiMessage(error);
                if (responseMessage) {
                    setViolations({
                        email: responseMessage,
                    });
                }

                return;
            }

            setViolations({
                email: "Unexpected error. Try again.",
            });
        } finally {
            setIsSubmitting(false);
        }
    }

    return (
        <section className="signup-page">
            <AuthFormCard
                eyebrow="Registration"
                title="Create a Vox account."
                text="Join once and step straight into your next server."
            >
                <form className="signup-page__form" onSubmit={handleSubmit}>
                    <FormTextField
                        label="Email"
                        name="email"
                        type="email"
                        autoComplete="email"
                        placeholder="name@example.com"
                        value={form.email}
                        error={violations.email}
                        onChange={handleInputChange}
                    />
                    <FormTextField
                        label="Password"
                        name="password"
                        type="password"
                        autoComplete="new-password"
                        placeholder="Create a strong password"
                        value={form.password}
                        error={violations.password}
                        onChange={handleInputChange}
                    />
                    <FormTextField
                        label="Confirm password"
                        name="passwordConfirmation"
                        type="password"
                        autoComplete="new-password"
                        placeholder="Repeat the password"
                        value={form.passwordConfirmation}
                        error={violations.passwordConfirmation}
                        onChange={handleInputChange}
                    />
                    <FormSubmitButton disabled={isSubmitting}>
                        {isSubmitting ? "Creating account..." : "Create account"}
                    </FormSubmitButton>
                </form>
                <p className="signup-page__switch">
                    Already have an account?{" "}
                    <Link className="signup-page__switch-link" to="/login">
                        Sign in
                    </Link>
                </p>
            </AuthFormCard>
        </section>
    );
}
