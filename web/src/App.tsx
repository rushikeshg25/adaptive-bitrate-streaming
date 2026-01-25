import VideoPlayer from './components/VideoPlayer';
import FileUpload from './components/FileUpload';
import { MonitorPlay, Shield } from 'lucide-react';

function App() {
  return (
    <div className="max-w-6xl mx-auto px-6 py-12 lg:py-20">
      {/* Header */}
      <header className="text-center mb-16 space-y-4">
        <h1 className="text-5xl lg:text-7xl font-bold tracking-tight bg-linear-to-b from-white to-white/60 bg-clip-text text-transparent">
          Adaptive Bitrate <br /> Streaming
        </h1>
      </header>

      {/* Main Content Area */}
      <main className="space-y-24">
        <section>
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-xl bg-white/5 flex items-center justify-center border border-white/10 shadow-inner">
              <MonitorPlay size={20} className="text-blue-400" />
            </div>
            <h2 className="text-2xl font-semibold">Live Preview</h2>
          </div>
          <VideoPlayer />
        </section>

        <section>
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-xl bg-white/5 flex items-center justify-center border border-white/10 shadow-inner">
              <Shield size={20} className="text-blue-400" />
            </div>
            <h2 className="text-2xl font-semibold">Secure Upload</h2>
          </div>
          <FileUpload />
        </section>
      </main>
    </div>
  )
}

export default App