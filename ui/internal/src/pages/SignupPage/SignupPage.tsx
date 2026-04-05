import { AxiosError } from "axios";
import { ChangeEvent, FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { getApiMessage, getApiViolations } from "../../api/getApiViolations";
import { storeAuthTokens } from "../../auth/authTokens";
import { useApiClient } from "../../hooks/useApiClient";
import "./SignupPage.sass";

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
      <div className="signup-page__panel">
        <p className="signup-page__eyebrow">Registration</p>
        <h1 className="signup-page__title">Create a Vox account.</h1>
        <p className="signup-page__text">Join once and step straight into your next server.</p>
        <form className="signup-page__form" onSubmit={handleSubmit}>
          <label className="signup-page__field">
            <span className="signup-page__label">Email</span>
            <input
              className="signup-page__input"
              name="email"
              type="email"
              autoComplete="email"
              placeholder="name@example.com"
              value={form.email}
              onChange={handleInputChange}
            />
            {violations.email ? (
              <span className="signup-page__error">{violations.email}</span>
            ) : null}
          </label>
          <label className="signup-page__field">
            <span className="signup-page__label">Password</span>
            <input
              className="signup-page__input"
              name="password"
              type="password"
              autoComplete="new-password"
              placeholder="Create a strong password"
              value={form.password}
              onChange={handleInputChange}
            />
            {violations.password ? (
              <span className="signup-page__error">{violations.password}</span>
            ) : null}
          </label>
          <label className="signup-page__field">
            <span className="signup-page__label">Confirm password</span>
            <input
              className="signup-page__input"
              name="passwordConfirmation"
              type="password"
              autoComplete="new-password"
              placeholder="Repeat the password"
              value={form.passwordConfirmation}
              onChange={handleInputChange}
            />
            {violations.passwordConfirmation ? (
              <span className="signup-page__error">{violations.passwordConfirmation}</span>
            ) : null}
          </label>
          <button className="signup-page__button" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Creating account..." : "Create account"}
          </button>
        </form>
        <p className="signup-page__switch">
          Already have an account?{" "}
          <Link className="signup-page__switch-link" to="/login">
            Sign in
          </Link>
        </p>
      </div>
    </section>
  );
}
