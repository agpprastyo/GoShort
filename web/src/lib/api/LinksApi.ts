import type {Links} from "@/types/links";


const baseUrl = import.meta.env.VITE_BASE_URL as string;
const api = `${baseUrl}/api/v1`;


export interface Request {
    ascending?: boolean;
    end_date?: string;
    limit?: number;
    offset?: number;
    order?: Order;
    search?: string;
    start_date?: Date;

    [property: string]: any;
}

export enum Order {
    CreatedAt = "created_at",
    IsActive = "is_active",
    Title = "title",
    UpdatedAt = "updated_at",
}

// Fix LinksApi.ts
export const getLinks = async (request: Request): Promise<Links> => {
    const params = new URLSearchParams(request as Record<string, string>).toString();
    const response = await fetch(`${api}/links?${params}`, {
        method: 'GET',
        credentials: 'include',
    });

    if (!response.ok) {
        throw new Error('Failed to fetch links');
    }

    // Remove the console.log that consumes the response
    // console.log("Response : ", response.json())  <- This was the problem!

    const data = await response.json();
    console.log("Response data 1:", data);

    return data as Links;
}