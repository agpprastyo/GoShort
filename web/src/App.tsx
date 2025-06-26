import {useState} from 'react';
import reactLogo from './assets/react.svg';
import viteLogo from '/vite.svg';
import {userLogout} from '@/lib/api/UserApi.ts';
import {toast} from 'sonner';
import {useNavigate} from 'react-router-dom';


function App() {
    const [count, setCount] = useState(0);
    const handleLogout = async () => {
        try {
            await userLogout();
            toast.success('You have been successfully logged out.');
        } catch {
            toast.error('Failed to log out. Please try again.');
        }
        localStorage.removeItem('user');
    };

    const navigate = useNavigate();

    const handleDashboard = () => {
        const user = JSON.parse(localStorage.getItem('user') || '{}');
        if (user && user.username) {
            navigate(`/${user.username}`);
        } else {
            // Optionally handle if user is not logged in
            navigate('/login');
        }
    };


    return (
        <>

            <div>
                <a href="https://vite.dev" target="_blank">
                    <img src={viteLogo} className="logo" alt="Vite logo"/>
                </a>
                <a href="https://react.dev" target="_blank">
                    <img src={reactLogo} className="logo react" alt="React logo"/>
                </a>
            </div>
            <h1>Vite + React</h1>
            <div className="card">
                <button onClick={() => setCount((count) => count + 1)}>
                    count is {count}
                </button>
                <p>
                    Edit <code>src/App.tsx</code> and save to test HMR
                </p>
            </div>
            <p className="read-the-docs">
                Click on the Vite and React logos to learn more
            </p>
            <div className="flex justify-center mt-4">
                <button
                    className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 mr-2"
                    onClick={handleDashboard}
                >
                    Dashboard
                </button>


                <button
                    className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                    onClick={handleLogout}
                >
                    Logout
                </button>
            </div>
        </>
    );
}

export default App;