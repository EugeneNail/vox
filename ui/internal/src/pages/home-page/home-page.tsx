import { Link } from "react-router-dom";
import "./home-page.sass";

export default function HomePage() {
    return (
        <main className="home-page">
            <section className="home-page__panel">
                <p className="home-page__eyebrow">Vox Platform</p>
                <h1 className="home-page__title">Single entrypoint for voice, text and streaming.</h1>
                <p className="home-page__text">
                    UI microservice is mounted and ready for future product screens.
                </p>
                <div className="home-page__actions">
                    <Link className="home-page__button" to="/login">
                        Open login
                    </Link>
                    <Link className="home-page__button home-page__button--secondary" to="/signup">
                        Create account
                    </Link>
                </div>
            </section>
        </main>
    );
}
