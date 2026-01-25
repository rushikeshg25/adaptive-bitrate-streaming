import React, { useState } from 'react';
import { Upload, FileVideo, CheckCircle2, Loader2 } from 'lucide-react';

interface FileUploadProps {
    onUploadSuccess?: (url: string) => void;
}

const FileUpload: React.FC<FileUploadProps> = ({ onUploadSuccess }) => {
    const [isDragging, setIsDragging] = useState(false);
    const [isUploading, setIsUploading] = useState(false);
    const [isSuccess, setIsSuccess] = useState(false);
    const [error, setError] = useState<string | null>(null);

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
        const file = e.dataTransfer.files[0];
        if (file && file.type.startsWith('video/')) {
            handleFileUpload(file);
        } else {
            setError('Please drop a valid video file.');
        }
    };

    const handleFileUpload = async (file: File) => {
        setIsUploading(true);
        setError(null);
        setIsSuccess(false);

        const formData = new FormData();
        formData.append('video', file);

        try {
            const response = await fetch('http://localhost:3000/upload', {
                method: 'POST',
                body: formData,
            });

            if (!response.ok) {
                throw new Error('Upload failed');
            }

            const data = await response.json();
            setIsSuccess(true);

            if (onUploadSuccess && data.url) {
                // The URL is relative from the backend, so we need to provide the full path
                onUploadSuccess(`http://localhost:3000${data.url}`);
            }

            setTimeout(() => setIsSuccess(false), 3000);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An unknown error occurred');
        } finally {
            setIsUploading(false);
        }
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
                    } backdrop-blur-xl overflow-hidden ${error ? 'border-red-500/50 bg-red-500/5' : ''}`}
            >
                <input
                    type="file"
                    className="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
                    onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) handleFileUpload(file);
                    }}
                    accept="video/*"
                />

                <div className="flex flex-col items-center text-center p-8">
                    {isUploading ? (
                        <>
                            <Loader2 className="w-16 h-16 text-blue-500 animate-spin mb-4" />
                            <h4 className="text-xl font-medium text-white mb-1">Uploading Video...</h4>
                            <p className="text-gray-400">Processing on server...</p>
                        </>
                    ) : isSuccess ? (
                        <>
                            <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mb-4">
                                <CheckCircle2 className="w-10 h-10 text-green-500" />
                            </div>
                            <h4 className="text-xl font-medium text-white mb-1">Upload Complete!</h4>
                            <p className="text-gray-400">Your video is ready to play</p>
                        </>
                    ) : (
                        <>
                            <div className={`w-16 h-16 rounded-2xl flex items-center justify-center mb-4 transition-all duration-300 ${isDragging ? 'bg-blue-500 scale-110' : 'bg-white/10 group-hover:bg-white/15'
                                } ${error ? 'bg-red-500/20' : ''}`}>
                                {isDragging ? <FileVideo className="text-white" size={32} /> : <Upload className="text-gray-400 group-hover:text-white" size={32} />}
                            </div>
                            <h4 className="text-xl font-medium text-white mb-2">
                                {error || 'Click or drag video to upload'}
                            </h4>
                            <p className="text-gray-400 text-sm max-w-xs">
                                Supports MP4, MOV, and AVI formats.
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
