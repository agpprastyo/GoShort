import {Navigate, useParams} from "react-router-dom";
import {getStoredUser} from "@/lib/utils";

export function UserRoute({children}: { children: JSX.Element }) {
    const user = getStoredUser();
    const {username} = useParams();

    if (!user) {
        return <Navigate to="/login" replace/>;
    }

    // Optionally, check if the username matches the stored user
    if (username && user.username !== username) {
        return <Navigate to="/" replace/>;
    }

    return children;
}