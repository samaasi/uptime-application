"use client"

import * as React from "react";
import {
    SidebarMenu,
    SidebarGroup,
    SidebarMenuItem,
    SidebarMenuButton,
    SidebarGroupContent,
} from "@/components/ui/sidebar";
import { type LucideIcon } from "lucide-react";

export type NavSecondaryItem = {
  id: string
  url: string
  title: string
  icon: LucideIcon
}

type NavSecondaryProps = {
  items: NavSecondaryItem[]
  currentPath: string
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>

export function NavSecondary({
  items,
  currentPath,
  ...props
}: NavSecondaryProps) {

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => {
            const isActive = item.url === currentPath

            return (
              <SidebarMenuItem key={item.id}>
                <SidebarMenuButton asChild className={isActive ? "bg-accent text-accent-foreground hover:bg-accent/90" : ""}>
                  <a
                    href={item.url}
                    aria-current={isActive ? "page" : undefined}
                    aria-label={item.title}
                  >
                    <item.icon className="h-4 w-4 shrink-0" />
                    <span className="truncate">{item.title}</span>
                  </a>
                </SidebarMenuButton>
              </SidebarMenuItem>
            )
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}