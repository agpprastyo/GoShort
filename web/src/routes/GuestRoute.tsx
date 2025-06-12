import {Navigate} from "react-router-dom";
import {getStoredUser} from "@/lib/utils";


export function GuestRoute({children}: { children: JSX.Element }) {
    const user = getStoredUser();
    if (user) {
        return <Navigate to="/" replace/>;
    }
    return children;
}