import { Camera } from "@/types/camera";
import { VideoStream } from "./VideoStream";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";

interface CameraLayoutProps {
  cameras: Camera[];
  user_id: string;
  onSelectCamera?: (id: string) => void;
}

export default function CameraLayout({ cameras, onSelectCamera,user_id }: CameraLayoutProps) {
  
  // Sort cameras by online status (online first)
  const sortedCameras = [...cameras].sort((a, b) => {
    if (a.isOnline === b.isOnline) return 0;
    return a.isOnline ? -1 : 1;
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
          key={camera.id} 
          className={`overflow-hidden ${!camera.isOnline ? 'opacity-70' : ''} shadow-md transition-shadow cursor-pointer`}
          onClick={() => onSelectCamera && onSelectCamera(camera.id)}
        >
          <CardHeader className="p-3">
            <div className="flex justify-between items-center">
              <CardTitle className="text-sm font-medium">{camera.name}</CardTitle>
              <Badge variant={camera.isOnline ? "default" : "destructive"}>
                {camera.isOnline ? "Online" : "Offline"}
              </Badge>
            </div>
          </CardHeader>
          <CardContent className="p-0 aspect-video bg-muted flex items-center justify-center">
            {camera.isOnline ? (
              <VideoStream showStats={false} camera_uuid={camera.id} />
            ) : (
              <div className="text-center text-muted-foreground">
                <p>Camera offline</p>
                <p className="text-xs mt-2">
                  {camera.lastSeen ? `Last online: ${new Date(camera.lastSeen!).toLocaleString()}` : 'Never connected'}
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}