import { apiGet } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router";
import { LogOut } from "lucide-react";

interface LogoutButtonProps {
  variant?: "default" | "outline" | "destructive" | "ghost" | "link";
}

export function LogoutButton({ variant = "default" }: LogoutButtonProps) {
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await apiGet('/api/logout');
      navigate('/auth');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  return (
    <Button variant={variant} onClick={handleLogout}>
      <LogOut className="w-4 h-4 mr-2" />
      Logout
    </Button>
  );
}
