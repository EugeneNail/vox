import { AxiosError } from "axios";
import { ChangeEvent, FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { getApiMessage, getApiViolations } from "../../api/getApiViolations";
import { storeAuthTokens } from "../../auth/authTokens";
import { useApiClient } from "../../hooks/useApiClient";
import "./LoginPage.sass";

type LoginForm = {
  email: string;
  password: string;
};

type LoginViolations = Partial<Record<keyof LoginForm, string>>;

type AuthenticateResponse = {
  loginToken: string;
  refreshToken: string;
};

const initialForm: LoginForm = {
  email: "",
  password: "",
};

export default function LoginPage() {
  const apiClient = useApiClient();
  const [form, setForm] = useState<LoginForm>(initialForm);
  const [violations, setViolations] = useState<LoginViolations>({});
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
      const { data } = await apiClient.post<AuthenticateResponse>(
        "/api/v1/auth/users/authenticate",
        form,
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
            password: responseMessage,
          });
        }

        return;
      }

      setViolations({
        email: "Unexpected error. Try again.",
        password: "Unexpected error. Try again.",
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="login-page">
      <div className="login-page__panel">
        <p className="login-page__eyebrow">Authentication</p>
        <h1 className="login-page__title">Sign in to Vox.</h1>
        <p className="login-page__text">Welcome back. Pick up the conversation in seconds.</p>
        <form className="login-page__form" onSubmit={handleSubmit}>
          <label className="login-page__field">
            <span className="login-page__label">Email</span>
            <input
              className="login-page__input"
              name="email"
              type="email"
              autoComplete="email"
              placeholder="name@example.com"
              value={form.email}
              onChange={handleInputChange}
            />
            {violations.email ? <span className="login-page__error">{violations.email}</span> : null}
          </label>
          <label className="login-page__field">
            <span className="login-page__label">Password</span>
            <input
              className="login-page__input"
              name="password"
              type="password"
              autoComplete="current-password"
              placeholder="Enter your password"
              value={form.password}
              onChange={handleInputChange}
            />
            {violations.password ? (
              <span className="login-page__error">{violations.password}</span>
            ) : null}
          </label>
          <button className="login-page__button" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Signing in..." : "Sign in"}
          </button>
        </form>
        <p className="login-page__switch">
          New here?{" "}
          <Link className="login-page__switch-link" to="/signup">
            Create an account
          </Link>
        </p>
      </div>
    </section>
  );
}
