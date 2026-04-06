import { Outlet } from "react-router-dom";
import "./desktop-layout.sass";

export default function DesktopLayout() {
    return (
        <main className="desktop-layout">
            <Outlet />
        </main>
    );
}
