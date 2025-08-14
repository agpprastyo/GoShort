import type {Links, CreateLinkPayload, CreateLinkResponse} from "@/types/links";


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

    [property: string]: string | number | boolean | Date | undefined;
}


export const Order = {
    CreatedAt: "created_at",
    IsActive: "is_active",
    Title: "title",
    UpdatedAt: "updated_at",
} as const;

export type Order = typeof Order[keyof typeof Order];


export const getLinks = async (request: Request): Promise<Links> => {
    const params = new URLSearchParams(request as Record<string, string>).toString();
    console.log("Request parameters:", params);
    const response = await fetch(`${api}/links?${params}`, {
        method: 'GET',
        credentials: 'include',
    });

    if (!response.ok) {
        throw new Error('Failed to fetch links');
    }

    const data = await response.json();
    console.debug("Response data 1:", data);

    return data as Links;
}

export const updateLinkStatus = async (id: string, isActive: boolean): Promise<any> => {
    const response = await fetch(`${api}/links/${id}/status`, {
        method: 'PATCH',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({is_active: isActive}),
    });
    if (!response.ok) {
        throw new Error('Failed to update link status');
    }
    return await response.json();
};

export const createLink = async (payload: CreateLinkPayload): Promise<CreateLinkResponse> => {
    const response = await fetch(`${api}/links`, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => null);
        throw new Error(errorData?.message || 'Failed to create short link');
    }

    return await response.json() as CreateLinkResponse;
};
