import React, { useState, useRef, useEffect } from 'react';
import { Play, Maximize2, Volume2, Settings, Check, Activity } from 'lucide-react';
import Hls from 'hls.js';

interface VideoPlayerProps {
    src?: string;
    poster?: string;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({ src, poster }) => {
    const [selectedQuality, setSelectedQuality] = useState('Auto');
    const [availableQualities, setAvailableQualities] = useState<string[]>(['Auto']);
    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const videoRef = useRef<HTMLVideoElement>(null);
    const hlsRef = useRef<Hls | null>(null);
    const menuRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!src || !videoRef.current) return;

        const video = videoRef.current;

        // Clean up previous HLS instance
        if (hlsRef.current) {
            hlsRef.current.destroy();
        }

        if (Hls.isSupported()) {
            const hls = new Hls({
                enableWorker: true,
                lowLatencyMode: true,
            });
            hlsRef.current = hls;

            hls.loadSource(src);
            hls.attachMedia(video);

            hls.on(Hls.Events.MANIFEST_PARSED, () => {
                const levels = hls.levels;
                const qualities = ['Auto', ...levels.map(l => `${l.height}p`)];
                // Reverse to show highest quality first
                setAvailableQualities(qualities);
                video.play().catch(() => {
                    // Autoplay might be blocked
                });
            });

            hls.on(Hls.Events.LEVEL_SWITCHED, (_, data) => {
                if (hls.autoLevelEnabled) {
                    const currentLevel = hls.levels[data.level];
                    if (currentLevel) {
                        // We don't necessarily want to change the text if it's "Auto"
                    }
                }
            });

            hls.on(Hls.Events.ERROR, (_, data) => {
                if (data.fatal) {
                    switch (data.type) {
                        case Hls.ErrorTypes.NETWORK_ERROR:
                            hls.startLoad();
                            break;
                        case Hls.ErrorTypes.MEDIA_ERROR:
                            hls.recoverMediaError();
                            break;
                        default:
                            hls.destroy();
                            break;
                    }
                }
            });
        }
        // Native HLS support (Safari)
        else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = src;
            video.addEventListener('loadedmetadata', () => {
                video.play().catch(() => { });
            });
        }

        return () => {
            if (hlsRef.current) {
                hlsRef.current.destroy();
            }
        };
    }, [src]);

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

        if (!hlsRef.current) return;

        const hls = hlsRef.current;
        if (quality === 'Auto') {
            hls.currentLevel = -1; // -1 means auto
        } else {
            const height = parseInt(quality);
            const levelIndex = hls.levels.findIndex(l => l.height === height);
            if (levelIndex !== -1) {
                hls.currentLevel = levelIndex;
            }
        }
    };

    return (
        <div className="relative group w-full aspect-video bg-black rounded-2xl overflow-hidden shadow-2xl transition-all duration-300 hover:shadow-[0_20px_50px_rgba(0,0,0,0.5)] border border-white/10">
            {src ? (
                <div className="relative w-full h-full">
                    <video
                        ref={videoRef}
                        className="w-full h-full object-contain"
                        poster={poster}
                        controls
                        playsInline
                    />

                    {/* Quality Selector Overlay */}
                    <div className="absolute bottom-16 right-4 flex flex-col items-end z-20">
                        {isMenuOpen && (
                            <div
                                ref={menuRef}
                                className="mb-2 w-44 bg-black/90 backdrop-blur-xl border border-white/10 rounded-xl overflow-hidden shadow-2xl animate-in fade-in slide-in-from-bottom-2 duration-200"
                            >
                                <div className="px-4 py-3 border-b border-white/5 flex items-center justify-between">
                                    <span className="text-[10px] font-bold text-gray-500 uppercase tracking-widest flex items-center gap-2">
                                        <Activity size={10} /> Bitrate Quality
                                    </span>
                                </div>
                                <div className="max-h-60 overflow-y-auto">
                                    {availableQualities.map((q) => (
                                        <button
                                            key={q}
                                            onClick={() => handleQualityChange(q)}
                                            className="w-full px-4 py-3 flex items-center justify-between text-xs transition-colors hover:bg-white/10 text-left border-b border-white/5 last:border-0"
                                        >
                                            <div className="flex flex-col">
                                                <span className={selectedQuality === q ? 'text-blue-400 font-semibold' : 'text-gray-300'}>
                                                    {q}
                                                </span>
                                                {q === 'Auto' && <span className="text-[9px] text-gray-500">Adaptive Switching</span>}
                                            </div>
                                            {selectedQuality === q && (
                                                <div className="w-5 h-5 rounded-full bg-blue-500/20 flex items-center justify-center">
                                                    <Check size={12} className="text-blue-400" />
                                                </div>
                                            )}
                                        </button>
                                    ))}
                                </div>
                            </div>
                        )}

                        <button
                            onClick={() => setIsMenuOpen(!isMenuOpen)}
                            className={`w-10 h-10 rounded-xl backdrop-blur-md border flex items-center justify-center transition-all duration-300 active:scale-90 ${isMenuOpen
                                ? 'bg-blue-500 border-blue-400 text-white shadow-[0_0_20px_rgba(59,130,246,0.5)]'
                                : 'bg-black/40 border-white/10 text-white hover:bg-white/20'
                                }`}
                            title="Streaming Quality"
                        >
                            <Settings size={20} className={`transition-transform duration-700 ${isMenuOpen ? 'rotate-180' : ''}`} />
                        </button>
                    </div>
                </div>
            ) : (
                <div className="absolute inset-0 flex flex-col items-center justify-center bg-linear-to-br from-gray-900 to-black text-white p-6 text-center">
                    <div className="w-20 h-20 bg-white/5 rounded-full flex items-center justify-center mb-6 backdrop-blur-sm border border-white/10 group-hover:scale-110 transition-transform duration-500">
                        <Play size={40} className="text-white fill-white ml-1" />
                    </div>
                    <h3 className="text-2xl font-semibold mb-2 tracking-tight text-white">No Stream Initialized</h3>
                    <p className="text-gray-400 max-w-sm text-sm">
                        Upload a video to begin the multi-bitrate transcoding process and launch your adaptive stream.
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
