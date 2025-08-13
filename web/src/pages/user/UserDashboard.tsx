import { Switch } from "@/components/ui/switch";
import { useAuth } from "@/hooks/useAuth";
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import type { Link, CreateLinkPayload } from "@/types/links.ts";
import { toast } from "sonner";
import { getLinks, Order, type Request, updateLinkStatus, createLink } from "@/lib/api/LinksApi.ts";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
    Pagination,
    PaginationContent,
    PaginationItem,
    PaginationNext,
    PaginationPrevious
} from "@/components/ui/pagination.tsx";
import { cn } from "@/lib/utils.ts";
import { Label } from "@/components/ui/label.tsx";


export default function UserDashboard() {
    const { username } = useParams<{ username: string }>();
    const { user, logout } = useAuth();


    const [shortLinks, setShortLinks] = useState<Link[]>([]);
    const [loadingLinks, setLoadingLinks] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);

    const [isCreating, setIsCreating] = useState(false);
    const [url, setUrl] = useState("");
    const [title, setTitle] = useState("");
    const [shortCode, setShortCode] = useState("");
    const [clickLimit, setClickLimit] = useState<number | string>("");
    const [expireAt, setExpireAt] = useState("");

    const [searchQuery, setSearchQuery] = useState("");
    const [sortOrder, setSortOrder] = useState<Order>(Order.CreatedAt);
    const [sortAscending, setSortAscending] = useState(false);

    const handleStatusToggle = async (id: string, currentStatus: boolean) => {
        try {
            await updateLinkStatus(id, !currentStatus);
            setShortLinks(links => links.map(link =>
                link.id === id ? { ...link, is_active: !currentStatus } : link
            ));
            toast.success("Link status updated successfully");
        } catch (error) {
            toast.error("Failed to update link status");
            console.error("Error updating link status:", error);
        }
    };

    const fetchUserLinks = async (
        page: number,
        search: string = searchQuery,
        order: Order = sortOrder,
        ascending: boolean = sortAscending
    ) => {
        setLoadingLinks(true);
        try {
            const limit = 10;
            const offset = (page - 1) * limit;

            const request: Request = {
                limit,
                offset,
                order,
                ascending,
                ...(search && search.trim() !== "" ? { search } : {})
            };

            const response = await getLinks(request);
            const linksData = response?.data?.links || [];
            const pagination = response?.data?.pagination || { total: 0 };

            const mappedLinks: Link[] = linksData.map(link => ({
                ...link,
                short_code: `${import.meta.env.VITE_BASE_URL || "http://localhost:8080"}/${link.short_code || ""}`,
            }));

            const totalItems = pagination.total;
            const calculatedTotalPages = Math.ceil(totalItems / limit);

            setShortLinks(mappedLinks);
            setTotalPages(calculatedTotalPages > 0 ? calculatedTotalPages : 1);

        } catch (error) {
            toast.error("Failed to load your links");
            console.error("Error fetching links:", error);
        } finally {
            setLoadingLinks(false);
        }
    };

    const handleCreateLink = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!url.trim()) {
            toast.error("Original URL cannot be empty.");
            return;
        }
        setIsCreating(true);

        // Membangun payload hanya dengan data yang diisi
        const payload: CreateLinkPayload = { original_url: url };
        if (title) payload.title = title;
        if (shortCode) payload.short_code = shortCode;
        if (clickLimit) payload.click_limit = Number(clickLimit);
        if (expireAt) payload.expire_at = new Date(expireAt).toISOString();


        try {
            const response = await createLink(payload);
            toast.success(response.message || "Link created successfully!");

            // Reset form
            setUrl("");
            setTitle("");
            setShortCode("");
            setClickLimit("");
            setExpireAt("");

            // Refresh daftar link
            if (currentPage !== 1) {
                setCurrentPage(1);
            } else {
                await fetchUserLinks(1);
            }

        } catch (error: any) {
            toast.error(error.message || "An unknown error occurred.");
            console.error("Error creating link:", error);
        } finally {
            setIsCreating(false);
        }
    };

    const handleSearch = () => {
        setCurrentPage(1);
        fetchUserLinks(1, searchQuery, sortOrder, sortAscending);
    };

    useEffect(() => {
        fetchUserLinks(currentPage, searchQuery, sortOrder, sortAscending);
    }, [username, currentPage, sortOrder, sortAscending]);

    const handleLogout = () => {
        logout();
        toast.success("Logged out successfully");
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast.success("Link copied to clipboard!");
    };

    const formatDate = (dateString: string | Date) => {
        const date = dateString ? new Date(dateString) : null;
        return date && !isNaN(date.getTime())
            ? date.toLocaleDateString()
            : "N/A";
    };

    return (
        <div className="min-h-screen flex flex-col bg-background">
            <header className="bg-card text-card-foreground shadow-sm border-b">
                <div className="max-w-7xl mx-auto px-4 py-5 sm:px-6 flex justify-between items-center">
                    <div>
                        <h1 className="text-2xl font-bold">GoShort Dashboard</h1>
                        <p className="text-muted-foreground">Welcome back, {user?.username}!</p>
                    </div>
                    <Button onClick={handleLogout} variant="destructive">Logout</Button>
                </div>
            </header>

            <main className="flex-grow max-w-7xl w-full mx-auto px-4 py-8 sm:px-6">
                <Card className="mb-8">
                    <CardHeader><h2 className="text-xl font-semibold">Create New Short Link</h2></CardHeader>
                    <CardContent>
                        <form onSubmit={handleCreateLink} className="space-y-4">
                            <div>
                                <Label htmlFor="originalUrl">Original URL *</Label>
                                <Input
                                    id="originalUrl"
                                    type="url"
                                    required
                                    placeholder="https://your-very-long-url.com/goes-here"
                                    value={url}
                                    onChange={(e) => setUrl(e.target.value)}
                                    disabled={isCreating}
                                />
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <Label htmlFor="title">Title (Optional)</Label>
                                    <Input
                                        id="title"
                                        type="text"
                                        placeholder="My Awesome Link"
                                        value={title}
                                        onChange={(e) => setTitle(e.target.value)}
                                        disabled={isCreating}
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="shortCode">Custom Short Code (Optional)</Label>
                                    <Input
                                        id="shortCode"
                                        type="text"
                                        placeholder="custom-code"
                                        value={shortCode}
                                        onChange={(e) => setShortCode(e.target.value)}
                                        disabled={isCreating}
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="clickLimit">Click Limit (Optional)</Label>
                                    <Input
                                        id="clickLimit"
                                        type="number"
                                        placeholder="e.g., 100"
                                        value={clickLimit}
                                        onChange={(e) => setClickLimit(e.target.value)}
                                        disabled={isCreating}
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="expireAt">Expiration Date (Optional)</Label>
                                    <Input
                                        id="expireAt"
                                        type="datetime-local"
                                        value={expireAt}
                                        onChange={(e) => setExpireAt(e.target.value)}
                                        disabled={isCreating}
                                    />
                                </div>
                            </div>
                            <div className="flex justify-end">
                                <Button
                                    type="submit"
                                    disabled={isCreating || !url}
                                >
                                    {isCreating ? "Creating..." : "Shorten URL"}
                                </Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <div className="space-y-3">
                            <h2 className="text-xl font-semibold">Your Short Links</h2>
                            <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
                                <Input
                                    placeholder="Search links..."
                                    className="w-full sm:w-64"
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                                />
                                <div className="flex items-center gap-2 ml-auto">
                                    <select
                                        className="text-sm border rounded p-1"
                                        value={sortOrder}
                                        onChange={(e) => setSortOrder(e.target.value as Order)}
                                    >
                                        <option value={Order.CreatedAt}>Created Date</option>
                                        <option value={Order.UpdatedAt}>Updated Date</option>
                                        <option value={Order.Title}>Title</option>
                                        <option value={Order.IsActive}>Status</option>
                                    </select>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => setSortAscending(!sortAscending)}
                                        className="flex gap-1"
                                    >
                                        {sortAscending ? "Ascending" : "Descending"}
                                        {sortAscending ? (
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                                 viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
                                                 strokeLinecap="round" strokeLinejoin="round">
                                                <path d="m3 8 4-4 4 4"/>
                                                <path d="M7 4v16"/>
                                                <path d="M11 12h4"/>
                                                <path d="M11 16h7"/>
                                                <path d="M11 20h10"/>
                                            </svg>
                                        ) : (
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                                 viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
                                                 strokeLinecap="round" strokeLinejoin="round">
                                                <path d="m3 16 4 4 4-4"/>
                                                <path d="M7 20V4"/>
                                                <path d="M11 4h4"/>
                                                <path d="M11 8h7"/>
                                                <path d="M11 12h10"/>
                                            </svg>
                                        )}
                                    </Button>
                                    <Button size="sm" onClick={handleSearch}>Apply</Button>
                                </div>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent>
                        {loadingLinks && shortLinks.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground">Loading your links...</div>
                        ) : shortLinks.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground">You haven't created any links
                                yet.</div>
                        ) : (
                            <>
                                <div className="overflow-x-auto">
                                    <table className="w-full text-sm">
                                        <thead className="text-left text-muted-foreground border-b">
                                        <tr>
                                            <th className="p-2 font-medium">Title</th>
                                            <th className="p-2 font-medium">Original URL</th>
                                            <th className="p-2 font-medium">Short URL</th>
                                            <th className="p-2 font-medium">Clicks</th>
                                            <th className="p-2 font-medium">Created</th>
                                            <th className="p-2 font-medium">Active</th>
                                            <th className="p-2 font-medium text-right">Actions</th>
                                        </tr>
                                        </thead>
                                        <tbody className="divide-y divide-border">
                                        {shortLinks.map((link) => (
                                            <tr key={link.id} className="hover:bg-muted/50">
                                                <td className="p-2 max-w-[200px] md:max-w-xs truncate cursor-pointer"
                                                    title={link.title}>
                                                    {link.title || "No title"}
                                                </td>
                                                <td
                                                    className="p-2 max-w-[200px] md:max-w-xs truncate cursor-pointer text-blue-600 hover:underline"
                                                    title={link.original_url}
                                                >
                                                    <a
                                                        href={link.original_url}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="flex items-center gap-1"
                                                    >
                                                        {link.original_url}
                                                        <svg
                                                            xmlns="http://www.w3.org/2000/svg"
                                                            className="w-4 h-4 inline-block"
                                                            fill="none"
                                                            viewBox="0 0 24 24"
                                                            stroke="currentColor"
                                                        >
                                                            <path strokeLinecap="round" strokeLinejoin="round"
                                                                  strokeWidth={2}
                                                                  d="M14 3h7m0 0v7m0-7L10 14m-4 0v7a2 2 0 002 2h7"/>
                                                        </svg>
                                                    </a>
                                                </td>
                                                <td className="p-2 max-w-[200px] md:max-w-xs truncate cursor-pointer text-green-500 hover:underline"
                                                    title={link.short_code}>
                                                    <a
                                                        href={link.short_code}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="flex items-center gap-1"
                                                    >
                                                        {link.short_code}
                                                        <svg
                                                            xmlns="http://www.w3.org/2000/svg"
                                                            className="w-4 h-4 inline-block"
                                                            fill="none"
                                                            viewBox="0 0 24 24"
                                                            stroke="currentColor"
                                                        >
                                                            <path strokeLinecap="round" strokeLinejoin="round"
                                                                  strokeWidth={2}
                                                                  d="M14 3h7m0 0v7m0-7L10 14m-4 0v7a2 2 0 002 2h7"/>
                                                        </svg>
                                                    </a>
                                                </td>
                                                <td className="p-2">{link.total_clicks}</td>
                                                <td className="p-2">{formatDate(link.created_at)}</td>
                                                <td className="p-2">
                                                    <Switch
                                                        checked={link.is_active}
                                                        onCheckedChange={() => handleStatusToggle(link.id, link.is_active)}
                                                        aria-label="Toggle link status"
                                                    />
                                                </td>
                                                <td className="p-2 text-right">
                                                    <Button onClick={() => copyToClipboard(link.short_code)}
                                                            variant="secondary" size="sm">Copy</Button>
                                                </td>
                                            </tr>
                                        ))}
                                        </tbody>
                                    </table>
                                </div>
                                <Pagination className="mt-6">
                                    <div className="flex w-full justify-between items-center">
                                        <span
                                            className="text-sm text-muted-foreground">Page {currentPage} of {totalPages}</span>
                                        <PaginationContent className="ml-auto">
                                            <PaginationItem><PaginationPrevious
                                                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                                                className={cn({"pointer-events-none opacity-50": currentPage === 1})}/></PaginationItem>
                                            <PaginationItem><PaginationNext
                                                onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                                                className={cn({"pointer-events-none opacity-50": currentPage === totalPages})}/></PaginationItem>
                                        </PaginationContent>
                                    </div>
                                </Pagination>
                            </>
                        )}
                    </CardContent>
                </Card>
            </main>
            <div className="my-8 pt-8 border-t text-center text-sm">
                <p>&copy; {new Date().getFullYear()} GoShort. All rights reserved.</p>
            </div>
        </div>
    );
}
