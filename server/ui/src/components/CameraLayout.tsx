import { Camera } from "@/types/camera";
import { VideoStream } from "./VideoStream";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import useUser from "@/hooks/useUser";

interface CameraLayoutProps {
  cameras: Camera[];
  user_id: string;
  onSelectCamera?: (cameraUUID: string) => void;
}

export default function CameraLayout({ cameras, onSelectCamera,user_id }: CameraLayoutProps) {
  
  // Sort cameras by online status (online first)
  const sortedCameras = [...cameras].sort((a, b) => {
    if (a.onlineStatus === b.onlineStatus) return 0;
    return a.onlineStatus ? -1 : 1;
  });

  if (cameras.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">No cameras available</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {sortedCameras.map((camera) => (
        <Card 
          key={camera.cameraUUID} 
          className={`overflow-hidden ${!camera.onlineStatus ? 'opacity-70' : ''} hover:shadow-md transition-shadow cursor-pointer`}
          onClick={() => onSelectCamera && onSelectCamera(camera.cameraUUID)}
        >
          <CardHeader className="p-3">
            <div className="flex justify-between items-center">
              <CardTitle className="text-sm font-medium">Camera {camera.cameraUUID.substring(0, 8)}</CardTitle>
              <Badge variant={camera.onlineStatus ? "default" : "destructive"}>
                {camera.onlineStatus ? "Online" : "Offline"}
              </Badge>
            </div>
          </CardHeader>
          <CardContent className="p-0 aspect-video bg-muted flex items-center justify-center">
            {camera.onlineStatus ? (
              <VideoStream showStats={false} user_uuid={user_id} camera_uuid={camera.cameraUUID} />
            ) : (
              <div className="text-center text-muted-foreground">
                <p>Camera offline</p>
                <p className="text-xs mt-2">
                  {camera.lastOnline ? `Last online: ${new Date(camera.lastOnline).toLocaleString()}` : 'Never connected'}
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}