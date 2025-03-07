import { Outlet, useLocation } from "react-router";
import { Toaster } from "@/components/ui/sonner"
import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/useAuth"
import { WebRTCProvider } from "@/context/WebRTCContext";
import { AppProvider } from "@/context/AppContext";
import { useMemo } from "react";
import * as React from "react";

export default function RootLayout() {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation();

  // Generate breadcrumb items based on the current path
  const breadcrumbItems = useMemo(() => {
    // Remove leading slash and split the path
    const pathSegments = location.pathname.split('/').filter(Boolean);
    
    if (pathSegments.length === 0) {
      return [{ label: "Dashboard", path: "/dash", isCurrentPage: true }];
    }

    // Create breadcrumb items from path segments
    return pathSegments.map((segment, index) => {
      // Create the complete path up to this segment
      const path = `/${pathSegments.slice(0, index + 1).join('/')}`;
      // Format the segment label (capitalize first letter, replace hyphens with spaces)
      const label = segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' ');
      // Check if this is the current (last) segment
      const isCurrentPage = index === pathSegments.length - 1;
      
      return { label, path, isCurrentPage };
    });
  }, [location.pathname]);

  if (isLoading) {
    return <div>Loading...</div>
  }

  if (!isAuthenticated) {
    return null // The useAuth hook will handle redirection
  }
  
  return (
    <AppProvider>
      <WebRTCProvider>
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset>
            <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
              <div className="flex items-center gap-2 px-4 w-full overflow-hidden">
                <SidebarTrigger className="-ml-1 flex-shrink-0" />
                <Separator orientation="vertical" className="mr-2 h-4 flex-shrink-0" />
                <Breadcrumb className="min-w-0 w-full">
                  <BreadcrumbList className="flex-nowrap whitespace-nowrap overflow-hidden">
                    {breadcrumbItems.map((item, index) => (
                      <React.Fragment key={item.path}>
                        {index > 0 && <BreadcrumbSeparator className="flex-shrink-0" />}
                        <BreadcrumbItem className={index === breadcrumbItems.length - 1 ? "min-w-0 flex-shrink" : "flex-shrink-0"}>
                          {item.isCurrentPage ? (
                            <BreadcrumbPage className="truncate max-w-[200px]">{item.label}</BreadcrumbPage>
                          ) : (
                            <BreadcrumbLink href={item.path} className="truncate max-w-[150px]">{item.label}</BreadcrumbLink>
                          )}
                        </BreadcrumbItem>
                      </React.Fragment>
                    ))}
                  </BreadcrumbList>
                </Breadcrumb>
              </div>
            </header>
            <Outlet />
          </SidebarInset>
        </SidebarProvider>
      </WebRTCProvider>
    </AppProvider>
  );
}