import * as React from "react"
import {
  AudioWaveform,
  BookOpen,
  Bot,
  Command,
  Frame,
  GalleryVerticalEnd,
  Map,
  PieChart,
  Settings2,
  SquareTerminal,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavGroups } from "@/components/nav-projects"
import { NavUser } from "@/components/nav-user"
import { LocationSwitcher } from "@/components/team-switcher"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"

// This is sample data.
const data = {
  user: {
    name: "shadcn",
    email: "m@example.com",
    avatar: "/avatars/shadcn.jpg",
  },
  locations: [
    {
      name: "Home",
      logo: GalleryVerticalEnd,
      plan: "Enterprise",
    },
    {
      name: "Shop",
      logo: AudioWaveform,
      plan: "Startup",
    },
    {
      name: "Mom's House",
      logo: Command,
      plan: "Free",
    },
  ],
  navMain: [
    {
      title: "Cameras",
      url: "#",
      icon: SquareTerminal,
      isActive: true,
      items: [
        {
          title: "Add Camera",
          url: "/dash/cameras/add"
        },
        {
          title: "All Cameras",
          url: "/dash",
        },
        {
          title: "Doorbell",
          url: "#",
        },
        {
          title: "Patio",
          url: "#",
        },
      ],
    },

   
    {
      title: "Settings",
      url: "#",
      icon: Settings2,
      items: [
        {
          title: "Access & Users",
          url: "#",
        },
        {
          title: "Notifications",
          url: "#",
        },
        {
          title: "Global Recording",
          url: "#",
        },
      ],
    },
  ],
  groups: [
    {
      name: "Outdoors",
      url: "#",
      icon: Frame,
    },
    {
      name: "Indoors",
      url: "#",
      icon: PieChart,
    },
    {
      name: "Gate",
      url: "#",
      icon: Map,
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <LocationSwitcher locations={data.locations} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavGroups groups={data.groups} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
