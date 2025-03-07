import * as React from "react"
import { useNavigate } from "react-router"
import {
  Camera as CameraIcon,
  PlusCircle,
  Settings2,
  LayoutGrid,
  PlusCircleIcon,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavGroups } from "@/components/nav-groups"
import { NavUser } from "@/components/nav-user"
import { LocationSwitcher } from "@/components/team-switcher"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"
import { useAppContext } from "@/context/AppContext"
import { Badge } from "./ui/badge"
import { ModeToggle } from "./mode-toggle"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { cameras, user } = useAppContext();

  // Generate navigation items dynamically based on camera data
  const navItems = React.useMemo(() => {
    const items = [
     
      {
        title: "Cameras",
        url: "/",
        icon: CameraIcon,
        isActive: true,
        items: [
          {
            title: "Add Camera",
            url: "/cameras/add",
            tag: PlusCircleIcon,
          },
          {
            title: "All Cameras",
            url: "/cameras",
          },
         
          ...cameras.map(camera => ({
            title: camera.name,
            url: `/cameras/${camera.cameraUUID}`,
            tag: () => (camera.isOnline ? null:  <Badge variant={"destructive"}>Offline</Badge>),
          
          }))
        ],
      },
      {
        title: "Groups",
        url: "/groups",
        icon: LayoutGrid,
        isActive: true,
        items: [
          {
            title: "All",
            url: "/groups",
          },
         
        ],
      },
      {
        title: "Settings",
        url: "/settings",
        icon: Settings2,
        items: [
          {
            title: "Account",
            url: "/settings/account",
          },
          {
            title: "Cameras",
            url: "/settings/cameras",
          },
          {
            title: "Notifications",
            url: "/settings/notifications",
          },
        ],
      },
    ];
    
    return items;
  }, [cameras]);

  const userData = user ? {
    name: user.name || user.email.split('@')[0],
    email: user.email,
    avatar: user.avatar || "/avatars/default.png",
  } : {
    name: "Guest",
    email: "guest@example.com",
    avatar: "/avatars/default.png",
  };

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        {/* <LocationSwitcher locations={locations.map(loc => ({
          name: loc.name,
          logo: LayoutGrid,
          address: loc.cameraCount.toString(),
        }))} /> */}
        <ModeToggle />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navItems} />
       
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={userData} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
