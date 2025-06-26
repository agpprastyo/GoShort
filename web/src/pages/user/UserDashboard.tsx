// Add these imports at the top
import {Switch} from "@/components/ui/switch";
import {useAuth} from "@/hooks/useAuth";
import {useEffect, useState} from "react";
import {useParams} from "react-router-dom";
import type {Link} from "@/types/links.ts";
import {toast} from "sonner";
import {getLinks, Order, type Request, updateLinkStatus} from "@/lib/api/LinksApi.ts";
import {Card, CardContent, CardHeader} from "@/components/ui/card";
import {Input} from "@/components/ui/input";
import {Button} from "@/components/ui/button";
import {
    Pagination,
    PaginationContent,
    PaginationItem,
    PaginationNext,
    PaginationPrevious
} from "@/components/ui/pagination.tsx";
import {cn} from "@/lib/utils.ts";


export default function UserDashboard() {
    const {username} = useParams<{ username: string }>();

    const {user, logout} = useAuth();

    const [shortLinks, setShortLinks] = useState<Link[]>([]);
    const [url, setUrl] = useState("");
    const [loadingLinks, setLoadingLinks] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);

    const [searchQuery, setSearchQuery] = useState("");
    const [sortOrder, setSortOrder] = useState<Order>(Order.CreatedAt);
    const [sortAscending, setSortAscending] = useState(false);

    const handleStatusToggle = async (id: string, currentStatus: boolean) => {
        try {
            await updateLinkStatus(id, !currentStatus);
            setShortLinks(links => links.map(link =>
                link.id === id ? {...link, isActive: !currentStatus, is_active: !currentStatus} : link
            ));
            toast.success("Link status updated successfully");
        } catch (error) {
            toast.error("Failed to update link status");
            console.error("Error updating link status:", error);
        }
    };


    // Move fetchUserLinks outside useEffect to make it available globally in the component
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
                // Only add search property if it has actual content
                ...(search && search.trim() !== "" ? {search} : {})
            };

            const response = await getLinks(request);
            console.log("Full response:", response);

            // The API response matches your Links interface - need to access data correctly
            const linksData = response?.data?.links || [];
            const pagination = response?.data?.pagination || {total: 0};

            const mappedLinks: Link[] = linksData.map(link => ({
                id: link.id || "",
                originalURL: link.original_url || "",
                shortCode: `${import.meta.env.VITE_BASE_URL || "http://localhost:8080"}/${link.short_code || ""}`,
                title: link.title || "",
                isActive: link.is_active,
                createdAt: link.created_at || new Date(),
                updatedAt: link.updated_at || link.created_at || new Date(),
                totalClicks: link.total_clicks || 0,
                clickLimit: link.click_limit ?? 0,
                expireAt: link.expire_at || new Date(),
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

    // Add this handler function
    const handleSearch = () => {
        setCurrentPage(1); // Reset to first page when search/filter changes
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
                        <p className="text-muted-foreground">Welcome back, {user?.data?.username}!</p>
                    </div>
                    <Button onClick={handleLogout} variant="destructive">Logout</Button>
                </div>
            </header>

            <main className="flex-grow max-w-7xl w-full mx-auto px-4 py-8 sm:px-6">
                <Card className="mb-8">
                    <CardHeader><h2 className="text-xl font-semibold">Create New Short Link</h2></CardHeader>
                    <CardContent>
                        {/*TODO implement link creation*/}
                        <form className="flex flex-col sm:flex-row gap-3">
                            <Input type="url" required placeholder="Paste your long URL here..." value={url}
                                   onChange={(e) => setUrl(e.target.value)} disabled={loadingLinks}/>
                            <Button type="submit"
                                    disabled={loadingLinks || !url}>{loadingLinks ? "Creating..." : "Shorten URL"}</Button>
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
                                                    title={link.originalURL}
                                                >
                                                    <a
                                                        href={link.originalURL}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="flex items-center gap-1"
                                                    >
                                                        {link.originalURL}
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
                                                    title={link.shortCode}>
                                                    <a
                                                        href={link.shortCode}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="flex items-center gap-1"
                                                    >
                                                        {link.shortCode}
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
                                                <td className="p-2">{link.totalClicks}</td>
                                                <td className="p-2">{formatDate(link.createdAt)}</td>
                                                <td className="p-2">
                                                    <Switch
                                                        checked={link.isActive}
                                                        onCheckedChange={() => handleStatusToggle(link.id, link.isActive)}
                                                        aria-label="Toggle link status"
                                                    />
                                                </td>
                                                <td className="p-2 text-right">
                                                    <Button onClick={() => copyToClipboard(link.shortCode)}
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
        </div>
    );
}