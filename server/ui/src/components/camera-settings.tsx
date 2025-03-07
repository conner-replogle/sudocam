import { useAppContext } from "@/context/AppContext";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useState } from "react";
import { useNavigate } from "react-router";
import { Trash2 } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { apiDelete, apiPut } from "@/lib/api";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

interface CameraSettingsProps {
  cameraUUID: string;
  children: React.ReactNode;
}

export function CameraSettings({ cameraUUID, children }: CameraSettingsProps) {
  const { cameras, refetchCameras } = useAppContext();
  const navigate = useNavigate();
  const camera = cameras.find((c) => c.cameraUUID === cameraUUID);
  const [name, setName] = useState(camera?.name || "");
  const [isDeleting, setIsDeleting] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [open, setOpen] = useState(false);
  
  if (!camera) {
    return null;
  }

  const handleUpdateName = async () => {
    if (name.trim() === "" || name === camera.name) return;
    
    setIsUpdating(true);
    try {
      // This would be implemented in a full application
      await apiPut('/api/cameras/update', { camera_uuid: cameraUUID, name });
      
      // For now, just show updating state
      await new Promise(resolve => setTimeout(resolve, 500));
      
      setOpen(false);
      await refetchCameras();
    } catch (error) {
      console.error("Error updating camera:", error);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDeleteCamera = async () => {
    setIsDeleting(true);
    try {
      await apiDelete('/api/cameras/delete', { camera_uuid: cameraUUID });
      await refetchCameras();
      setOpen(false);
      // Navigate back to cameras list
      navigate('/cameras');
    } catch (error) {
      console.error("Error deleting camera:", error);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Dialog open={open}  onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Camera Settings</DialogTitle>
          <DialogDescription>
            Configure settings for your camera. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        
        <Tabs defaultValue="general" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="danger">Danger Zone</TabsTrigger>
          </TabsList>
          
          <TabsContent value="general" className="mt-4 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Camera Name</Label>
              <Input 
                id="name" 
                value={name} 
                onChange={(e) => setName(e.target.value)} 
                placeholder="Enter camera name" 
              />
            </div>
            
            <div className="space-y-2">
              <Label>Camera UUID</Label>
              <div className="p-2 bg-muted text-muted-foreground rounded text-sm break-all">
                {camera.cameraUUID}
              </div>
            </div>
            
            <div className="space-y-2">
              <Label>Status</Label>
              <div className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${camera.isOnline ? 'bg-green-500' : 'bg-red-500'}`}></div>
                <span>{camera.isOnline ? 'Online' : 'Offline'}</span>
              </div>
            </div>
            
            {camera.lastSeen && (
              <div className="space-y-2">
                <Label>Last Seen</Label>
                <div className="text-sm text-muted-foreground">
                  {new Date(camera.lastSeen).toLocaleString()}
                </div>
              </div>
            )}
          </TabsContent>
          
          <TabsContent value="danger" className="mt-4 space-y-4">
            <div className="space-y-2">
              <Label className="text-red-500 font-bold">Danger Zone</Label>
              <p className="text-sm text-muted-foreground">
                Actions here can't be undone. Be careful.
              </p>
              
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive" className="w-full mt-2">
                    <Trash2 className="h-4 w-4 mr-2" /> Delete Camera
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This will permanently delete the camera "{camera.name}" from your account.
                      This action cannot be undone.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction 
                      onClick={handleDeleteCamera}
                      disabled={isDeleting}
                    >
                      {isDeleting ? 'Deleting...' : 'Delete Camera'}
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          </TabsContent>
        </Tabs>
        
        <DialogFooter>
          <Button onClick={handleUpdateName} disabled={isUpdating}>
            {isUpdating ? 'Saving...' : 'Save Changes'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}