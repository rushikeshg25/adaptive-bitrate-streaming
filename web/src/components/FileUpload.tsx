import React, { useState, useRef, useEffect } from 'react';
import axios from 'axios';
import { Upload, FileVideo, CheckCircle2, Loader2 } from 'lucide-react';


interface FileUploadProps {
    onUploadSuccess?: (url: string) => void;
}

const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB

const FileUpload: React.FC<FileUploadProps> = ({ onUploadSuccess }) => {
    const request = useRef<AbortController | null>(null);
    useEffect(() => () => request.current?.abort(), []);
    const [isDragging, setIsDragging] = useState(false);
    const [isUploading, setIsUploading] = useState(false);
    const [isSuccess, setIsSuccess] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [progress, setProgress] = useState(0);
    const [isProcessing, setIsProcessing] = useState(false);

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
        if (request.current) return;
        if (file.size === 0 || file.size > CHUNK_SIZE * 200) { setError("Choose a video between 1 byte and 1000 MiB."); return; }
        const controller = new AbortController(); request.current = controller;
        setIsUploading(true);
        setError(null);
        setIsSuccess(false);
        setProgress(0);

        const uploadId = crypto.randomUUID();
        const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

        try {
            for (let i = 0; i < totalChunks; i++) {
                const start = i * CHUNK_SIZE;
                const end = Math.min(start + CHUNK_SIZE, file.size);
                const chunk = file.slice(start, end);

                const formData = new FormData();
                formData.append('chunk', chunk);
                formData.append('uploadId', uploadId);
                formData.append('index', i.toString());

                await axios.post('http://localhost:3000/api/upload/chunk', formData, {
                    signal: controller.signal,
                    timeout: 60000,
                    onUploadProgress: (progressEvent) => {
                        const chunkProgress = progressEvent.loaded / (progressEvent.total || (end - start));
                        const totalProgress = Math.round(((i + chunkProgress) / totalChunks) * 100);
                        setProgress(totalProgress);
                    },
                });
            }

            // Start reassembly and background transcoding
            setIsProcessing(true);
            const response = await completeWithRetry({
                uploadId,
                filename: file.name,
                total: totalChunks,
            }, controller.signal);

            const { videoID, url } = response.data;

            // Poll for transcoding completion
            await pollForCompletion(videoID, controller.signal);

            setIsSuccess(true);
            if (onUploadSuccess && url) {
                onUploadSuccess(`http://localhost:3000${url}`);
            }

            
        } catch (err) {
            if (!controller.signal.aborted) setError(axios.isAxiosError(err) ? (typeof err.response?.data === 'string' ? err.response.data : err.response?.data?.error) || err.message : err instanceof Error ? err.message : 'Upload failed');
        } finally {
            request.current = null;
            setIsUploading(false);
            setIsProcessing(false);
        }
    };

    const completeWithRetry = async (payload: { uploadId: string; filename: string; total: number }, signal: AbortSignal) => {
        const deadline = Date.now() + 180000;
        while (Date.now() < deadline) {
            try {
                return await axios.post<{ videoID: string; url: string }>('http://localhost:3000/api/upload/complete', payload, { signal, timeout: Math.max(1, Math.min(30000, deadline - Date.now())) });
            } catch (err) {
                if (!axios.isAxiosError(err) || err.response?.status !== 503 || signal.aborted) throw err;
                await new Promise(resolve => setTimeout(resolve, Math.max(0, Math.min(2000, deadline - Date.now()))));
            }
        }
        throw new Error('The processor is busy. Please try again later.');
    };

    const pollForCompletion = async (videoID: string, signal: AbortSignal) => {
        const deadline = Date.now() + 180000;
        while (Date.now() < deadline) {
            const response = await axios.get<{ id: string; status: string }[]>('http://localhost:3000/api/videos', { signal, timeout: Math.max(1, Math.min(15000, deadline - Date.now())) });
            const video = response.data.find(v => v.id === videoID);
            if (video?.status === 'Completed') return;
            if (video?.status === 'Failed') throw new Error('The video could not be processed. Try another video.');
            await new Promise(resolve => setTimeout(resolve, Math.max(0, Math.min(2000, deadline - Date.now()))));
        }
        throw new Error('Processing is taking too long. Please check the video list later.');
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
                        <div className="w-full px-8">
                            <div className="relative w-full h-2 bg-white/10 rounded-full overflow-hidden mb-4">
                                <div
                                    className="absolute top-0 left-0 h-full bg-blue-500 transition-all duration-300 ease-out"
                                    style={{ width: `${progress}%` }}
                                />
                            </div>
                            <div className="flex justify-between items-center mb-1">
                                <h4 className="text-xl font-medium text-white">
                                    {isProcessing ? 'Processing Video...' : 'Uploading...'}
                                </h4>
                                {!isProcessing && <span className="text-blue-500 font-bold">{progress}%</span>}
                                {isProcessing && <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />}
                            </div>
                            <p className="text-gray-400 text-sm">
                                {isProcessing
                                    ? 'Transcoding to multiple qualities (ABR)...'
                                    : 'Transferring video chunks to server'}
                            </p>
                        </div>
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
