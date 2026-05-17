import { Outlet } from "react-router-dom";
import "./desktop-layout.sass";

export default function DesktopLayout() {
    return (
        <main className="desktop-layout">
            <div className="desktop-layout__shell">
                <Outlet />
            </div>
        </main>
    );
}
