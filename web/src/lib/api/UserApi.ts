import type {Login} from "@/types/user";


const baseUrl = "http://localhost:8080";
const api = `${baseUrl}/api/v1`;



export const userLogin = async (
  data: { email: string; password: string }
): Promise<Login> => {
  const response = await fetch(`${api}/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    throw new Error('Login failed');
  }

  return response.json() as Promise<Login>;
};

// logout function
export const userLogout = async (): Promise<void> => {
  const response = await fetch(`${api}/logout`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Logout failed');
  }
};