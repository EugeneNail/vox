import { Link } from "react-router-dom";
import "./NotFoundPage.sass";

export default function NotFoundPage() {
  return (
    <main className="not-found-page">
      <section className="not-found-page__panel">
        <p className="not-found-page__eyebrow">404</p>
        <h1 className="not-found-page__title">Page not found.</h1>
        <p className="not-found-page__text">The requested UI route does not exist yet.</p>
        <div className="not-found-page__actions">
          <Link className="not-found-page__button" to="/">
            Back to home
          </Link>
        </div>
      </section>
    </main>
  );
}
