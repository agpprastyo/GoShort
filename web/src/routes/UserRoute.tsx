import {Navigate, useParams} from "react-router-dom";
import {useAuth} from "@/hooks/useAuth";
import {type JSX} from "react";

export function UserRoute({children}: { children: JSX.Element }) {
    const {user, loading} = useAuth();
    const {username} = useParams();

    if (loading) {
        return <div>Loading...</div>; // Or a spinner component
    }

    if (!user) {
        return <Navigate to="/login" replace/>;
    }

    // Updated to access username through the nested data property
    if (username && user.username !== username) {
        return <Navigate to="/" replace/>;
    }

    return children;
}