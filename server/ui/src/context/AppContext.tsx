import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiGet } from '@/lib/api';
import { User } from '@/types/user';
import { Camera } from '@/types/camera';
import { useConnection } from '@/hooks/use-connection';
import { Message } from '@/types/binding';
import { useUserContext } from './UserContext';

interface AppContextType {
  user: User | null;
  cameras: Camera[];
  loading: boolean;
  refetchCameras: () => Promise<void>;
}

// Create context with default values
const AppContext = createContext<AppContextType>({
  user: null,
  cameras: [],
  loading: true,
  refetchCameras: async () => {},
});

// Create provider component
export function AppProvider({ children }: { children: ReactNode }) {
  const { user } = useUserContext();


  const [cameras, setCameras] = useState<Camera[]>([]);
  const [loading, setLoading] = useState(true);

  // Fetch user data

  const { } = useConnection({
    onMessage: async (message:Message) => {
      if (message.trigger_refresh){
        console.log("Received trigger_refresh message");
        refetchCameras();
      }
    } 
  });
  

  // Function to fetch cameras
  const refetchCameras = async () => {
    if (!user) {
      setCameras([]);
      return;
    }

    try {
      setLoading(true);
      const data = await apiGet('/api/users/cameras');
      setCameras(data || []);
    } catch (error) {
      console.error('Error fetching cameras:', error);
      setCameras([]);
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => {
    refetchCameras();
  }, [user]);


  return (
    <AppContext.Provider value={{ user, cameras, loading, refetchCameras }}>
      {children}
    </AppContext.Provider>
  );
}

// Create hook for using the context
export const useAppContext = () => useContext(AppContext);
