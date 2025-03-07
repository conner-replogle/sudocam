import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiGet } from '@/lib/api';
import { User } from '@/types/user';
import { Camera } from '@/types/camera';




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
  const [user, setUser] = useState<User | null>(null);
  const [cameras, setCameras] = useState<Camera[]>([]);
  const [loading, setLoading] = useState(true);

  // Fetch user data
  useEffect(() => {
    async function fetchUserData() {
      try {
        const userData = await apiGet('/api/users/me');
        setUser(userData);
      } catch (error) {
        console.error('Error fetching user data:', error);
        setUser(null);
      }
    }

    fetchUserData();
  }, []);

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

  // Fetch cameras when user changes
  useEffect(() => {
    if (user) {
      refetchCameras();
    } else {
      setCameras([]);
      setLoading(false);
    }
  }, [user]);
  if (!user) {
    return <div>Loading...</div>;
  }
  console.log(user)

  return (
    <AppContext.Provider value={{ user, cameras, loading, refetchCameras }}>
      {children}
    </AppContext.Provider>
  );
}

// Create hook for using the context
export const useAppContext = () => useContext(AppContext);
