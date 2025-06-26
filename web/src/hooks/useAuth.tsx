import {createContext, type ReactNode, useContext, useEffect, useState,} from 'react';
import {useNavigate} from 'react-router-dom';
import type {DataData,} from "@/types/user.ts";
import {userLogout} from "@/lib/api/UserApi.ts";


interface AuthContextType {
    user: DataData | null;
    loading: boolean;
    login: (data: DataData, expires_at: string) => void;
    logout: () => void;
}

// Create the context with a default value
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Define a helper function to get and validate the stored user
const getStoredUser = (): DataData | null => {
    const storedUser = localStorage.getItem("user");
    if (!storedUser) {
        return null;
    }

    try {
        const {data, expires_at} = JSON.parse(storedUser);
        if (new Date() > new Date(expires_at)) {
            localStorage.removeItem("user");
            return null; // Session expired
        }
        return data as DataData;
    } catch {
        localStorage.removeItem("user");
        return null;
    }
};


// Create the AuthProvider component
export function AuthProvider({children}: { children: ReactNode }) {
    const [user, setUser] = useState<DataData | null>(null);
    const [loading, setLoading] = useState(true); // Start with loading true
    const navigate = useNavigate();

    // Check for user session on initial load
    useEffect(() => {
        const storedUser = getStoredUser();
        if (storedUser) {
            setUser(storedUser);
        }
        setLoading(false); // Finished initial check
    }, []);

    const login = (data: DataData, expires_at: string) => {
        localStorage.setItem("user", JSON.stringify({data, expires_at}));
        setUser(data);
        navigate(`/${data.username}`);
    };

    const logout = async () => {
        try {
            await userLogout();
            // Clear user data from state
            setUser(null);
            // Clear any stored tokens or session data
            localStorage.removeItem("user");
            sessionStorage.removeItem("user");
            // Redirect to login page if needed
            window.location.href = "/login";
        } catch (error) {
            console.error("Logout failed:", error);
            throw error;
        }
    };

    const value = {user, loading, login, logout};

    // Don't render children until the loading state is false
    return (
        <AuthContext.Provider value={value}>
            {!loading && children}
        </AuthContext.Provider>
    );
}

// Create a custom hook to easily use the auth context
export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};