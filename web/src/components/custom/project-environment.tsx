'use client';

import * as React from "react";
import Image from "next/image";
import {
  Select,
  SelectItem,
  SelectValue,
  SelectContent,
  SelectTrigger,
} from '@/components/ui/select';
import { 
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { ChevronsUpDown } from "lucide-react";
import { StatusIndicator } from "./status-indicator";

interface ProjectProps {
    value: string;
    label: string;
    image: string;
}

const projects: ProjectProps[] = [
    { value: '1', label: 'Main project', image: 'https://placehold.co/20x20' },
    { value: '2', label: 'Origin project', image: 'https://placehold.co/20x20' },
    { value: '3', label: 'Legacy project', image: 'https://placehold.co/20x20' },
];

const environments = [
    {value: '1', label: 'Development', status: 'development'},
    {value: '2', label: 'Production', status: 'production'},
    {value: '3', label: 'Staging', status: 'staging'},
]

export const ProjectEnvironment = () => {
    const defaultProject: string = "1";
    const defaultEnvironment: string = "2";
    const [project, setProject] = React.useState('');
    const [environment, setEnvironment] = React.useState('');

    const handleProjectChange = (value: string) => {
        setProject(value);
    }

    const handleEnvironmentChange = (value: string) => {
        setEnvironment(value);
    }
    
  return (
    <div className="flex items-center justify-between bg-gray-100 shadow-md rounded-md">
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                  <Select 
                    defaultValue={defaultProject}
                    onValueChange={handleProjectChange}
                  >
                    <SelectTrigger className="focus-visible:bg-accent text-foreground h-8 px-1.5 focus-visible:ring-0 border-none shadow-none bg-transparent hover:bg-accent [&>svg:not(.lucide-chevrons-up-down)]:hidden">
                      <SelectValue placeholder="Select project">
                        {(project || defaultProject) && (
                          <div className="flex items-center gap-2">
                            <Image 
                                width={16}
                                height={16}
                                className="w-4 h-4 rounded-sm object-cover"
                                src={projects.find(p => p.value === (project || defaultProject))?.image || 'https://placehold.co/20x20'}
                                alt={"image"}
                            />
                            <span>{projects.find(p => p.value === (project || defaultProject))?.label}</span>
                          </div>
                        )}
                      </SelectValue>
                      <ChevronsUpDown
                        size={14}
                        className="text-muted-foreground/80 ml-1 lucide-chevrons-up-down"
                      />
                    </SelectTrigger>
                    <SelectContent className="[&_*[role=option]]:ps-2 [&_*[role=option]]:pe-8 [&_*[role=option]>span]:start-auto [&_*[role=option]>span]:end-2">
                      {projects.map((project) => (
                        <SelectItem key={project.value} value={project.value}>
                          <div className="flex items-center gap-2">
                            <Image
                              src={project.image} 
                              alt={project.label} 
                              className="w-4 h-4 rounded-sm object-cover"
                              width={16}
                              height={16}
                            />
                            <span>{project.label}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </BreadcrumbItem>
                <BreadcrumbSeparator> / </BreadcrumbSeparator>
                <BreadcrumbItem>
                  <Select 
                    defaultValue={defaultEnvironment}
                    onValueChange={handleEnvironmentChange}
                  >
                    <SelectTrigger className="focus-visible:bg-accent text-foreground h-8 px-1.5 focus-visible:ring-0 border-none shadow-none bg-transparent hover:bg-accent [&>svg:not(.lucide-chevrons-up-down)]:hidden">
                      <SelectValue placeholder="Select environment">
                        {(environment || defaultEnvironment) && (
                          <div className="flex items-center gap-2">
                            <StatusIndicator 
                              status={environments.find(e => e.value === (environment || defaultEnvironment))?.status || 'offline'} 
                            />
                            <span>{environments.find(e => e.value === (environment || defaultEnvironment))?.label}</span>
                          </div>
                        )}
                      </SelectValue>
                      <ChevronsUpDown
                        size={14}
                        className="text-muted-foreground/80 ml-1 lucide-chevrons-up-down"
                      />
                    </SelectTrigger>
                    <SelectContent className="[&_*[role=option]]:ps-2 [&_*[role=option]]:pe-8 [&_*[role=option]>span]:start-auto [&_*[role=option]>span]:end-2">
                      {environments.map((environment) => (
                        <SelectItem key={environment.value} value={environment.value}>
                          <div className="flex items-center gap-2">
                            <StatusIndicator status={environment.status} />
                            <span>{environment.label}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
    </div>
  );
}