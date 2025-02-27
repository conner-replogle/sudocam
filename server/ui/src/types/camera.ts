export interface Camera {
    id: number;
    cameraUUID: string;
    userID: number;
    lastOnline: string | null;
    onlineStatus: boolean;
}
