export interface Links {
    links: Link[];
    pagination: Pagination;
}

export interface Link {
    id: string;
    originalURL: string;
    shortCode: string;
    title: string;
    isActive: boolean;
    clickLimit: number;
    expireAt: Date | null;
    createdAt: Date;
    updatedAt: Date;
    totalClicks: number;
}

export interface Pagination {
    total: number;
    limit: number;
    offset: number;
    hasMore: boolean;
}
