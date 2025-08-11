import React, {useState} from 'react';
import {Link as ReactRouterLink, useNavigate} from "react-router-dom";
import {ArrowRight, BarChart, Check, Copy, Facebook, Linkedin, Search, ShieldCheck, Twitter} from 'lucide-react';
import {cn} from "@/lib/utils";
import {Button, buttonVariants} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {useAuth} from "@/hooks/useAuth.tsx";


// --- LANDING PAGE COMPONENT ---
export default function LandingPage() {
    const [url, setUrl] = useState("");
    const [shortUrl, setShortUrl] = useState<string | null>(null);
    const [loadingShorten, setLoadingShorten] = useState(false);
    const [copied, setCopied] = useState(false);

    const {user} = useAuth();
    const navigate = useNavigate();

    const handleShorten = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!url) return;
        setLoadingShorten(true);
        setShortUrl(null);
        setCopied(false);

        setTimeout(() => {
            const randomString = Math.random().toString(36).substring(2, 8);
            setShortUrl(`https://goshrt.co/${randomString}`);
            setLoadingShorten(false);
        }, 1500);
    };

    const handleCopy = () => {
        if (!shortUrl) return;
        const textArea = document.createElement("textarea");
        textArea.value = shortUrl;
        document.body.appendChild(textArea);
        textArea.select();
        try {
            navigator.clipboard.writeText(shortUrl);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy text: ', err);
        }
        document.body.removeChild(textArea);
    };

    const handleDashboard = () => {
        console.log('user clicked dashboard, user:', user);
        if (user && user.username) {
            console.log('username:', user.username);
            navigate(`/${user.username}`);
        } else {
            console.log('username not available:', user?.username);
            navigate('/login');
        }
    };

    const features = [
        {
            icon: <Search className="w-8 h-8 text-primary"/>,
            title: "Blazing Fast Shortening",
            description: "Our algorithm creates short, memorable links in milliseconds. Paste your long URL and watch the magic happen."
        },
        {
            icon: <BarChart className="w-8 h-8 text-primary"/>,
            title: "Advanced Analytics",
            description: "Gain insights into every click. Track geolocation, referrers, and devices to understand your audience better."
        },
        {
            icon: <ShieldCheck className="w-8 h-8 text-primary"/>,
            title: "Secure & Reliable",
            description: "With 99.9% uptime and robust security measures, your links are always safe and accessible."
        }
    ];

    const testimonials = [
        {
            quote: "GoShort has become an indispensable tool for our marketing campaigns. The analytics are incredibly detailed and have helped us optimize our strategy.",
            name: "Sarah J.",
            title: "Marketing Manager, TechCorp"
        },
        {
            quote: "I love the simplicity. It's fast, elegant, and does exactly what it promises without any clutter. The custom domain feature is a huge plus!",
            name: "Mike R.",
            title: "Content Creator"
        },
        {
            quote: "As a developer, I appreciate a good API. GoShort's API is well-documented and was a breeze to integrate into our internal tools.",
            name: "Chen W.",
            title: "Lead Developer, Innovate LLC"
        }
    ];

    const Link = ReactRouterLink;

    return (
        <div className="bg-background text-foreground font-sans">
            <nav className="w-full bg-background/80 backdrop-blur-lg border-b fixed top-0 left-0 z-50">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex items-center justify-between h-16">
                        <Link to="/" className="text-2xl font-bold text-primary tracking-tight">GoShort</Link>
                        <div className="hidden md:flex items-center gap-1">
                            <a href="#features" className={buttonVariants({variant: "ghost"})}>Features</a>
                            <a href="#pricing" className={buttonVariants({variant: "ghost"})}>Pricing</a>
                            <a href="#testimonials" className={buttonVariants({variant: "ghost"})}>Testimonials</a>
                        </div>
                        <div className="flex items-center gap-2">
                            {user ? (
                                <Button onClick={handleDashboard}>Dashboard</Button>
                            ) : (
                                <>
                                    <Link to="/login" className={buttonVariants({variant: "ghost"})}>Login</Link>
                                    <Link to="/register" className={buttonVariants({variant: "default"})}>Sign Up
                                        Free</Link>
                                </>
                            )}
                        </div>
                    </div>
                </div>
            </nav>

            <main className="pt-16">
                <header className="relative text-center py-24 sm:py-32 lg:py-40 px-4 overflow-hidden">
                    <div
                        className="absolute inset-0 -z-10 bg-grid-gray-200/40 [mask-image:radial-gradient(ellipse_at_center,transparent_20%,black)]"
                        style={{'--grid-color': 'hsl(var(--border))'} as React.CSSProperties}></div>
                    <div className="max-w-3xl mx-auto">
                        <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tighter">Short
                            Links, <span className="text-primary">Big Results</span></h1>
                        <p className="mt-6 text-lg sm:text-xl text-muted-foreground max-w-2xl mx-auto">The ultimate
                            platform to shorten, share, and analyze your links. Transform long, ugly URLs into powerful
                            marketing assets.</p>

                        <form onSubmit={handleShorten} className="mt-10 max-w-xl mx-auto" aria-label="Shorten URL form">
                            <div className="flex flex-col sm:flex-row gap-3">
                                <Input type="url" placeholder="Paste your long URL here..." value={url}
                                       onChange={(e) => setUrl(e.target.value)} className="h-12 text-base" required/>
                                <Button type="submit" className="h-12 text-base" disabled={loadingShorten || !url}>
                                    {loadingShorten && <svg className="animate-spin -ml-1 mr-3 h-5 w-5"
                                                            xmlns="http://www.w3.org/2000/svg" fill="none"
                                                            viewBox="0 0 24 24">
                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                                strokeWidth="4"></circle>
                                        <path className="opacity-75" fill="currentColor"
                                              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                    </svg>}
                                    {loadingShorten ? "Shortening..." : "Shorten It!"}
                                </Button>
                            </div>
                        </form>

                        {shortUrl && (
                            <div
                                className="mt-6 p-4 bg-secondary border rounded-lg max-w-xl mx-auto flex items-center justify-between transition-all animate-fade-in-up">
                                <a href={shortUrl} target="_blank" rel="noopener noreferrer"
                                   className="font-mono text-primary hover:underline truncate">{shortUrl}</a>
                                <Button onClick={handleCopy} variant={copied ? "secondary" : "default"} size="icon"
                                        className={`transition-all ${copied && 'bg-green-500 hover:bg-green-600'}`}>
                                    {copied ? <Check className="w-4 h-4"/> : <Copy className="w-4 h-4"/>}
                                </Button>
                            </div>
                        )}
                        <p className="mt-4 text-sm text-muted-foreground">No registration required for basic
                            shortening.</p>
                    </div>
                </header>

                <section id="features" className="py-20 sm:py-24 px-4 bg-secondary/50">
                    <div className="max-w-5xl mx-auto text-center">
                        <h2 className="text-3xl sm:text-4xl font-bold">Everything You Need, and More</h2>
                        <p className="mt-4 text-lg text-muted-foreground">GoShort is more than just a link shortener.
                            It's a powerful tool for growth.</p>
                        <div className="mt-16 grid md:grid-cols-3 gap-8">
                            {features.map((feature, index) => (
                                <Card key={index} className="text-center">
                                    <CardHeader>
                                        <div
                                            className="flex items-center justify-center w-16 h-16 bg-primary/10 rounded-full mb-5 mx-auto">{feature.icon}</div>
                                        <CardTitle>{feature.title}</CardTitle>
                                    </CardHeader>
                                    <CardContent><p className="text-muted-foreground">{feature.description}</p>
                                    </CardContent>
                                </Card>
                            ))}
                        </div>
                    </div>
                </section>

                <section id="how-it-works" className="py-20 sm:py-24 px-4 bg-background">
                    <div className="max-w-5xl mx-auto text-center">
                        <h2 className="text-3xl sm:text-4xl font-bold">Get Started in 3 Easy Steps</h2>
                        <p className="mt-4 text-lg text-muted-foreground">Simplify your workflow and start sharing in
                            seconds.</p>
                        <div className="mt-16 grid md:grid-cols-3 gap-8 text-left relative">
                            <div className="hidden md:block absolute top-8 left-0 w-full h-px">
                                <svg className="w-full" preserveAspectRatio="none" fill="none"
                                     xmlns="http://www.w3.org/2000/svg">
                                    <path d="M0 1H1182" stroke="hsl(var(--border))" strokeWidth="2"
                                          strokeDasharray="8 8"/>
                                </svg>
                            </div>
                            <div className="relative z-10 p-6 bg-background">
                                <div className="flex items-center gap-4">
                                    <div
                                        className="w-10 h-10 flex-shrink-0 flex items-center justify-center font-bold text-lg bg-primary text-primary-foreground rounded-full">1
                                    </div>
                                    <h3 className="text-xl font-semibold">Paste URL</h3></div>
                                <p className="mt-3 text-muted-foreground">Copy your long, cumbersome URL and paste it
                                    into the input field above.</p>
                            </div>
                            <div className="relative z-10 p-6 bg-background">
                                <div className="flex items-center gap-4">
                                    <div
                                        className="w-10 h-10 flex-shrink-0 flex items-center justify-center font-bold text-lg bg-primary text-primary-foreground rounded-full">2
                                    </div>
                                    <h3 className="text-xl font-semibold">Shorten It</h3></div>
                                <p className="mt-3 text-muted-foreground">Click the button. Our system will generate a
                                    unique, short, and shareable link.</p>
                            </div>
                            <div className="relative z-10 p-6 bg-background">
                                <div className="flex items-center gap-4">
                                    <div
                                        className="w-10 h-10 flex-shrink-0 flex items-center justify-center font-bold text-lg bg-primary text-primary-foreground rounded-full">3
                                    </div>
                                    <h3 className="text-xl font-semibold">Share & Track</h3></div>
                                <p className="mt-3 text-muted-foreground">Copy your new link and share it. Log in to see
                                    real-time analytics.</p>
                            </div>
                        </div>
                    </div>
                </section>

                <section id="testimonials" className="py-20 sm:py-24 px-4 bg-secondary/50">
                    <div className="max-w-5xl mx-auto text-center">
                        <h2 className="text-3xl sm:text-4xl font-bold">Loved by Professionals Worldwide</h2>
                        <p className="mt-4 text-lg text-muted-foreground">Don't just take our word for it. See what our
                            users are saying.</p>
                        <div className="mt-16 grid lg:grid-cols-3 gap-8">
                            {testimonials.map((testimonial, index) => (
                                <Card key={index} className="flex flex-col">
                                    <CardContent className="pt-6 flex-grow"><p
                                        className="italic">"{testimonial.quote}"</p></CardContent>
                                    <CardHeader className="mt-auto pt-4 border-t">
                                        <p className="font-semibold text-foreground">{testimonial.name}</p>
                                        <p className="text-sm text-muted-foreground">{testimonial.title}</p>
                                    </CardHeader>
                                </Card>
                            ))}
                        </div>
                    </div>
                </section>

                <section id="pricing" className="py-20 sm:py-24 px-4 bg-primary text-primary-foreground">
                    <div className="max-w-3xl mx-auto text-center">
                        <h2 className="text-3xl sm:text-4xl font-extrabold">Ready to Supercharge Your Links?</h2>
                        <p className="mt-4 text-lg text-primary-foreground/80">Join thousands of creators, marketers,
                            and businesses who trust GoShort to build their brand and track their impact.</p>
                        <Link to="/register"
                              className={cn(buttonVariants({variant: 'secondary', size: 'lg'}), "mt-8 text-lg")}>Create
                            Your Free Account <ArrowRight className="w-5 h-5 ml-2"/></Link>
                    </div>
                </section>
            </main>

            <footer className="bg-muted text-muted-foreground py-12 px-4 sm:px-6 lg:px-8">
                <div className="max-w-7xl mx-auto grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-8">
                    <div className="col-span-2 md:col-span-4 lg:col-span-1">
                        <h3 className="text-xl font-bold text-foreground">GoShort</h3>
                        <p className="mt-2 text-sm">Short links, big results.</p>
                        <div className="mt-4 flex gap-2">
                            <a href="#" className={buttonVariants({variant: "ghost", size: "icon"})}><Twitter/></a>
                            <a href="#" className={buttonVariants({variant: "ghost", size: "icon"})}><Facebook/></a>
                            <a href="#" className={buttonVariants({variant: "ghost", size: "icon"})}><Linkedin/></a>
                        </div>
                    </div>
                    <div>
                        <h4 className="font-semibold text-foreground">Solutions</h4>
                        <ul className="mt-4 space-y-2 text-sm">
                            <li><a href="#" className="hover:text-foreground">Social Media</a></li>
                            <li><a href="#" className="hover:text-foreground">Digital Marketing</a></li>
                            <li><a href="#" className="hover:text-foreground">Developers</a></li>
                            <li><a href="#" className="hover:text-foreground">Branded Links</a></li>
                        </ul>
                    </div>
                    <div>
                        <h4 className="font-semibold text-foreground">Features</h4>
                        <ul className="mt-4 space-y-2 text-sm">
                            <li><a href="#" className="hover:text-foreground">Link Management</a></li>
                            <li><a href="#" className="hover:text-foreground">QR Codes</a></li>
                            <li><a href="#" className="hover:text-foreground">Analytics</a></li>
                        </ul>
                    </div>
                    <div>
                        <h4 className="font-semibold text-foreground">Resources</h4>
                        <ul className="mt-4 space-y-2 text-sm">
                            <li><a href="#" className="hover:text-foreground">Blog</a></li>
                            <li><a href="#" className="hover:text-foreground">Help Center</a></li>
                            <li><a href="#" className="hover:text-foreground">API Docs</a></li>
                        </ul>
                    </div>
                    <div>
                        <h4 className="font-semibold text-foreground">Company</h4>
                        <ul className="mt-4 space-y-2 text-sm">
                            <li><a href="#" className="hover:text-foreground">About Us</a></li>
                            <li><a href="#" className="hover:text-foreground">Careers</a></li>
                            <li><a href="#" className="hover:text-foreground">Contact</a></li>
                        </ul>
                    </div>
                </div>
                <div className="mt-8 pt-8 border-t text-center text-sm">
                    <p>&copy; {new Date().getFullYear()} GoShort. All rights reserved.</p>
                </div>
            </footer>

            <style>{`
                @keyframes fade-in-up {
                    0% {
                        opacity: 0;
                        transform: translateY(10px);
                    }
                    100% {
                        opacity: 1;
                        transform: translateY(0);
                    }
                }

                .animate-fade-in-up {
                    animation: fade-in-up 0.5s ease-out forwards;
                }

                .bg-grid-gray-200\\/40 {
                    background-image: linear-gradient(to right, var(--grid-color) 1px, transparent 1px),
                    linear-gradient(to bottom, var(--grid-color) 1px, transparent 1px);
                    background-size: 40px 40px;
                }
            `}</style>
        </div>
    );
}

