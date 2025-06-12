import {RegisterForm} from "@/components/register-form.tsx";

function RegisterPage() {
    return (
        <div className='flex items-center justify-center min-h-screen'>
            <RegisterForm  className='md:w-xl w-full mx-4'/>
        </div>

    );
}

export default RegisterPage;