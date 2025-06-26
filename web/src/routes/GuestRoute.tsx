import {Navigate} from "react-router-dom";
import {useAuth} from "@/hooks/useAuth";
import {type JSX} from "react";

export function GuestRoute({children}: { children: JSX.Element }) {
    const {user, loading} = useAuth();

    if (loading) {
        return <div>Loading...</div>; // Or a spinner component
    }

    if (user && user.data && user.data.username) {
        return <Navigate to={`/${user.data.username}`} replace/>;
    }

    return children;
}