import React, { useState, useRef, useEffect } from 'react';
import { Play, Maximize2, Volume2, Settings, Check } from 'lucide-react';

interface VideoPlayerProps {
    src?: string;
    poster?: string;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({ src, poster }) => {
    const [selectedQuality, setSelectedQuality] = useState('Auto');
    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const videoRef = useRef<HTMLVideoElement>(null);
    const menuRef = useRef<HTMLDivElement>(null);

    const qualities = ['1080p', '720p', '480p', '360p', 'Auto'];

    // Close menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setIsMenuOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleQualityChange = (quality: string) => {
        setSelectedQuality(quality);
        setIsMenuOpen(false);
        // In a real HLS implementation, we would tell the HLS instance to switch levels
    };

    return (
        <div className="relative group w-full aspect-video bg-black rounded-2xl overflow-hidden shadow-2xl transition-all duration-300 hover:shadow-[0_20px_50px_rgba(0,0,0,0.5)] border border-white/10">
            {src ? (
                <div className="relative w-full h-full">
                    <video
                        ref={videoRef}
                        className="w-full h-full object-contain"
                        src={src}
                        poster={poster}
                        controls
                    />

                    {/* Quality Selector Overlay */}
                    <div className="absolute bottom-16 right-4 flex flex-col items-end z-20">
                        {isMenuOpen && (
                            <div
                                ref={menuRef}
                                className="mb-2 w-40 bg-black/90 backdrop-blur-xl border border-white/10 rounded-xl overflow-hidden shadow-2xl animate-in fade-in slide-in-from-bottom-2 duration-200"
                            >
                                <div className="px-4 py-2 border-b border-white/5 text-[10px] font-bold text-gray-500 uppercase tracking-widest">
                                    Quality
                                </div>
                                {qualities.map((q) => (
                                    <button
                                        key={q}
                                        onClick={() => handleQualityChange(q)}
                                        className="w-full px-4 py-2.5 flex items-center justify-between text-xs transition-colors hover:bg-white/10 text-left"
                                    >
                                        <span className={selectedQuality === q ? 'text-blue-400 font-medium' : 'text-gray-300'}>
                                            {q}
                                        </span>
                                        {selectedQuality === q && <Check size={12} className="text-blue-400" />}
                                    </button>
                                ))}
                            </div>
                        )}

                        <button
                            onClick={() => setIsMenuOpen(!isMenuOpen)}
                            className="w-8 h-8 rounded-full bg-black/40 backdrop-blur-md border border-white/10 flex items-center justify-center text-white hover:bg-white/20 transition-all hover:scale-110 active:scale-95"
                            title="Settings"
                        >
                            <Settings size={16} className={`transition-transform duration-500 ${isMenuOpen ? 'rotate-90' : ''}`} />
                        </button>
                    </div>
                </div>
            ) : (
                <div className="absolute inset-0 flex flex-col items-center justify-center bg-linear-to-br from-gray-900 to-black text-white p-6 text-center">
                    <div className="w-20 h-20 bg-white/5 rounded-full flex items-center justify-center mb-6 backdrop-blur-sm border border-white/10 group-hover:scale-110 transition-transform duration-500">
                        <Play size={40} className="text-white fill-white ml-1" />
                    </div>
                    <h3 className="text-2xl font-semibold mb-2 tracking-tight">No Video Selected</h3>
                    <p className="text-gray-400 max-w-sm">
                        Upload a video below or select a stream to start watching in high quality.
                    </p>

                    <div className="absolute bottom-6 left-6 right-6 flex items-center justify-between opacity-40 group-hover:opacity-100 transition-opacity duration-300">
                        <div className="flex gap-4">
                            <div className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center"><Play size={16} /></div>
                            <div className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center"><Volume2 size={16} /></div>
                        </div>
                        <div className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center"><Maximize2 size={16} /></div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default VideoPlayer;
