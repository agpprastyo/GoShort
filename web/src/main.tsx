import {StrictMode} from 'react'
import {createRoot} from 'react-dom/client'
import {BrowserRouter, Route, Routes} from 'react-router-dom'
import './index.css'
import {GuestRoute} from "@/routes/GuestRoute.tsx";
import {UserRoute} from "@/routes/UserRoute.tsx";
import {Toaster} from "@/components/ui/sonner"


import RegisterPage from "@/pages/auth/RegisterPage.tsx";
import LoginPage from "@/pages/auth/LoginPage.tsx";
import UserDashboard from "@/pages/user/UserDashboard.tsx";
import LandingPage from "@/pages/landing/LandingPage.tsx";


function Admin() {
    return <h2>Admin Page</h2>
}

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<LandingPage/>}/>
                <Route
                    path="/login"
                    element={
                        <GuestRoute>
                            <LoginPage/>
                        </GuestRoute>
                    }
                />
                <Route
                    path="/register"
                    element={
                        <GuestRoute>
                            <RegisterPage/>
                        </GuestRoute>
                    }
                />

                <Route
                    path="/:username"
                    element={
                        <UserRoute>
                            <UserDashboard/>
                        </UserRoute>
                    }
                />
                <Route path="/admin" element={<Admin/>}/>
            </Routes>
        </BrowserRouter>
        <Toaster/>
    </StrictMode>,
)