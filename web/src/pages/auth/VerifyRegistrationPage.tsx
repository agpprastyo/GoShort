

function VerifyRegistrationPage(email?: string) {

    return (
        <div>
            <h2 className='h2-bold'>Verify your email</h2>
            <p className='text-muted-foreground'>
                A verification link has been sent to your email address {email}. Please check your inbox and click on the link to verify your account.
            </p>
        </div>
    );
}

export default VerifyRegistrationPage;