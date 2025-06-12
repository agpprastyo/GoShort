import {LoginForm} from "@/components/login-form.tsx";

function LoginPage() {
    return (
        <div className='flex items-center justify-center min-h-screen'>
            <LoginForm  className='md:w-xl w-full mx-4'/>
        </div>

    );
}

export default LoginPage;