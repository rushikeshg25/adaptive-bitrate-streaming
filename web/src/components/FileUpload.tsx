import React, { useState } from 'react';
import { Upload, FileVideo, CheckCircle2, Loader2 } from 'lucide-react';

const FileUpload: React.FC = () => {
    const [isDragging, setIsDragging] = useState(false);
    const [isUploading, setIsUploading] = useState(false);
    const [isSuccess, setIsSuccess] = useState(false);

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = () => {
        setIsDragging(false);
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
        // Handle file drop logic here
        simulateUpload();
    };

    const simulateUpload = () => {
        setIsUploading(true);
        setTimeout(() => {
            setIsUploading(false);
            setIsSuccess(true);
            setTimeout(() => setIsSuccess(false), 3000);
        }, 2000);
    };

    return (
        <div className="w-full max-w-3xl mx-auto mt-12">
            <div
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                className={`relative group flex flex-col items-center justify-center w-full h-64 border-2 border-dashed rounded-3xl transition-all duration-300 ${isDragging
                        ? 'border-blue-500 bg-blue-500/5'
                        : 'border-white/10 bg-white/5 hover:border-white/20 hover:bg-white/[0.07]'
                    } backdrop-blur-xl overflow-hidden`}
            >
                <input
                    type="file"
                    className="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
                    onChange={() => simulateUpload()}
                    accept="video/*"
                />

                <div className="flex flex-col items-center text-center p-8">
                    {isUploading ? (
                        <>
                            <Loader2 className="w-16 h-16 text-blue-500 animate-spin mb-4" />
                            <h4 className="text-xl font-medium text-white mb-1">Processing Video...</h4>
                            <p className="text-gray-400">Optimizing for adaptive bitrate streaming</p>
                        </>
                    ) : isSuccess ? (
                        <>
                            <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mb-4">
                                <CheckCircle2 className="w-10 h-10 text-green-500" />
                            </div>
                            <h4 className="text-xl font-medium text-white mb-1">Upload Complete!</h4>
                            <p className="text-gray-400">Your video is ready for processing</p>
                        </>
                    ) : (
                        <>
                            <div className={`w-16 h-16 rounded-2xl flex items-center justify-center mb-4 transition-all duration-300 ${isDragging ? 'bg-blue-500 scale-110' : 'bg-white/10 group-hover:bg-white/15'
                                }`}>
                                {isDragging ? <FileVideo className="text-white" size={32} /> : <Upload className="text-gray-400 group-hover:text-white" size={32} />}
                            </div>
                            <h4 className="text-xl font-medium text-white mb-2">
                                Click or drag video to upload
                            </h4>
                            <p className="text-gray-400 text-sm max-w-xs">
                                Supports MP4, MOV, and AVI formats. High-quality bitrate transcoding will start automatically.
                            </p>
                        </>
                    )}
                </div>

                {/* Decorative elements */}
                <div className="absolute top-0 right-0 p-4 opacity-10">
                    <FileVideo size={120} />
                </div>
            </div>
        </div>
    );
};

export default FileUpload;
