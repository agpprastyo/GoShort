import {useEffect, useState} from "react";

import {Link, useNavigate} from "react-router-dom";
import {getStoredUser} from "@/lib/utils";
import {Button} from "@/components/ui/button.tsx";
import {Input} from "@/components/ui/input.tsx";

export default function LandingPage() {
    const [url, setUrl] = useState("");
    const [shortUrl, setShortUrl] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    const handleShorten = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setShortUrl(null);
        setTimeout(() => {
            setShortUrl("https://goshort.io/abcd12");
            setLoading(false);
        }, 1200);
    };

    const [user, setUser] = useState(() => getStoredUser());
    const navigate = useNavigate();

    useEffect(() => {
        setUser(getStoredUser());
    }, []);

    const handleDashboard = () => {
        if (user && user.username) {
            navigate(`/${user.username}`);
        }
    };

    return (
        <main className="min-h-screen flex flex-col ">
            {/* Navbar */}
            <nav
                className="w-full flex items-center justify-between px-6 py-4 bg-primary-foreground shadow-sm fixed top-0 left-0 z-10 backdrop-blur">
                <Link to="/" className="text-2xl font-extrabold text-primary tracking-tight">
                    GoShort
                </Link>
                <div className="flex gap-4">
                    {user ? (
                        <Button
                            variant="outline"
                            className="text-sm font-medium"
                            onClick={handleDashboard}
                        >
                            Dashboard
                        </Button>
                    ) : (
                        <>
                            <Link to="/login" className="text-blue-600 font-medium hover:underline">
                                Login
                            </Link>
                            <Link to="/register" className="text-blue-600 font-medium hover:underline">
                                Register
                            </Link>
                        </>
                    )}
                </div>
            </nav>

            {/* Add padding top to avoid overlap with fixed navbar */}
            <div className="pt-20">
                {/* Hero Section */}
                <header className="py-12 px-4 text-center">
                    <h1 className="text-4xl md:text-5xl font-extrabold text-primary mb-4">
                        GoShort
                    </h1>
                    <p className="text-lg md:text-xl text-gray-700 mb-6">
                        The simplest way to shorten, share, and track your links.
                    </p>
                    <form
                        className="max-w-xl mx-auto flex flex-col md:flex-row gap-3 items-center"
                        onSubmit={handleShorten}
                        aria-label="Shorten your URL"
                    >

                        <Input
                            type="url"
                            placeholder="Enter your long URL here"
                            value={url}
                            onChange={(e) => setUrl(e.target.value)}
                            className="flex-1 w-full md:w-auto"
                            required
                        />

                        <Button
                            type="submit"
                            className="w-full md:w-auto"
                            disabled={loading || !url}
                        >

                            {loading ? "Shortening..." : "Shorten URL"}
                        </Button>
                    </form>
                    {shortUrl && (
                        <div className="mt-4">
                            <span className="text-gray-600">Your short link:</span>
                            <a
                                href={shortUrl}
                                className="ml-2 text-blue-600 font-semibold underline"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {shortUrl}
                            </a>
                        </div>
                    )}
                </header>

                {/* Features Section */}
                <section className="flex-1 py-12 px-4 bg-white">
                    <div className="max-w-4xl mx-auto grid md:grid-cols-3 gap-8 text-center">
                        <div>
                            <div
                                className="mx-auto mb-3 w-12 h-12 flex items-center justify-center rounded-full bg-blue-100 text-blue-600 text-2xl">
                                🔗
                            </div>
                            <h2 className="font-bold text-lg mb-2">Instant Shortening</h2>
                            <p className="text-gray-600">
                                Shorten any link in seconds with a single click.
                            </p>
                        </div>
                        <div>
                            <div
                                className="mx-auto mb-3 w-12 h-12 flex items-center justify-center rounded-full bg-green-100 text-green-600 text-2xl">
                                📊
                            </div>
                            <h2 className="font-bold text-lg mb-2">Track Clicks</h2>
                            <p className="text-gray-600">
                                Monitor your link performance with real-time analytics.
                            </p>
                        </div>
                        <div>
                            <div
                                className="mx-auto mb-3 w-12 h-12 flex items-center justify-center rounded-full bg-yellow-100 text-yellow-600 text-2xl">
                                🔒
                            </div>
                            <h2 className="font-bold text-lg mb-2">Safe & Secure</h2>
                            <p className="text-gray-600">
                                Your links are protected and your data is private.
                            </p>
                        </div>
                    </div>
                </section>

                {/* Call to Action Section */}
                <section className="py-12 px-4 bg-gray-50">
                    <div className="max-w-2xl mx-auto text-center">
                        <h2 className="text-2xl md:text-3xl font-bold text-primary mb-4">
                            Ready to get started?
                        </h2>
                        <p className="text-lg text-gray-700 mb-6">
                            Join thousands of users who trust GoShort for their link management.
                        </p>
                        <Link
                            to="/register"
                            className="inline-block bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
                        >
                            Create an Account
                        </Link>
                    </div>
                </section>


                // {/* Testimonials Section */}
                <section className="py-12 px-4 bg-white">
                    <div className="max-w-2xl mx-auto text-center">
                        <h2 className="text-2xl md:text-3xl font-bold text-primary mb-4">
                            What our users say
                        </h2>
                        <p className="text-lg text-gray-700 mb-6">
                            Hear from our satisfied users about their experience with GoShort.
                        </p>
                        <div className="space-y-6">
                            <blockquote className="border-l-4 border-blue-600 pl-4 italic text-gray-600">
                                "GoShort has transformed the way I share links. It's fast, reliable, and easy to use!"
                                - Jane Doe
                            </blockquote>
                            <blockquote className="border-l-4 border-green-600 pl-4 italic text-gray-600">
                                "The analytics feature is a game-changer. I can see exactly how my links are
                                performing."
                                - John Smith
                            </blockquote>
                        </div>
                    </div>
                </section>

            </div>

            {/* Footer */}
            <footer className="py-6 text-center text-gray-500 text-sm border-t bg-primary-foreground">
                &copy; {new Date().getFullYear()} GoShort. All rights reserved.
            </footer>
        </main>
    );
}