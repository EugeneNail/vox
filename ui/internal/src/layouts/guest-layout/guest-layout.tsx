import { Outlet } from "react-router-dom";
import "./guest-layout.sass";

export default function GuestLayout() {
    return (
        <main className="guest-layout">
            <div className="guest-layout__body">
                <Outlet />
            </div>
        </main>
    );
}
