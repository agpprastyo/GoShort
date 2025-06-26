export interface Login {
    message: string;
    data: LoginData;
}

export interface LoginData {
    data: DataData;
    expires_at: Date;
    logged_in: boolean;
}

export interface DataData {
    id: string;
    username: string;
    email: string;
    first_name: null;
    last_name: null;
    role: string;
}
