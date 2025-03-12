import { useAppContext } from "@/context/AppContext";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { Trash2, Plus, X } from "lucide-react";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { MotionConfig, RecordingType, Schedule, UserConfig } from "@/types/binding";

interface CameraSettingsProps {
  id: string;
  children: React.ReactNode;
}

export function CameraSettings({ id, children }: CameraSettingsProps) {
  const { cameras, refetchCameras } = useAppContext();
  const navigate = useNavigate();
  const camera = cameras.find((c) => c.id === id);
  const [name, setName] = useState(camera?.name || "");
  const [isDeleting, setIsDeleting] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [open, setOpen] = useState(false);
  // Recording config state
  const [recordingConfig, setRecordingConfig] = useState<UserConfig>({
    recording_type: RecordingType.RECORDING_TYPE_OFF,
    motion_enabled: false,
    motion_config: {
      sensitivity: 50,
      pre_record_seconds: 5,
      post_record_seconds: 10
    },
    schedules: []
  });

  // Load camera config when dialog opens
  useEffect(() => {
    if (open && camera) {
      // If camera has config, use it, otherwise use default                                                                                                                                                                                                  
      if (camera.config) {
        setRecordingConfig(camera.config);
      }                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
    }
  }, [open, camera]);
  
  if (!camera) {
    return null;
  }

  const handleUpdateCamera = async () => {
    setIsUpdating(true);
    try {
      // Update both name and config
      await apiPut('/api/cameras/update', { 
        id, 
        name: name !== camera.name ? name : undefined,
        config: recordingConfig
      });
      
      await refetchCameras();
      setOpen(false);
    } catch (error) {
      console.error("Error updating camera:", error);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDeleteCamera = async () => {
    setIsDeleting(true);
    try {
      await apiDelete('/api/cameras/delete', { camera_uuid: id });
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

  const addSchedule = () => {
    const newSchedule: Schedule = {
      days_of_week: [0, 1, 2, 3, 4, 5, 6], // All days by default
      start_time: "08:00",
      end_time: "18:00"
    };
    setRecordingConfig({
      ...recordingConfig,
      schedules: [...(recordingConfig.schedules || []), newSchedule]
    });
  };

  const removeSchedule = (index: number) => {
    const updatedSchedules = [...(recordingConfig.schedules || [])];
    updatedSchedules.splice(index, 1);
    setRecordingConfig({
      ...recordingConfig,
      schedules: updatedSchedules
    });
  };

  const updateSchedule = (index: number, schedule: Partial<Schedule>) => {
    const updatedSchedules = [...(recordingConfig.schedules || [])];
    updatedSchedules[index] = {
      ...updatedSchedules[index],
      ...schedule
    };
    setRecordingConfig({
      ...recordingConfig,
      schedules: updatedSchedules
    });
  };

  const handleDayToggle = (index: number, day: number) => {
    const schedule = recordingConfig.schedules?.[index];
    if (!schedule) return;

    const days = [...(schedule.days_of_week || [])];
    const dayIndex = days.indexOf(day);

    if (dayIndex >= 0) {
      days.splice(dayIndex, 1);
    } else {
      days.push(day);
      days.sort((a, b) => a - b);
    }

    updateSchedule(index, { days_of_week: days });
  };

  const formatDayInitial = (day: number): string => {
    return ['S', 'M', 'T', 'W', 'T', 'F', 'S'][day];
  };
  console.log(recordingConfig);
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[550px]">
        <DialogHeader>
          <DialogTitle>Camera Settings</DialogTitle>
          <DialogDescription>
            Configure settings for your camera. Click save when you're done.
          </DialogDescription>
        </DialogHeader>
        
        <Tabs defaultValue="general" className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="recording">Recording</TabsTrigger>
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
                {camera.id}
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
          
          <TabsContent value="recording" className="mt-4 space-y-6">
            <div className="space-y-3">
              <h3 className="font-medium text-sm">Recording Mode</h3>
              <Select 
                value={recordingConfig.recording_type} 
                onValueChange={(value) => setRecordingConfig({
                  ...recordingConfig,
                  recording_type: value as RecordingType
                })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select recording mode" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={RecordingType.RECORDING_TYPE_OFF}>Off (No Recording)</SelectItem>
                  <SelectItem value={RecordingType.RECORDING_TYPE_CONTINUOUS}>Continuous (24/7)</SelectItem>
                  <SelectItem value={RecordingType.RECORDING_TYPE_CONTINUOUS_SCHEDULED}>Scheduled</SelectItem>
                  <SelectItem value={RecordingType.RECORDING_TYPE_MOTION}>Motion Detection</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            {recordingConfig.recording_type === RecordingType.RECORDING_TYPE_MOTION && (
              <div className="space-y-4 p-4 border rounded-md">
                <h3 className="font-medium">Motion Detection Settings</h3>
                
                <div className="space-y-2">
                  <div className="flex justify-between items-center">
                    <Label htmlFor="motion-sensitivity">Sensitivity</Label>
                    <span className="text-sm">{recordingConfig.motion_config?.sensitivity}%</span>
                  </div>
                  <Slider 
                    id="motion-sensitivity"
                    min={1} 
                    max={100} 
                    step={1}
                    value={[recordingConfig.motion_config?.sensitivity || 50]} 
                    onValueChange={(value) => setRecordingConfig({
                      ...recordingConfig,
                      motion_config: {
                        ...(recordingConfig.motion_config || {}),
                        sensitivity: value[0]
                      }
                    })} 
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="pre-record">Pre-record (seconds)</Label>
                    <Input 
                      id="pre-record"
                      type="number" 
                      min={0}
                      max={30}
                      value={recordingConfig.motion_config?.pre_record_seconds || 0} 
                      onChange={(e) => setRecordingConfig({
                        ...recordingConfig,
                        motion_config: {
                          ...(recordingConfig.motion_config || {}),
                          pre_record_seconds: parseInt(e.target.value)
                        }
                      })} 
                    />
                  </div>
                  
                  <div className="space-y-2">
                    <Label htmlFor="post-record">Post-record (seconds)</Label>
                    <Input 
                      id="post-record"
                      type="number" 
                      min={0}
                      max={60}
                      value={recordingConfig.motion_config?.post_record_seconds || 0} 
                      onChange={(e) => setRecordingConfig({
                        ...recordingConfig,
                        motion_config: {
                          ...(recordingConfig.motion_config || {}),
                          post_record_seconds: parseInt(e.target.value)
                        }
                      })} 
                    />
                  </div>
                </div>
              </div>
            )}
            
            {recordingConfig.recording_type === RecordingType.RECORDING_TYPE_CONTINUOUS_SCHEDULED && (
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <h3 className="font-medium">Recording Schedule</h3>
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={addSchedule}
                  >
                    <Plus className="h-4 w-4 mr-1" /> Add Schedule
                  </Button>
                </div>
                
                {recordingConfig.schedules?.length === 0 && (
                  <div className="text-center p-4 border rounded-md text-muted-foreground">
                    No schedules added. Add a schedule to record during specific times.
                  </div>
                )}
                
                {recordingConfig.schedules?.map((schedule, index) => (
                  <div key={index} className="p-4 border rounded-md space-y-3">
                    <div className="flex justify-between items-center">
                      <h4 className="font-medium text-sm">Schedule {index + 1}</h4>
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={() => removeSchedule(index)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                    
                    <div className="space-y-2">
                      <Label>Days</Label>
                      <div className="flex space-x-2">
                        {[0, 1, 2, 3, 4, 5, 6].map((day) => (
                          <div 
                            key={day}
                            onClick={() => handleDayToggle(index, day)}
                            className={`
                              w-8 h-8 flex items-center justify-center rounded-full cursor-pointer
                              ${schedule.days_of_week?.includes(day) 
                                ? 'bg-primary text-primary-foreground' 
                                : 'bg-muted text-muted-foreground'}
                            `}
                          >
                            {formatDayInitial(day)}
                          </div>
                        ))}
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor={`start-time-${index}`}>Start Time</Label>
                        <Input 
                          id={`start-time-${index}`}
                          type="time" 
                          value={schedule.start_time || "00:00"} 
                          onChange={(e) => updateSchedule(index, { start_time: e.target.value })} 
                        />
                      </div>
                      
                      <div className="space-y-2">
                        <Label htmlFor={`end-time-${index}`}>End Time</Label>
                        <Input 
                          id={`end-time-${index}`}
                          type="time" 
                          value={schedule.end_time || "00:00"} 
                          onChange={(e) => updateSchedule(index, { end_time: e.target.value })} 
                        />
                      </div>
                    </div>
                  </div>
                ))}
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
          <Button onClick={handleUpdateCamera} disabled={isUpdating}>
            {isUpdating ? 'Saving...' : 'Save Changes'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}