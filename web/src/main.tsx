// src/main.tsx
import {StrictMode} from 'react';
import {createRoot} from 'react-dom/client';
import {BrowserRouter, Route, Routes} from 'react-router-dom';
import './index.css';
import {Toaster} from "@/components/ui/sonner";

// Import your new AuthProvider
import {AuthProvider} from '@/hooks/useAuth';

// Import Routes and Pages
import RegisterPage from "@/pages/auth/RegisterPage.tsx";
import LoginPage from "@/pages/auth/LoginPage.tsx";
import UserDashboard from "@/pages/user/UserDashboard.tsx";
import LandingPage from "@/pages/landing/LandingPage.tsx";
import {UserRoute} from './routes/UserRoute';
import {GuestRoute} from "@/routes/GuestRoute.tsx";


createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <BrowserRouter>
            <AuthProvider>
                <Routes>
                    <Route path="/" element={<LandingPage/>}/>

                    <Route path="register" element={<GuestRoute><RegisterPage/></GuestRoute>}/>
                    <Route path="login" element={<GuestRoute><LoginPage/></GuestRoute>}/>

                    <Route path=":username" element={<UserRoute><UserDashboard/></UserRoute>}/>
                </Routes>
            </AuthProvider>
        </BrowserRouter>
        <Toaster/>
    </StrictMode>
);
