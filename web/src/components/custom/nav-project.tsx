"use client"

import {
  Folder,
  Trash2,
  Forward,
  PlusCircle,
  MoreHorizontal,
  type LucideIcon,
} from "lucide-react"

import {
  DropdownMenu,
  DropdownMenuItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu"
import {
  useSidebar,
  SidebarMenu,
  SidebarGroup,
  SidebarMenuItem,
  SidebarGroupLabel,
  SidebarMenuAction,
  SidebarMenuButton,
} from "@/components/ui/sidebar"

export type ProjectItem = {
  id: string | number
  name: string
  url: string
  icon: LucideIcon
}

type NavProjectsProps = {
  projects: ProjectItem[]
  onViewProject?: (project: ProjectItem) => void
  onShareProject?: (project: ProjectItem) => void
  onDeleteProject?: (project: ProjectItem) => void
  onAddNewProject?: () => void
}

/**
 * Renders a group of project links in a sidebar, including a dropdown
 * menu for contextual actions on each project.
 * @param {NavProjectsProps} props - The component props.
 */
export function NavProjects({
  projects,
  onViewProject,
  onShareProject,
  onDeleteProject,
  onAddNewProject,
}: NavProjectsProps) {
  const { isMobile } = useSidebar()
  const hasProjects = projects && projects.length > 0

  return (
    <SidebarGroup className="group-data-[collapsible=icon]:hidden">
      <SidebarGroupLabel>Projects</SidebarGroupLabel>
      <SidebarMenu>
        {hasProjects ? (
          projects.map((item) => (
            <SidebarMenuItem key={item.id}>
              <SidebarMenuButton asChild>
                <a href={item.url} aria-label={`Go to ${item.name} project`}>
                  <item.icon className="h-4 w-4 shrink-0" />
                  <span className="truncate">{item.name}</span>
                </a>
              </SidebarMenuButton>

              {/* Contextual Actions Dropdown */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  {/* showOnHover prop is nice, but ensure it's accessible */}
                  <SidebarMenuAction
                    aria-label={`More actions for ${item.name}`}
                    title={`More actions for ${item.name}`}
                  >
                    <MoreHorizontal className="h-4 w-4" />
                    <span className="sr-only">More actions</span>
                  </SidebarMenuAction>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  className="w-48 rounded-lg"
                  side={isMobile ? "bottom" : "right"}
                  align={isMobile ? "end" : "start"}
                  sideOffset={4}
                >
                  <DropdownMenuItem onClick={() => onViewProject?.(item)}>
                    <Folder className="mr-2 h-4 w-4 text-muted-foreground" />
                    <span>View Project</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onShareProject?.(item)}>
                    <Forward className="mr-2 h-4 w-4 text-muted-foreground" />
                    <span>Share Project</span>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="text-red-600 focus:text-red-600 focus:bg-red-50" // Highlighting dangerous action
                    onClick={() => onDeleteProject?.(item)}
                  >
                    <Trash2 className="mr-2 h-4 w-4 text-red-600" />
                    <span>Delete Project</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </SidebarMenuItem>
          ))
        ) : (
          <SidebarMenuItem className="text-muted-foreground italic">
            No projects found.
          </SidebarMenuItem>
        )}

        <SidebarMenuItem>
          <SidebarMenuButton
            className="text-sidebar-foreground/70 hover:text-sidebar-foreground"
            onClick={onAddNewProject}
          >
            <PlusCircle className="h-4 w-4 shrink-0 text-sidebar-foreground/70" />
            <span>Add New Project</span>
          </SidebarMenuButton>
        </SidebarMenuItem>

        {/* The original 'More' button is redundant with 'Add New Project' unless it serves a distinct purpose,
            but for compatibility, let's keep it as an example of an alternative action. 
            I've opted to replace it with a more concrete action: 'Add New Project'. 
            If "More" is required for other features, it would need a distinct `onClick` handler.
        */}
      </SidebarMenu>
    </SidebarGroup>
  )
}