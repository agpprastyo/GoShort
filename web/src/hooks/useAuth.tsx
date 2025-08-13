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

const AuthContext = createContext<AuthContextType | undefined>(undefined);

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

export function AuthProvider({children}: { children: ReactNode }) {
    const [user, setUser] = useState<DataData | null>(null);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();


    useEffect(() => {
        const storedUser = getStoredUser();
        if (storedUser) {
            setUser(storedUser);
        }
        setLoading(false);
    }, []);

    const login = (data: DataData, expires_at: string) => {
        localStorage.setItem("user", JSON.stringify({data, expires_at}));
        setUser(data);
        navigate(`/${data.username}`);
    };

    const logout = async () => {
        setUser(null);
        localStorage.removeItem("user");
        sessionStorage.removeItem("user");
        navigate("/login");

        try {
            await userLogout();
        } catch (error) {
            console.error("Server logout failed, but client-side logout was successful:", error);
        }
    };

    const value = {user, loading, login, logout};

    return (
        <AuthContext.Provider value={value}>
            {!loading && children}
        </AuthContext.Provider>
    );
}

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};