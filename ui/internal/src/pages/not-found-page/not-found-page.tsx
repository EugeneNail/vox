import { Link } from "react-router-dom";
import UiButton from "../../components/ui-button/ui-button";
import UiSurface from "../../components/ui-surface/ui-surface";
import "./not-found-page.sass";

export default function NotFoundPage() {
    return (
        <main className="not-found-page">
            <UiSurface className="not-found-page__panel" tone="accent">
                <p className="not-found-page__eyebrow">404</p>
                <h1 className="not-found-page__title">Page not found.</h1>
                <p className="not-found-page__text">The requested route is outside the current messaging shell.</p>
                <div className="not-found-page__actions">
                    <Link to="/">
                        <UiButton>Back to home</UiButton>
                    </Link>
                </div>
            </UiSurface>
        </main>
    );
}
