export interface Links {
    message: string;
    data: Data;
}

export interface Data {
    links: Link[];
    pagination: Pagination;
}

export interface Link {
    id: string;
    original_url: string;
    short_code: string;
    title: string;
    is_active: boolean;
    click_limit: number;
    expire_at: Date;
    created_at: Date;
    updated_at: Date;
    total_clicks: number;
}

export interface Pagination {
    total: number;
    total_query: number;
    limit: number;
    offset: number;
    has_more: boolean;
}

// Tipe untuk data yang dikirim saat membuat link baru
export interface CreateLinkPayload {
    original_url: string;
    title?: string;
    click_limit?: number;
    expire_at?: Date;
    short_code?: string;
}

export interface CreateLinkResponse {
    message: string;
    data: Link;
}