import { VideoStream } from "@/components/VideoStream";
import { CameraStatus } from "@/components/CameraStatus";
import { useEffect, useState } from "react";
import { Camera } from "@/types/camera";
import useUser from "@/hooks/useUser";

export default function HomePage() {
    const {user,} = useUser();
    const [cameras, setCameras] = useState<Camera[]>([]);
    const [selectedCamera, setSelectedCamera] = useState<string>("");
    const [isLoading, setIsLoading] = useState(false);

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
                if (data.length > 0) {
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
    }, []);

    const selectedCameraObj = cameras.find(cam => cam.cameraUUID === selectedCamera);

    return (
        <div className="p-4">
            <div className="mb-4 flex flex-col gap-2">
                <select 
                    value={selectedCamera}
                    onChange={(e) => setSelectedCamera(e.target.value)}
                    className="border rounded p-2"
                >
                    {cameras.map((camera) => (
                        <option key={camera.cameraUUID} value={camera.cameraUUID}>
                            Camera {camera.cameraUUID}
                        </option>
                    ))}
                </select>
                
                {selectedCameraObj && (
                    <CameraStatus camera={selectedCameraObj} />
                )}
            </div>
            
            {selectedCamera && <VideoStream user_uuid={user!.userID} camera_uuid={selectedCamera} />}
            {isLoading && <div>Loading cameras...</div>}
        </div>
    );
}