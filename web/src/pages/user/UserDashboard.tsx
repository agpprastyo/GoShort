import {useNavigate, useParams} from "react-router-dom";
import {useEffect, useState} from "react";
import {getStoredUser} from "@/lib/utils";
import {toast} from "sonner";
import {
    Pagination,
    PaginationContent,
    PaginationItem,
    PaginationNext,
    PaginationPrevious,
} from "@/components/ui/pagination";
import {getLinks, Order, Request} from "@/lib/api/LinksApi";

interface ShortLink {
    id: string;
    originalUrl: string;
    shortUrl: string;
    clicks: number;
    createdAt: string;
}

export default function UserDashboard() {
    const {username} = useParams();
    const navigate = useNavigate();
    const user = getStoredUser();
    const [url, setUrl] = useState("");
    const [loading, setLoading] = useState(false);
    const [links, setLinks] = useState<ShortLink[]>([]);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);

    useEffect(() => {
        // Redirect if not logged in or username doesn't match
        if (!user || user.username !== username) {
            navigate("/login");
            return;
        }

        fetchUserLinks(currentPage);
    }, [username, currentPage]);

    const fetchUserLinks = async (page: number) => {
        setLoading(true);
        try {
            const limit = 10;
            const offset = (page - 1) * limit;

            const request: Request = {
                limit,
                offset,
                order: Order.CreatedAt,
                ascending: false
            };

            const response = await getLinks(request);

            const mappedLinks: ShortLink[] = response.links.map(link => {
                // Format date safely
                let formattedDate;
                try {
                    formattedDate = typeof link.createdAt === 'string'
                        ? link.createdAt
                        : typeof link.createdAt === 'string'
                            ? link.createdAt
                            : new Date().toString();
                } catch {
                    formattedDate = new Date().toString();
                }

                return {
                    id: link.id,
                    originalUrl: link.originalURL || link.original_url, // Handle both naming conventions
                    shortUrl: `${import.meta.env.VITE_BASE_URL}/${link.shortCode || link.short_code}`,
                    clicks: link.totalClicks || link.total_clicks || 0,
                    createdAt: formattedDate
                };
            });

            const totalItems = response.pagination.total;
            const calculatedTotalPages = Math.ceil(totalItems / limit);

            setLinks(mappedLinks);
            setTotalPages(calculatedTotalPages > 0 ? calculatedTotalPages : 1);
        } catch (error) {
            toast.error("Failed to load your links");
            console.error("Error fetching links:", error);
        } finally {
            setLoading(false);
        }
    };
    const handleCreateShortLink = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!url) return;

        setLoading(true);
        try {
            // Replace with actual API call
            setTimeout(() => {
                const newLink: ShortLink = {
                    id: `link-new-${Math.random()}`,
                    originalUrl: url,
                    shortUrl: `https://goshort.io/${Math.random().toString(36).substring(2, 7)}`,
                    clicks: 0,
                    createdAt: new Date().toISOString()
                };

                setLinks([newLink, ...links.slice(0, 9)]);
                setUrl("");
                toast.success("Link created successfully!");
                setLoading(false);
            }, 800);
        } catch (error) {
            toast.error("Failed to create short link");
            setLoading(false);
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast.success("Link copied to clipboard!");
    };

    const handleLogout = () => {
        localStorage.removeItem("user");
        navigate("/login");
        toast.success("Logged out successfully");
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString();
    };

    return (
        <div className="min-h-screen flex flex-col bg-background">
            {/* Header */}
            <header className="bg-card text-card-foreground shadow-md">
                <div className="max-w-7xl mx-auto px-4 py-5 sm:px-6 flex justify-between items-center">
                    <div>
                        <h1 className="text-2xl font-bold">GoShort Dashboard</h1>
                        <p className="text-muted-foreground">Welcome back, {username}!</p>
                    </div>
                    <button
                        onClick={handleLogout}
                        className="bg-primary text-primary-foreground px-4 py-2 rounded hover:bg-primary/90 transition"
                    >
                        Logout
                    </button>
                </div>
            </header>

            <main className="flex-grow max-w-7xl w-full mx-auto px-4 py-8 sm:px-6">
                {/* Create new short link section */}
                <section className="mb-8 bg-card p-6 rounded-lg shadow-sm border">
                    <h2 className="text-xl font-semibold mb-4">Create New Short Link</h2>
                    <form onSubmit={handleCreateShortLink} className="flex flex-col sm:flex-row gap-3">
                        <input
                            type="url"
                            required
                            placeholder="Paste your long URL here..."
                            className="flex-grow px-4 py-2 rounded border border-input focus:outline-none focus:ring-2 focus:ring-primary"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                            disabled={loading}
                        />
                        <button
                            type="submit"
                            disabled={loading}
                            className="px-6 py-2 bg-primary text-primary-foreground font-medium rounded hover:bg-primary/90 transition disabled:opacity-50"
                        >
                            {loading ? "Creating..." : "Shorten URL"}
                        </button>
                    </form>
                </section>

                {/* Links listing */}
                <section className="bg-card p-6 rounded-lg shadow-sm border">
                    <h2 className="text-xl font-semibold mb-4">Your Short Links</h2>

                    {loading && currentPage === 1 ? (
                        <div className="text-center py-8">Loading your links...</div>
                    ) : links.length === 0 ? (
                        <div className="text-center py-8 text-muted-foreground">
                            You haven't created any links yet.
                        </div>
                    ) : (
                        <>
                            <div className="overflow-x-auto">
                                <table className="w-full">
                                    <thead className="text-left text-muted-foreground border-b">
                                    <tr>
                                        <th className="pb-2 pl-2">Original URL</th>
                                        <th className="pb-2">Short URL</th>
                                        <th className="pb-2">Clicks</th>
                                        <th className="pb-2">Created</th>
                                        <th className="pb-2">Actions</th>
                                    </tr>
                                    </thead>
                                    <tbody className="divide-y">
                                    {links.map((link) => (
                                        <tr key={link.id} className="hover:bg-muted/50">
                                            <td className="py-3 pl-2 max-w-[200px] truncate" title={link.originalUrl}>
                                                {link.originalUrl}
                                            </td>
                                            <td className="py-3 font-medium">{link.shortUrl}</td>
                                            <td className="py-3">{link.clicks}</td>
                                            <td className="py-3">{formatDate(link.createdAt)}</td>
                                            <td className="py-3">
                                                <button
                                                    onClick={() => copyToClipboard(link.shortUrl)}
                                                    className="text-sm px-3 py-1 bg-secondary text-secondary-foreground rounded hover:bg-secondary/80 transition mr-2"
                                                >
                                                    Copy
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>

                            {/* Pagination */}

                            <Pagination className="mt-6">
                                <div className="flex w-full justify-between items-center">
    <span className="text-sm text-muted-foreground">
      Page {currentPage} of {totalPages}
    </span>

                                    <PaginationContent className="ml-auto">
                                        <PaginationItem>
                                            <PaginationPrevious
                                                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                                                className={currentPage === 1 ? "pointer-events-none opacity-50" : ""}
                                                href="#"
                                            />
                                        </PaginationItem>

                                        <PaginationItem>
                                            <PaginationNext
                                                onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                                                className={currentPage === totalPages ? "pointer-events-none opacity-50" : ""}
                                                href="#"
                                            />
                                        </PaginationItem>
                                    </PaginationContent>
                                </div>
                            </Pagination>
                        </>
                    )}
                </section>
            </main>

            {/* Footer */}
            <footer className="border-t bg-muted/50">
                <div className="max-w-7xl mx-auto px-4 py-6 text-center text-muted-foreground text-sm">
                    &copy; {new Date().getFullYear()} GoShort. All rights reserved.
                </div>
            </footer>
        </div>
    );
}