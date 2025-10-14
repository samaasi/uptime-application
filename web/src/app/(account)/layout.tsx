import * as React from "react";
import { Separator } from "@/components/ui/separator";
import { AppSidebar } from "@/components/custom/app-sidebar";
import { AppBreadcrumb } from "@/components/custom/app-breadcrumb";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"

export default function AccountLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider style={{ "--sidebar-width": "19rem" } as React.CSSProperties }>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <AppBreadcrumb />
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          { children }
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}