import { Link, Outlet } from "react-router-dom";
import "./GuestLayout.sass";

export default function GuestLayout() {
  return (
    <main className="guest-layout">
      <header className="guest-layout__header">
        <Link className="guest-layout__brand" to="/">
          vox/ui
        </Link>
      </header>
      <div className="guest-layout__body">
        <Outlet />
      </div>
    </main>
  );
}
