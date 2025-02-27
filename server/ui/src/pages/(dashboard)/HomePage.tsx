import { VideoStream } from "@/components/VideoStream";
import { useEffect, useState } from "react";
import { Camera } from "@/types/camera";
import useUser from "@/hooks/useUser";
import CameraLayout from "@/components/CameraLayout";

export default function HomePage() {
    const { user } = useUser();
    const [cameras, setCameras] = useState<Camera[]>([]);
    const [selectedCamera, setSelectedCamera] = useState<string>("");
    const [isLoading, setIsLoading] = useState(false);
    const [viewMode, setViewMode] = useState<'grid' | 'single'>('grid');

    useEffect(() => {
        const fetchCameras = async () => {
            setIsLoading(true);
            try {
                const token = localStorage.getItem('token')
                const response = await fetch('/api/users/cameras',{
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });
                if (!response.ok) throw new Error('Failed to fetch cameras');
                const data = await response.json();
                setCameras(data);
                if (data.length > 0 && !selectedCamera) {
                    setSelectedCamera(data[0].cameraUUID);
                }
            } catch (error) {
                console.error('Error fetching cameras:', error);
            } finally {
                setIsLoading(false);
            }
        };

        fetchCameras();
        
        // Poll for camera status updates every 30 seconds
        const intervalId = setInterval(fetchCameras, 30000);
        
        return () => clearInterval(intervalId);
    }, [selectedCamera]);

    const handleSelectCamera = (cameraUUID: string) => {
        setSelectedCamera(cameraUUID);
        setViewMode('single');
    };

    if (isLoading && cameras.length === 0 ) {
        return <div className="flex items-center justify-center h-64">Loading cameras...</div>;
    }

    return (
        <div className="p-4 container mx-auto">
            <div className="mb-6 flex justify-between items-center">
                <h1 className="text-2xl font-bold">My Cameras</h1>
                <div className="flex gap-2">
                    <button 
                        onClick={() => setViewMode('grid')} 
                        className={`px-3 py-1 rounded ${viewMode === 'grid' ? 'bg-primary text-primary-foreground' : 'bg-secondary'}`}
                    >
                        Grid View
                    </button>
                    <button 
                        onClick={() => setViewMode('single')} 
                        className={`px-3 py-1 rounded ${viewMode === 'single' ? 'bg-primary text-primary-foreground' : 'bg-secondary'}`}
                    >
                        Single View
                    </button>
                </div>
            </div>

            {viewMode === 'grid' && user?.userID ? (
                <CameraLayout user_id={user.userID} cameras={cameras} onSelectCamera={handleSelectCamera} />
            ) : (
                <div className="space-y-4">
                    {selectedCamera && user?.userID && (
                        <VideoStream showStats={true} user_uuid={user.userID} camera_uuid={selectedCamera} />
                    )}
                </div>
            )}
        </div>
    );
}