import { AxiosError } from "axios";
import { useState } from "react";
import { useApiClient } from "../../hooks/useApiClient";
import "./LoginPage.sass";

export default function LoginPage() {
  const apiClient = useApiClient();
  const [statusText, setStatusText] = useState("Request has not been sent yet.");

  async function handleDefaultLogin() {
    try {
      const response = await apiClient.post("/api/v1/auth/users/authenticate", {
        email: "test@example.com",
        password: "Strong123!",
      });

      setStatusText(`Request completed successfully with status ${response.status}.`);
    } catch (error) {
      if (error instanceof AxiosError) {
        setStatusText(`Request failed with status ${error.response?.status ?? "unknown"}.`);
        return;
      }

      setStatusText("Request failed with an unexpected error.");
    }
  }

  return (
    <section className="login-page">
      <p className="login-page__eyebrow">Login</p>
      <h1 className="login-page__title">Login page rendered successfully.</h1>
      <p className="login-page__text">
        Guest layout is active and ready for future authentication forms.
      </p>
      <div className="login-page__actions">
        <button className="login-page__button" type="button" onClick={handleDefaultLogin}>
          Send default login request
        </button>
      </div>
      <p className="login-page__status">{statusText}</p>
    </section>
  );
}
