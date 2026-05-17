import { Link } from "react-router-dom";
import UiButton from "../../components/ui-button/ui-button";
import UiSurface from "../../components/ui-surface/ui-surface";
import "./home-page.sass";

export default function HomePage() {
    return (
        <main className="home-page">
            <UiSurface className="home-page__panel" tone="accent">
                <p className="home-page__eyebrow">Vox Platform</p>
                <h1 className="home-page__title">Messaging, presence and media in one Telegram-inspired workspace.</h1>
                <p className="home-page__text">
                    The backend contract remains stable while the frontend is rebuilt around cleaner state boundaries and a lighter product language.
                </p>
                <div className="home-page__actions">
                    <Link to="/login">
                        <UiButton>Open login</UiButton>
                    </Link>
                    <Link to="/signup">
                        <UiButton variant="secondary">Create account</UiButton>
                    </Link>
                </div>
            </UiSurface>
        </main>
    );
}
