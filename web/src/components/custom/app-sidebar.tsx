"use client";

import * as React from "react";
import Image from "next/image";
import {
    Sidebar,
    SidebarRail,
    SidebarMenu,
    SidebarGroup,
    SidebarHeader,
    SidebarFooter,
    SidebarContent,
    SidebarMenuItem,
    SidebarMenuButton,
    SidebarGroupContent,
} from "@/components/ui/sidebar";
import { NavMain } from "@/components/custom/nav-main";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { NavProjects, type ProjectItem } from "@/components/custom/nav-project";
import { OrganizationSwitcher } from "@/components/custom/organization-switcher";
import { Frame, ActivitySquareIcon, Settings, GalleryVerticalEnd, SquareTerminal, Map, Bot, Settings2, BookOpen, PieChart } from "lucide-react";
import Link from "next/link";

const data = {
    organizations: [
      {
        name: "Team A",
        logo: GalleryVerticalEnd,
        plan: "Free",
      },
      {
        name: "Team B",
        logo: GalleryVerticalEnd,
        plan: "Pro",
      },
    ],
    navMain: [
    {
      title: "Dashboard",
      url: "#",
      icon: ActivitySquareIcon,
      isActive: true
    },
    {
      title: "Playground",
      url: "#",
      icon: SquareTerminal,
      items: [
        {
          title: "History",
          url: "#",
        },
        {
          title: "Starred",
          url: "#",
        },
        {
          title: "Settings",
          url: "#",
        },
      ],
    },
    {
      title: "Models",
      url: "#",
      icon: Bot,
      items: [
        {
          title: "Genesis",
          url: "#",
        },
        {
          title: "Explorer",
          url: "#",
        },
        {
          title: "Quantum",
          url: "#",
        },
      ],
    },
    {
      title: "Documentation",
      url: "#",
      icon: BookOpen,
      items: [
        {
          title: "Introduction",
          url: "#",
        },
        {
          title: "Get Started",
          url: "#",
        },
        {
          title: "Tutorials",
          url: "#",
        },
        {
          title: "Changelog",
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
          title: "General",
          url: "#",
        },
        {
          title: "Team",
          url: "#",
        },
        {
          title: "Billing",
          url: "#",
        },
        {
          title: "Limits",
          url: "#",
        },
      ],
    },
  ],
}

const projects: ProjectItem[] = [
    {
      id: 'proj-design-eng',
      name: "Design Engineering",
      url: "#",
      icon: Frame,
    },
    {
      id: 'proj-sales-mktg',
      name: "Sales & Marketing",
      url: "#",
      icon: PieChart,
    },
    {
      id: 'proj-travel-plan',
      name: "Travel",
      url: "#",
      icon: Map,
    },
  ];

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {

  const handleViewProject = (project: ProjectItem) => {
    alert(`Viewing project: ${project.name} (ID: ${project.id})`)
    // Example: router.push(`/projects/${project.id}`);
  }

  const handleShareProject = (project: ProjectItem) => {
    alert(`Sharing project: ${project.name}`)
    // Example: copyToClipboard(`${window.location.origin}/projects/${project.id}`);
  }

  const handleDeleteProject = (project: ProjectItem) => {
    if (confirm(`Are you sure you want to delete "${project.name}"?`)) {
      console.log("Deleting project:", project)
    }
  }

  const handleAddNewProject = () => {
    alert("Opening the 'Add New Project' modal...")
    // Example: setModalOpen(true);
  }
  return (
    <Sidebar variant="floating" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <OrganizationSwitcher organizations={data.organizations} />
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain label="Platform" items={data.navMain} />
        <NavProjects
          projects={projects}
          onViewProject={handleViewProject}
          onShareProject={handleShareProject}
          onDeleteProject={handleDeleteProject}
          onAddNewProject={handleAddNewProject}
        />
      </SidebarContent>
      <SidebarRail />
      <SidebarFooter>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild className={"hover:text-accent-foreground hover:bg-gray-200"}>
                  <Link
                    href={"#account"}
                    aria-current={"page"}
                    aria-label={'account'}
                  >
                    <Avatar className="rounded-sm h-6 w-6">
                      <AvatarImage src="https://github.com/maxleiter.png" alt="@maxleiter" />
                      <AvatarFallback>CN</AvatarFallback>
                    </Avatar>
                    <span className="truncate">Account</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton asChild className={"hover:text-accent-foreground hover:bg-gray-200"}>
                  <Link
                    href={"#settings"}
                    aria-current={"page"}
                    aria-label={'settings'}
                  >
                    <Settings className="h-6 w-6" />
                    <span className="truncate ml-1.5">Settings</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarFooter>
    </Sidebar>
  )
}
