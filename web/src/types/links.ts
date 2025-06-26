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
