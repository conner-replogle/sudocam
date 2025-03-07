import { useAppContext } from "@/context/AppContext";
import { Link, useNavigate } from "react-router";
import { CameraSettings } from "@/components/camera-settings";
import { Button } from "@/components/ui/button";
import { Settings } from "lucide-react";

export function CamerasListPage(){
    const { cameras, user } = useAppContext();
    const navigate = useNavigate();
    
    // Helper function to truncate UUID
    const truncateUUID = (uuid: string) => {
        if (!uuid) return '';
        // Show first 8 and last 4 characters with ellipsis in the middle
        return `${uuid.substring(0, 8)}...${uuid.substring(uuid.length - 4)}`;
    };

    return (
        <div className="p-4 container mx-auto">
            <div className="mb-6 flex justify-between items-center">
                <h1 className="text-2xl font-bold">My Cameras</h1>
                {cameras.length > 0 && (
                    <Link className="bg-primary text-primary-foreground px-4 py-2 rounded" to={`/cameras/add`}>
                        Add Camera
                    </Link>
                )}
            </div>

            {cameras.length === 0 || !user?.id ? (
                <div className="text-center py-12">
                    <h2 className="text-xl font-semibold mb-2">No cameras found</h2>
                    <p className="text-muted-foreground mb-4">Add your first camera to get started</p>
                    <Link className="bg-primary text-primary-foreground px-4 py-2 rounded" to={`/cameras/add`}>
                            Add Camera
                    </Link>
                </div>
            ) : (
                <div className="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
                    {cameras.map((camera) => (
                        <div key={camera.cameraUUID} className="bg-primary-foreground shadow-md rounded-lg p-4 flex flex-row">
                            <div className="mb-4 flex-grow">
                                <h2 className="text-lg font-semibold text-foreground truncate" title={camera.name}>{camera.name}</h2>
                                <p className="text-muted-foreground text-sm font-mono" title={camera.cameraUUID}>
                                    {truncateUUID(camera.cameraUUID)}
                                </p>
                                <div className="flex items-center space-x-2 mt-1">
                                    <div className={`w-2 h-2 rounded-full ${camera.isOnline ? 'bg-green-500' : 'bg-red-500'}`}></div>
                                    <span className="text-xs text-muted-foreground">{camera.isOnline ? 'Online' : 'Offline'}</span>
                                </div>
                            </div>
                            <div className="flex space-x-2 items-center justify-end">
                                <CameraSettings cameraUUID={camera.cameraUUID}>
                                    <Button variant="outline" size="icon">
                                        <Settings className="h-4 w-4" />
                                    </Button>
                                </CameraSettings>
                                <Button 
                                    onClick={() => navigate(`/cameras/${camera.cameraUUID}`)} 
                                    className="bg-primary text-primary-foreground"
                                >
                                    View
                                </Button>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}