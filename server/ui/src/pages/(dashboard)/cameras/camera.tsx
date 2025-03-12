import { VideoStream } from "@/components/VideoStream";
import { useAppContext } from "@/context/AppContext";
import { Link, useParams } from "react-router";
import { CameraSettings } from "@/components/camera-settings";
import { Button } from "@/components/ui/button";
import { Settings } from "lucide-react";

export function CameraPage() {
    const { id } = useParams();
    const { cameras, user } = useAppContext();
    const camera = cameras.find((c) => c.id === id);

    if (!camera || !user?.id) {
        return <div>Camera not found</div>;
    }

    return (
        <div className="container mx-auto p-4">
            <div className="flex justify-between items-center mb-4">
                <h1 className="text-2xl font-bold">{camera.name}</h1>
                <Link to={`/recordings/${camera.id}`} className="text-blue-500">Recordings</Link>
                <CameraSettings id={camera.id}>
                    <Button variant="outline" size="sm">
                        <Settings className="h-4 w-4 mr-2" /> Settings
                    </Button>
                </CameraSettings>
            </div>
            <div className="relative aspect-w-16 aspect-h-9 flex-col justify-center items-center">
                {camera.isOnline ? (
                <VideoStream showStats={false}  camera_uuid={camera.id} />
                ) : (
                <div className="text-center text-muted-foreground">
                    <p>Camera offline</p>
                    <p className="text-xs mt-2">
                    {camera.lastSeen ? `Last online: ${new Date(camera.lastSeen!).toLocaleString()}` : 'Never connected'}
                    </p>
                </div>
                )}
            </div>
        </div>
    );
}