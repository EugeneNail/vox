import { Outlet } from "react-router-dom";
import "./guest-layout.sass";

export default function GuestLayout() {
    return (
        <main className="guest-layout">
            <section className="guest-layout__hero">
                <p className="guest-layout__eyebrow">Vox Messenger</p>
                <h1 className="guest-layout__title">Fast, clear messaging with a Telegram-inspired desktop shell.</h1>
                <p className="guest-layout__text">Authentication and chat behavior stay intact while the interface is rebuilt on a lighter visual system.</p>
            </section>
            <div className="guest-layout__body">
                <Outlet />
            </div>
        </main>
    );
}
