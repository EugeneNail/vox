import { Link, Route, Routes } from "react-router-dom";

function HomePage() {
  return (
    <section className="layout__panel">
      <p className="layout__eyebrow">Vox UI</p>
      <h1 className="layout__title">Frontend shell is online.</h1>
      <p className="layout__text">
        This Vite + React + TypeScript microservice is ready for future screens and API wiring.
      </p>
    </section>
  );
}

function NotFoundPage() {
  return (
    <section className="layout__panel">
      <p className="layout__eyebrow">404</p>
      <h1 className="layout__title">Page not found.</h1>
      <p className="layout__text">The requested UI route does not exist yet.</p>
    </section>
  );
}

export default function App() {
  return (
    <main className="layout">
      <header className="layout__header">
        <Link className="layout__brand" to="/">
          vox/ui
        </Link>
      </header>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </main>
  );
}
