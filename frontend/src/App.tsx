import { useEffect, useState } from 'react'
import './App.css'

interface FileInfo {
  name: string
  size: number
  isDir: boolean
  modTime: string
  path: string
}

function App() {
  const [files, setFiles] = useState<FileInfo[]>([])
  const [currentPath, setCurrentPath] = useState('.')
  const [error, setError] = useState('')

  const fetchFiles = async (path: string) => {
    try {
      const response = await fetch(`/api/files?path=${encodeURIComponent(path)}`)
      if (!response.ok) {
        throw new Error('Failed to fetch files')
      }
      const data = await response.json()
      setFiles(data)
      setCurrentPath(path)
      setError('')
    } catch (err) {
      setError('Error fetching files: ' + (err as Error).message)
    }
  }

  useEffect(() => {
    fetchFiles(currentPath)
  }, [])

  const handleFileClick = (file: FileInfo) => {
    if (file.isDir) {
      fetchFiles(file.path)
    }
  }

  const handleBackClick = () => {
    if (currentPath === '.') return
    const parentPath = currentPath.split('/').slice(0, -1).join('/')
    fetchFiles(parentPath || '.')
  }

  const formatSize = (size: number) => {
    if (size < 1024) return `${size} B`
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
    if (size < 1024 * 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`
    return `${(size / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }

  return (
    <div className="max-w-6xl mx-auto p-8 text-gray-100">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <button
            onClick={handleBackClick}
            disabled={currentPath === '.'}
            className="px-4 py-2 bg-gray-800 text-gray-100 rounded-md cursor-pointer transition-colors duration-200 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Back
          </button>
          <span className="text-gray-400 text-sm">{currentPath === '.' ? 'Root' : currentPath}</span>
        </div>
        <div className="text-sm text-gray-400">Directory Size: {files.reduce((acc, file) => acc + file.size, 0) / (1024 * 1024 * 1024)} GB</div>
      </div>
      {error && <div className="text-red-400 p-4 bg-red-900/50 rounded-md mb-4">{error}</div>}
      <div className="flex flex-col gap-2 bg-gray-800/50 rounded-lg p-4">
        {files.map((file) => (
          <div
            key={file.path}
            className={`flex items-center p-3 rounded-md cursor-pointer transition-colors duration-200 hover:bg-gray-700/50 ${file.isDir ? 'bg-gray-800/80' : 'bg-gray-800/40'}`}
            onClick={() => handleFileClick(file)}
          >
            <span className="mr-3 text-lg">{file.isDir ? '📁' : '📄'}</span>
            <span className="flex-1 font-medium">{file.name}</span>
            <div className="flex gap-4 text-gray-400 text-sm items-center">
              {!file.isDir && <span className="min-w-[80px]">{formatSize(file.size)}</span>}
              <span className="text-gray-500">{file.modTime}</span>
              <div className="flex gap-2">
                <button className="p-2 hover:bg-blue-500/20 rounded-md">
                  <span className="text-blue-400">⬇️</span>
                </button>
                <button className="p-2 hover:bg-gray-600/20 rounded-md">
                  <span>🔗</span>
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default App
