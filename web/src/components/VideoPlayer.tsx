import React from 'react';
import { Play, Maximize2, Volume2 } from 'lucide-react';

interface VideoPlayerProps {
    src?: string;
    poster?: string;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({ src, poster }) => {
    return (
        <div className="relative group w-full aspect-video bg-black rounded-2xl overflow-hidden shadow-2xl transition-all duration-300 hover:shadow-[0_20px_50px_rgba(0,0,0,0.5)] border border-white/10">
            {src ? (
                <video
                    className="w-full h-full object-contain"
                    src={src}
                    poster={poster}
                    controls
                />
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
