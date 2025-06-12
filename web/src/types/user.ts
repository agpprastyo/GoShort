
export interface Login {
    data: Data;
    expires_at: number;
    logged_in: boolean;
}

export interface Data {
    id: string;
    username: string;
    email: string;
    first_name: string | null;
    last_name: string | null;
    role: string;
}